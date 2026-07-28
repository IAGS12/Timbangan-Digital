package config

import (
	"log"
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
		}{
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
			_, _ = DB.Exec(`INSERT INTO cows (cow_code, name, breed, gender, status) VALUES (?, ?, ?, 'jantan', 'active')`, s.code, s.name, s.breed)
		}
		log.Println("15 Rumpun Sapi standar berhasil di-seed ke database.")
	}

	log.Println("Semua tabel database siap.")
}

func CloseDatabase() {
	if DB != nil {
		DB.Close()
		log.Println("Koneksi database ditutup.")
	}
}