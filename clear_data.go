//go:build ignore
// +build ignore

package main

import (
	"log"

	"smart-livestock-backend/config"
)

func main() {
	log.Println("Membersihkan data Sapi dan Riwayat Timbangan...")

	config.LoadConfig()
	config.InitializeDatabase()
	defer config.CloseDatabase()

	db := config.DB

	// Hapus semua log timbangan, prediksi, dan perangkat
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

	// Reset auto increment (opsional)
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'weight_records'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'cows'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'devices'")
	db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'predictions'")

	log.Println("Pembersihan selesai! Data sapi, timbangan, dan perangkat sekarang kosong.")
}
