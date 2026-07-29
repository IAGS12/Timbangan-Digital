package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"smart-livestock-backend/config"
	"smart-livestock-backend/utils"
)

// ResetDatabase POST /api/admin/reset-db — Hapus SEMUA data dan seed ulang
// Endpoint ini SANGAT BERBAHAYA. Gunakan hanya untuk development / debugging.
func ResetDatabase(c *fiber.Ctx) error {
	db := config.DB

	log.Println("⚠️ RESET DATABASE dimulai via API endpoint...")

	// 1. Hapus semua data (urutan penting karena foreign key)
	_, _ = db.Exec("DELETE FROM predictions")
	_, _ = db.Exec("DELETE FROM weight_records")
	_, _ = db.Exec("DELETE FROM devices")
	_, _ = db.Exec("DELETE FROM cows")
	_, _ = db.Exec("DELETE FROM users")

	// 2. Reset auto increment
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'predictions'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'weight_records'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'devices'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'cows'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'users'")

	// 3. Re-seed user admin baru
	hash, _ := bcrypt.GenerateFromPassword([]byte("Agustin123"), bcrypt.DefaultCost)
	_, err := db.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)",
		"indra", "indra@livestock.com", string(hash), "admin")
	if err != nil {
		log.Printf("Gagal insert user admin: %v", err)
	}

	// 4. Re-seed device (pending — admin harus approve manual)
	_, _ = db.Exec("INSERT INTO devices (device_code, device_name, status, pairing_status) VALUES (?, ?, ?, ?)",
		"SCALE-ESP32-01", "Timbangan Utama Barn 1", "active", "pending")

	// 5. Re-seed 15 rumpun sapi + 3 penimbangan per sapi
	seeds := []struct {
		code, name, breed string
		w1, w2, w3        float64
	}{
		{"BALI-01", "Sapi Bali", "Sapi Bali", 280.0, 310.0, 345.0},
		{"MAD-01", "Sapi Madura", "Sapi Madura", 250.0, 275.0, 305.0},
		{"PO-01", "Sapi PO", "Peranakan Ongole", 320.0, 350.0, 385.0},
		{"LIM-01", "Sapi Limosin", "Sapi Limosin", 420.0, 460.0, 505.0},
		{"SIM-01", "Sapi Simental", "Sapi Simental", 430.0, 470.0, 515.0},
		{"FH-01", "Sapi FH", "Friesian Holstein", 380.0, 410.0, 445.0},
		{"BRH-01", "Sapi Brahman", "Sapi Brahman", 350.0, 385.0, 425.0},
		{"ACEH-01", "Sapi Aceh", "Sapi Aceh", 220.0, 245.0, 270.0},
		{"PES-01", "Sapi Pesisir", "Sapi Pesisir", 190.0, 210.0, 235.0},
		{"BRG-01", "Sapi Brangus", "Sapi Brangus", 390.0, 425.0, 465.0},
		{"WAG-01", "Sapi Wagyu", "Sapi Wagyu", 360.0, 395.0, 435.0},
		{"ANG-01", "Sapi Angus", "Black Angus", 400.0, 440.0, 485.0},
		{"HER-01", "Sapi Hereford", "Sapi Hereford", 410.0, 450.0, 495.0},
		{"CHA-01", "Sapi Charolais", "Sapi Charolais", 440.0, 485.0, 535.0},
		{"JAB-01", "Sapi Jabres", "Sapi Jabres", 260.0, 285.0, 315.0},
	}

	t1 := time.Now().AddDate(0, -2, 0).Format("2006-01-02 15:04:05")
	t2 := time.Now().AddDate(0, -1, 0).Format("2006-01-02 15:04:05")
	t3 := time.Now().Format("2006-01-02 15:04:05")

	for _, s := range seeds {
		res, err := db.Exec("INSERT INTO cows (cow_code, name, breed, gender, status) VALUES (?, ?, ?, 'jantan', 'active')",
			s.code, s.name, s.breed)
		if err == nil {
			cowID, _ := res.LastInsertId()
			adg1 := (s.w2 - s.w1) / 30.0
			adg2 := (s.w3 - s.w2) / 30.0
			_, _ = db.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'RESET-INIT')`, cowID, s.w1, 0.0, t1)
			_, _ = db.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'RESET-INIT')`, cowID, s.w2, adg1, t2)
			_, _ = db.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'RESET-INIT')`, cowID, s.w3, adg2, t3)
		}
	}

	log.Println("✅ RESET DATABASE selesai! Data baru telah di-seed.")

	return utils.Success(c, fiber.Map{
		"message":  "Database berhasil di-reset dan data sampel sudah di-seed ulang",
		"username": "indra",
		"password": "Agustin123",
		"device":   "SCALE-ESP32-01 (pending — approve manual di web)",
		"cows":     15,
	})
}

// ClearDevices POST /api/admin/clear-devices — Hapus HANYA data perangkat/pairing
func ClearDevices(c *fiber.Ctx) error {
	db := config.DB
	result, _ := db.Exec("DELETE FROM devices")
	rows, _ := result.RowsAffected()
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'devices'")
	log.Printf("🗑️ Semua data devices dihapus (%d baris)", rows)
	return utils.Success(c, fiber.Map{
		"message": "Semua data pairing/devices berhasil dihapus",
		"deleted": rows,
	})
}

// ClearWeighings POST /api/admin/clear-weighings — Hapus HANYA data timbangan dan prediksi
func ClearWeighings(c *fiber.Ctx) error {
	db := config.DB
	log.Println("⚠️ MENGHAPUS SEMUA DATA TIMBANGAN & PREDIKSI via API...")

	_, _ = db.Exec("DELETE FROM predictions")
	_, _ = db.Exec("DELETE FROM weight_records")
	
	// Reset auto increment untuk tabel yang dihapus
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'predictions'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'weight_records'")

	// Update kolom last_weight dan last_weighing_date di tabel cows menjadi null/0
	_, _ = db.Exec("UPDATE cows SET last_weight = 0, last_weighing_date = NULL")

	return utils.Success(c, fiber.Map{
		"message": "Semua data timbangan & prediksi berhasil dikosongkan.",
	})
}

// SeedDemo POST /api/admin/seed-demo
// Menghapus semua sapi & timbangan, lalu seed ulang hanya 3 sapi demo:
//   - Sapi Bali   → Layak Dipertahankan  (bobot naik pesat)
//   - Sapi Madura → Perlu Evaluasi       (bobot naik lambat)
//   - Sapi PO     → Tidak Layak          (bobot turun)
func SeedDemo(c *fiber.Ctx) error {
	db := config.DB
	log.Println("🌱 SEED DEMO dimulai via API...")

	// Hapus data lama (timbangan & sapi)
	_, _ = db.Exec("DELETE FROM predictions")
	_, _ = db.Exec("DELETE FROM weight_records")
	_, _ = db.Exec("DELETE FROM cows")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'predictions'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'weight_records'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'cows'")

	t1 := time.Now().AddDate(0, -2, 0).Format("2006-01-02 15:04:05")
	t2 := time.Now().AddDate(0, -1, 0).Format("2006-01-02 15:04:05")
	t3 := time.Now().Format("2006-01-02 15:04:05")

	// 3 sapi demo dengan data timbangan
	demoCows := []struct {
		code, name, breed, birthDate string
		w1, w2, w3                   float64 // bobot bulan -2, -1, sekarang
		label                        string
	}{
		// LAYAK DIPERTAHANKAN: bobot naik pesat (ADG > 0.5 kg/hari)
		// Sapi Bali usia ~24 bulan (2 tahun) — umur ideal penggemukan
		{"RFID-BALI-01", "Sapi Bali", "Bali", "2024-07-29", 340.0, 367.0, 400.0, "Layak Dipertahankan"},
		// PERLU EVALUASI: bobot naik sangat lambat (ADG ~0.13 kg/hari)
		// Sapi Madura usia ~18 bulan — masih muda tapi pertumbuhan stagnan
		{"RFID-MADURA-02", "Sapi Madura", "Madura", "2025-01-29", 290.0, 294.0, 298.0, "Perlu Evaluasi"},
		// TIDAK LAYAK: bobot turun (ADG negatif)
		// Sapi PO usia ~30 bulan (2.5 tahun) — sudah dewasa tapi berat turun
		{"RFID-PO-03", "Sapi PO", "Peranakan Ongole", "2024-01-29", 430.0, 418.0, 403.0, "Tidak Layak Dipertahankan"},
	}

	inserted := 0
	for _, s := range demoCows {
		res, err := db.Exec(
			"INSERT INTO cows (cow_code, name, breed, gender, birth_date, owner, status) VALUES (?, ?, ?, 'jantan', ?, 'Peternak Demo', 'active')",
			s.code, s.name, s.breed, s.birthDate,
		)
		if err != nil {
			log.Printf("Gagal insert sapi %s: %v", s.name, err)
			continue
		}
		cowID, _ := res.LastInsertId()
		adg1 := (s.w2 - s.w1) / 30.0
		adg2 := (s.w3 - s.w2) / 30.0
		_, _ = db.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'SEED-DEMO')`, cowID, s.w1, 0.0, t1)
		_, _ = db.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'SEED-DEMO')`, cowID, s.w2, adg1, t2)
		_, _ = db.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'SEED-DEMO')`, cowID, s.w3, adg2, t3)
		log.Printf("✓ %s seeded → %s", s.name, s.label)
		inserted++
	}

	log.Println("✅ SEED DEMO selesai!")
	return utils.Success(c, fiber.Map{
		"message": "Seed demo berhasil! 3 sapi demo telah di-insert.",
		"sapi": []fiber.Map{
			{"nama": "Sapi Bali", "kategori": "Layak Dipertahankan", "bobot": "340→367→400 kg"},
			{"nama": "Sapi Madura", "kategori": "Perlu Evaluasi", "bobot": "290→294→298 kg"},
			{"nama": "Sapi PO", "kategori": "Tidak Layak Dipertahankan", "bobot": "430→418→403 kg"},
		},
		"inserted": inserted,
	})
}
