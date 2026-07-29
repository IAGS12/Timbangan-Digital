//go:build ignore
// +build ignore

package main

import (
	"log"
	"time"

	"smart-livestock-backend/config"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.Println("Memulai Seeder Database Smart Livestock...")

	// 1. Muat konfigurasi dan DB
	config.LoadConfig()
	config.InitializeDatabase()
	defer config.CloseDatabase()

	db := config.DB

	// ============================================================
	// 2. Data 3 Sapi Demo (Bali, Madura, PO)
	//    Kode RFID menentukan kategori DSS:
	//    - RFID-BALI-01   → Layak Dipertahankan   (tren naik pesat)
	//    - RFID-MADURA-02 → Perlu Evaluasi         (tren naik lambat)
	//    - RFID-PO-03     → Tidak Layak            (tren turun)
	// ============================================================
	cows := []struct {
		CowCode   string `db:"cow_code"`
		Name      string `db:"name"`
		Breed     string `db:"breed"`
		Gender    string `db:"gender"`
		BirthDate string `db:"birth_date"`
		Owner     string `db:"owner"`
		Status    string `db:"status"`
	}{
		{
			CowCode:   "RFID-BALI-01",
			Name:      "Sapi Bali",
			Breed:     "Bali",
			Gender:    "jantan",
			BirthDate: "2024-01-10",
			Owner:     "Peternak Demo",
			Status:    "active",
		},
		{
			CowCode:   "RFID-MADURA-02",
			Name:      "Sapi Madura",
			Breed:     "Madura",
			Gender:    "jantan",
			BirthDate: "2024-02-20",
			Owner:     "Peternak Demo",
			Status:    "active",
		},
		{
			CowCode:   "RFID-PO-03",
			Name:      "Sapi PO",
			Breed:     "Peranakan Ongole",
			Gender:    "jantan",
			BirthDate: "2024-03-05",
			Owner:     "Peternak Demo",
			Status:    "active",
		},
	}

	// Insert Cows
	for _, cow := range cows {
		var exists int
		err := db.Get(&exists, "SELECT COUNT(*) FROM cows WHERE cow_code = ?", cow.CowCode)
		if err == nil && exists == 0 {
			query := `INSERT INTO cows (cow_code, name, breed, gender, birth_date, owner, status) 
			          VALUES (:cow_code, :name, :breed, :gender, :birth_date, :owner, :status)`
			_, err = db.NamedExec(query, cow)
			if err != nil {
				log.Printf("Gagal insert sapi %s: %v\n", cow.Name, err)
			} else {
				log.Printf("✓ Berhasil insert sapi: %s (%s)\n", cow.Name, cow.CowCode)
			}
		} else {
			log.Printf("→ Sapi %s sudah ada di database, dilewati.\n", cow.Name)
		}
	}

	// ============================================================
	// 3. Data Riwayat Penimbangan (3 data per sapi, interval ~30 hari)
	//    Dirancang agar engine regresi linear menghasilkan DSS label yg sesuai
	// ============================================================
	log.Println("Menambahkan Data Riwayat Penimbangan...")

	// Map kode sapi → data timbangan
	type WeighingData struct {
		date   time.Time
		weight float64
		adg    float64
	}

	weighingsByCowCode := map[string][]WeighingData{
		// --- LAYAK DIPERTAHANKAN ---
		// ADG tinggi (+0.9 kg/hari), bobot naik pesat: 340 → 367 → 400 kg
		"RFID-BALI-01": {
			{time.Date(2026, time.May, 15, 8, 0, 0, 0, time.Local), 340.0, 0.0},
			{time.Date(2026, time.June, 15, 8, 0, 0, 0, time.Local), 367.0, 0.9},
			{time.Date(2026, time.July, 15, 8, 0, 0, 0, time.Local), 400.0, 1.1},
		},
		// --- PERLU EVALUASI ---
		// ADG rendah (+0.13 kg/hari), bobot naik lambat: 290 → 294 → 298 kg
		"RFID-MADURA-02": {
			{time.Date(2026, time.May, 15, 8, 0, 0, 0, time.Local), 290.0, 0.0},
			{time.Date(2026, time.June, 15, 8, 0, 0, 0, time.Local), 294.0, 0.13},
			{time.Date(2026, time.July, 15, 8, 0, 0, 0, time.Local), 298.0, 0.13},
		},
		// --- TIDAK LAYAK DIPERTAHANKAN ---
		// ADG negatif (-0.4 kg/hari), bobot turun: 430 → 418 → 403 kg
		"RFID-PO-03": {
			{time.Date(2026, time.May, 15, 8, 0, 0, 0, time.Local), 430.0, 0.0},
			{time.Date(2026, time.June, 15, 8, 0, 0, 0, time.Local), 418.0, -0.4},
			{time.Date(2026, time.July, 15, 8, 0, 0, 0, time.Local), 403.0, -0.5},
		},
	}

	// Ambil ID sapi yang baru di-insert
	rows, err := db.Queryx("SELECT id, cow_code FROM cows WHERE cow_code IN ('RFID-BALI-01', 'RFID-MADURA-02', 'RFID-PO-03')")
	if err != nil {
		log.Fatalf("Gagal mengambil data sapi: %v", err)
	}

	type CowRow struct {
		ID      int64  `db:"id"`
		CowCode string `db:"cow_code"`
	}
	var cowRows []CowRow
	for rows.Next() {
		var r CowRow
		rows.StructScan(&r)
		cowRows = append(cowRows, r)
	}
	rows.Close()

	for _, cr := range cowRows {
		var weightCount int
		db.Get(&weightCount, "SELECT COUNT(*) FROM weight_records WHERE cow_id = ?", cr.ID)

		if weightCount == 0 {
			weighings, ok := weighingsByCowCode[cr.CowCode]
			if !ok {
				continue
			}
			for i, w := range weighings {
				_, err := db.Exec(
					`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id, operator_id) 
					 VALUES (?, ?, ?, ?, ?, ?)`,
					cr.ID, w.weight, w.adg, w.date.Format(time.RFC3339), "SEEDER-IOT", nil,
				)
				if err != nil {
					log.Printf("  Gagal insert timbangan ke-%d untuk %s: %v\n", i+1, cr.CowCode, err)
				} else {
					log.Printf("  ✓ Timbangan ke-%d %s: %.1f kg (ADG: %.2f)\n", i+1, cr.CowCode, w.weight, w.adg)
				}
			}
		} else {
			log.Printf("→ Data timbangan untuk %s sudah ada (%d record), dilewati.\n", cr.CowCode, weightCount)
		}
	}

	// ============================================================
	// 4. Data Dummy User Admin
	// ============================================================
	adminPass, _ := bcrypt.GenerateFromPassword([]byte("Agustin123"), bcrypt.DefaultCost)
	var userExists int
	db.Get(&userExists, "SELECT COUNT(*) FROM users WHERE username = 'indra'")
	if userExists == 0 {
		db.MustExec(`INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)`,
			"indra", "indra@livestock.com", string(adminPass), "admin")
		log.Println("✓ Berhasil membuat user 'indra' (password: Agustin123)")
	} else {
		log.Println("→ User 'indra' sudah ada, dilewati.")
	}

	// ============================================================
	// 5. Data Dummy Perangkat Timbangan ESP32
	// ============================================================
	var devCount int
	db.Get(&devCount, "SELECT COUNT(*) FROM devices WHERE device_code = 'SCALE-ESP32-01'")
	if devCount == 0 {
		db.MustExec(
			`INSERT INTO devices (device_code, device_name, user_id, status, pairing_status) VALUES (?, ?, ?, ?, ?)`,
			"SCALE-ESP32-01", "Timbangan Utama Kandang A", 1, "active", "approved",
		)
		log.Println("✓ Berhasil membuat device 'SCALE-ESP32-01'")
	} else {
		log.Println("→ Device 'SCALE-ESP32-01' sudah ada, dilewati.")
	}

	log.Println("========================================")
	log.Println("✅ Seeder Selesai!")
	log.Println("   Sapi Bali      → Layak Dipertahankan")
	log.Println("   Sapi Madura    → Perlu Evaluasi")
	log.Println("   Sapi PO        → Tidak Layak Dipertahankan")
	log.Println("========================================")
}
