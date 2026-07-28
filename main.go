package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"time"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"smart-livestock-backend/config"
	"smart-livestock-backend/routes"
)

func main() {
	// 1. Muat Konfigurasi dari .env
	config.LoadConfig()

	// 2. Inisialisasi Database SQLite
	config.InitializeDatabase()
	// Migrasi: Set pairing_status=approved untuk device lama yang sudah ada (NULL)
	_, _ = config.DB.Exec(`UPDATE devices SET pairing_status = 'approved' WHERE pairing_status IS NULL OR pairing_status = ''`)
	
	// Auto-seed data sampel jika database masih kosong
	autoSeedIfEmpty()

	// Pastikan database ditutup saat aplikasi mati
	defer config.CloseDatabase()

	// 3. Buat Aplikasi Fiber
	app := fiber.New(fiber.Config{
		AppName:           "Smart Livestock Management API",
		EnablePrintRoutes: false,
	})

	// 4. Pasang Middleware Global
	app.Use(recover.New()) // Mencegah server crash jika ada panic
	app.Use(logger.New())  // Logging setiap request HTTP
	allowOrigins := config.Config.AllowedOrigins
	if allowOrigins == "" {
		allowOrigins = "*"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowCredentials: true,
	}))

	// 5. Endpoint Cek Status Server (Health Check)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "Running",
			"message": "Smart Livestock API Backend (Golang Fiber) is up!",
		})
	})

	// 6. Setup Semua Routes API
	routes.SetupRoutes(app)

	// 7. Handle Graceful Shutdown (Bisa dihentikan dengan rapi pakai Ctrl+C)
	go func() {
		port := ":" + config.Config.Port
		log.Printf("🚀 Server berjalan di http://localhost%s", port)
		if err := app.Listen(port); err != nil {
			log.Fatalf("❌ Gagal menjalankan server: %v", err)
		}
	}()

	// Menunggu sinyal interupsi sistem
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	_ = <-c // Block eksekusi sampai ada sinyal

	log.Println("Gracefully shutting down...")
	_ = app.Shutdown()
	log.Println("Server dihentikan.")
}

func autoSeedIfEmpty() {
	var count int
	err := config.DB.Get(&count, "SELECT COUNT(*) FROM cows")
	if err == nil && count == 0 {
		log.Println("🌱 Database sapi kosong di Railway server, menjalankan auto-seed data sampel...")
		
		// Insert default admin user if not exists
		var userCount int
		_ = config.DB.Get(&userCount, "SELECT COUNT(*) FROM users WHERE username = 'indra'")
		if userCount == 0 {
			hash, _ := bcrypt.GenerateFromPassword([]byte("Agustin123"), bcrypt.DefaultCost)
			_, _ = config.DB.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)", "indra", "indra@livestock.com", string(hash), "admin")
		}

		// Insert default device if not exists
		var devCount int
		_ = config.DB.Get(&devCount, "SELECT COUNT(*) FROM devices WHERE device_code = 'SCALE-ESP32-01'")
		if devCount == 0 {
			_, _ = config.DB.Exec("INSERT INTO devices (device_code, device_name, status, pairing_status) VALUES (?, ?, ?, ?)", "SCALE-ESP32-01", "Timbangan Utama Barn 1", "active", "pending")
		}

		// Insert default cows
		cows := []struct {
			Code, Name, Breed, Gender, Birth, Owner, Status string
		}{
			{"RFID-0001-TEST", "Sapi Brahman A", "Brahman", "jantan", "2024-01-01", "Peternak A", "active"},
			{"RFID-0002-EVAL", "Sapi Brahman B", "Brahman", "jantan", "2024-02-15", "Peternak B", "active"},
			{"RFID-0003-FAIL", "Sapi Brahman C", "Brahman", "jantan", "2024-03-10", "Peternak C", "active"},
		}

		now := time.Now()
		for _, c := range cows {
			res, err := config.DB.Exec("INSERT INTO cows (cow_code, name, breed, gender, birth_date, owner, status) VALUES (?, ?, ?, ?, ?, ?, ?)", c.Code, c.Name, c.Breed, c.Gender, c.Birth, c.Owner, c.Status)
			if err == nil {
				cowID, _ := res.LastInsertId()
				// Seed 3 weighings per cow
				var baseWeight, adg float64
				if c.Code == "RFID-0001-TEST" { baseWeight = 320.0; adg = 1.2 }
				if c.Code == "RFID-0002-EVAL" { baseWeight = 320.0; adg = 0.5 }
				if c.Code == "RFID-0003-FAIL" { baseWeight = 320.0; adg = -0.3 }

				for month := 0; month < 3; month++ {
					wTime := now.AddDate(0, -(2 - month), 0)
					weight := baseWeight + (float64(month) * 30 * adg)
					var actualAdg float64
					if month > 0 { actualAdg = adg }
					config.DB.Exec("INSERT INTO weight_records (cow_id, device_code, weight, adg, measurement_date, status) VALUES (?, ?, ?, ?, ?, ?)", cowID, "SCALE-ESP32-01", weight, actualAdg, wTime.Format("2006-01-02 15:04:05"), "valid")
				}
			}
		}
		log.Println("✅ Auto-seed data sampel selesai!")
	}
}
