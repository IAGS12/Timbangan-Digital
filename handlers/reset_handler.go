package handlers

import (
	"log"

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

	// 4. Re-seed device approved
	_, _ = db.Exec("INSERT INTO devices (device_code, device_name, status, pairing_status) VALUES (?, ?, ?, ?)",
		"SCALE-ESP32-01", "Timbangan Utama Barn 1", "active", "approved")

	// 5. Re-seed 15 rumpun sapi
	seeds := []struct{ code, name, breed string }{
		{"BALI-01", "Sapi Bali", "Sapi Bali"},
		{"MAD-01", "Sapi Madura", "Sapi Madura"},
		{"PO-01", "Sapi PO", "Peranakan Ongole"},
		{"LIM-01", "Sapi Limosin", "Sapi Limosin"},
		{"SIM-01", "Sapi Simental", "Sapi Simental"},
		{"FH-01", "Sapi FH", "Friesian Holstein"},
		{"BRH-01", "Sapi Brahman", "Sapi Brahman"},
		{"ACEH-01", "Sapi Aceh", "Sapi Aceh"},
		{"PES-01", "Sapi Pesisir", "Sapi Pesisir"},
		{"BRG-01", "Sapi Brangus", "Sapi Brangus"},
		{"WAG-01", "Sapi Wagyu", "Sapi Wagyu"},
		{"ANG-01", "Sapi Angus", "Black Angus"},
		{"HER-01", "Sapi Hereford", "Sapi Hereford"},
		{"CHA-01", "Sapi Charolais", "Sapi Charolais"},
		{"JAB-01", "Sapi Jabres", "Sapi Jabres"},
	}
	for _, s := range seeds {
		_, _ = db.Exec("INSERT INTO cows (cow_code, name, breed, gender, status) VALUES (?, ?, ?, 'jantan', 'active')",
			s.code, s.name, s.breed)
	}

	log.Println("✅ RESET DATABASE selesai! Data baru telah di-seed.")

	return utils.Success(c, fiber.Map{
		"message":  "Database berhasil di-reset dan data sampel sudah di-seed ulang",
		"username": "indra",
		"password": "Agustin123",
		"device":   "SCALE-ESP32-01 (approved)",
		"cows":     15,
	})
}
