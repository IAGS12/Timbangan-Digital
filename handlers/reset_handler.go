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
