//go:build ignore
// +build ignore

package main

import (
	"log"

	"smart-livestock-backend/config"
)

func main() {
	log.Println("Membersihkan SEMUA data (Sapi, Timbangan, Perangkat, DAN User)...")

	config.LoadConfig()
	config.InitializeDatabase()
	defer config.CloseDatabase()

	db := config.DB

	// Hapus semua log timbangan, prediksi, perangkat, DAN user
	_, _ = db.Exec("DELETE FROM predictions")
	_, _ = db.Exec("DELETE FROM devices")
	_, err := db.Exec("DELETE FROM weight_records")
	if err != nil {
		log.Fatalf("Gagal menghapus weight_records: %v", err)
	}
	log.Println("Berhasil menghapus seluruh data histori timbangan & prediksi.")

	// Hapus semua data sapi
	_, err = db.Exec("DELETE FROM cows")
	if err != nil {
		log.Fatalf("Gagal menghapus data cows: %v", err)
	}
	log.Println("Berhasil menghapus seluruh data sapi.")

	// Hapus semua data user (agar autoSeedIfEmpty bisa membuat ulang user admin baru)
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		log.Fatalf("Gagal menghapus data users: %v", err)
	}
	log.Println("Berhasil menghapus seluruh data user.")

	// Reset auto increment (opsional)
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'weight_records'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'cows'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'devices'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'predictions'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'users'")

	log.Println("Pembersihan selesai! SEMUA data (sapi, timbangan, perangkat, user) sekarang kosong.")
	log.Println("Saat server restart, autoSeedIfEmpty akan membuat ulang data sampel + user admin.")
}
