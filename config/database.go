package config

import (
	"log"
	"time"
	"github.com/jmoiron/sqlx"
	_"modernc.org/sqlite"
)

var DB *sqlx.DB

func InitializeDatabase() {
	var err error
	connStr := Config.DBPath
	if connStr == "" {
		connStr = "./timbangan.db"
	}
	connStr += "?_busy_timeout=5000&_journal_mode=WAL"

	DB, err = sqlx.Connect("sqlite", connStr)
	if err !=nil {
		log.Fatalf("Gagal membuka database %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf(" Gagal Koneksi ke Database: %v", err)
	}

	log.Println(" Koneksi ke database SQLite berhasil.")

	DB.MustExec("PRAGMA foreign_keys = ON;")

	createTables()
	
}

func createTables() {
	DB.MustExec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE,
			password TEXT NOT NULL,
			role TEXT DEFAULT 'peternak' CHECK(role IN ('admin', 'peternak')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	log.Println("Tabel 'users' siap.")
	// Tabel Cows
	DB.MustExec(`
		CREATE TABLE IF NOT EXISTS cows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cow_code TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			breed TEXT NOT NULL,
			gender TEXT CHECK(gender IN ('jantan', 'betina')),
			birth_date DATE,
			owner TEXT,
			status TEXT DEFAULT 'active' CHECK(status IN ('active', 'sold', 'deceased')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	log.Println("Tabel 'cows' siap.")
	DB.MustExec(`
		CREATE TABLE IF NOT EXISTS weight_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cow_id INTEGER NOT NULL,
			weight REAL NOT NULL,
			adg REAL,
			measurement_date DATETIME DEFAULT CURRENT_TIMESTAMP,
			device_id TEXT,
			operator_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (cow_id) REFERENCES cows(id),
			FOREIGN KEY (operator_id) REFERENCES users(id)
		)
	`)
	log.Println("Tabel 'weight_records' siap.")
	DB.MustExec(`
		CREATE TABLE IF NOT EXISTS predictions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cow_id INTEGER NOT NULL,
			prediction_weight REAL NOT NULL,
			prediction_date DATE NOT NULL,
			horizon_days INTEGER DEFAULT 30,
			trend TEXT CHECK(trend IN ('meningkat', 'stagnan', 'menurun')),
			r_squared REAL,
			recommendation TEXT CHECK(recommendation IN ('LAYAK_DIPERTAHANKAN', 'PERLU_EVALUASI', 'TIDAK_LAYAK_DIPERTAHANKAN')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (cow_id) REFERENCES cows(id)
		)
	`)
	log.Println("Tabel 'predictions' siap.")

	DB.MustExec(`
		CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_code TEXT UNIQUE NOT NULL,
			device_name TEXT NOT NULL,
			user_id INTEGER,
			status TEXT DEFAULT 'active' CHECK(status IN ('active', 'inactive')),
			pairing_status TEXT DEFAULT 'pending' CHECK(pairing_status IN ('pending', 'approved', 'rejected', 'unpaired')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	// Migrasi: tambahkan kolom pairing_status jika belum ada (untuk database lama)
	_, _ = DB.Exec(`ALTER TABLE devices ADD COLUMN pairing_status TEXT DEFAULT 'pending'`)
	log.Println("Tabel 'devices' siap.")

	// Seed 15 Rumpun Sapi default jika tabel cows masih kosong
	var cowCount int
	_ = DB.Get(&cowCount, "SELECT COUNT(*) FROM cows")
	if cowCount == 0 {
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
			res, err := DB.Exec(`INSERT INTO cows (cow_code, name, breed, gender, status) VALUES (?, ?, ?, 'jantan', 'active')`, s.code, s.name, s.breed)
			if err == nil {
				cowID, _ := res.LastInsertId()
				adg1 := (s.w2 - s.w1) / 30.0
				adg2 := (s.w3 - s.w2) / 30.0
				_, _ = DB.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'SEEDER-INIT')`, cowID, s.w1, 0.0, t1)
				_, _ = DB.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'SEEDER-INIT')`, cowID, s.w2, adg1, t2)
				_, _ = DB.Exec(`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id) VALUES (?, ?, ?, ?, 'SEEDER-INIT')`, cowID, s.w3, adg2, t3)
			}
		}
		log.Println("15 Rumpun Sapi standar + 3 penimbangan awal per sapi berhasil di-seed ke database.")
	}

	log.Println("Semua tabel database siap.")
}

func CloseDatabase() {
	if DB != nil {
		DB.Close()
		log.Println("Koneksi database ditutup.")
	}
}