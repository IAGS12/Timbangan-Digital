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

	// 2. Data Dummy Sapi
	cows := []struct {
		CowCode   string `db:"cow_code"`
		Name      string `db:"name"`
		Breed     string `db:"breed"`
		Gender    string `db:"gender"`
		BirthDate string `db:"birth_date"`
		Owner     string `db:"owner"`
		Status    string `db:"status"`
	}{
		{"RFID-0001-TEST", "Sapi Brahman A", "Brahman", "jantan", "2024-01-01", "Peternak A", "active"},
		{"RFID-0002-EVAL", "Sapi Brahman B", "Brahman", "jantan", "2024-02-15", "Peternak B", "active"},
		{"RFID-0003-FAIL", "Sapi Brahman C", "Brahman", "jantan", "2024-03-10", "Peternak C", "active"},
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
				log.Printf("Berhasil insert sapi: %s\n", cow.Name)
			}
		} else {
			log.Printf("Sapi %s sudah ada di database.\n", cow.Name)
		}
	}

	// 3. Data Dummy Riwayat Penimbangan
	log.Println("Menambahkan Data Riwayat Penimbangan...")
	
	rows, err := db.Queryx("SELECT id, cow_code, birth_date FROM cows")
	if err != nil {
		log.Fatalf("Gagal mengambil data sapi untuk seeder timbangan: %v", err)
	}

	type CowInfo struct {
		ID        int64
		Code      string
		BirthDate string
	}
	var cowInfos []CowInfo

	for rows.Next() {
		var info CowInfo
		rows.Scan(&info.ID, &info.Code, &info.BirthDate)
		cowInfos = append(cowInfos, info)
	}
	rows.Close() // Tutup rows agar SQLite tidak terkunci

	for _, info := range cowInfos {
		id := info.ID
		code := info.Code

		var weightCount int
		db.Get(&weightCount, "SELECT COUNT(*) FROM weight_records WHERE cow_id = ?", id)
		
		if weightCount == 0 {
			var weighings []struct {
				date   time.Time
				weight float64
				adg    float64
			}

			if code == "RFID-0001-TEST" {
				// Layak dipertahankan (Kenaikan bobot harian tinggi)
				weighings = []struct {
					date   time.Time
					weight float64
					adg    float64
				}{
					{time.Date(2026, time.May, 15, 8, 0, 0, 0, time.Local), 350.0, 0.8},
					{time.Date(2026, time.June, 15, 8, 0, 0, 0, time.Local), 375.0, 0.83},
					{time.Date(2026, time.July, 15, 8, 0, 0, 0, time.Local), 405.0, 1.0},
				}
			} else if code == "RFID-0002-EVAL" {
				// Perlu evaluasi (Kenaikan bobot harian lambat/sedang)
				weighings = []struct {
					date   time.Time
					weight float64
					adg    float64
				}{
					{time.Date(2026, time.May, 15, 8, 0, 0, 0, time.Local), 300.0, 0.1},
					{time.Date(2026, time.June, 15, 8, 0, 0, 0, time.Local), 304.0, 0.13},
					{time.Date(2026, time.July, 15, 8, 0, 0, 0, time.Local), 308.0, 0.13},
				}
			} else if code == "RFID-0003-FAIL" {
				// Tidak layak dipertahankan (Bobot badan menurun)
				weighings = []struct {
					date   time.Time
					weight float64
					adg    float64
				}{
					{time.Date(2026, time.May, 15, 8, 0, 0, 0, time.Local), 420.0, -0.3},
					{time.Date(2026, time.June, 15, 8, 0, 0, 0, time.Local), 410.0, -0.33},
					{time.Date(2026, time.July, 15, 8, 0, 0, 0, time.Local), 395.0, -0.5},
				}
			}

			if len(weighings) > 0 {
				for _, w := range weighings {
					db.MustExec(`
						INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id, operator_id) 
						VALUES (?, ?, ?, ?, ?, ?)`,
						id, w.weight, w.adg, w.date.Format(time.RFC3339), "SEEDER-IOT", nil)
				}
			}
		}
	}

	// 4. Data Dummy User Admin
	adminPass, _ := bcrypt.GenerateFromPassword([]byte("Agustin123"), bcrypt.DefaultCost)
	var userExists int
	db.Get(&userExists, "SELECT COUNT(*) FROM users WHERE username = 'indra'")
	if userExists == 0 {
		db.MustExec(`INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)`,
			"indra", "indra@livestock.com", string(adminPass), "admin")
		log.Println("Berhasil membuat user 'indra' (password: Agustin123)")
	}

	// 5. Data Dummy Perangkat Timbangan ESP32
	var devCount int
	db.Get(&devCount, "SELECT COUNT(*) FROM devices WHERE device_code = 'SCALE-ESP32-01'")
	if devCount == 0 {
		db.MustExec(`INSERT INTO devices (device_code, device_name, user_id, status, pairing_status) VALUES (?, ?, ?, ?, ?)`,
			"SCALE-ESP32-01", "Timbangan Utama Kandang A", 1, "active", "approved")
		log.Println("Berhasil membuat device dummy 'SCALE-ESP32-01' (pairing_status: approved)")
	}

	log.Println("Seeder Selesai!")
}
