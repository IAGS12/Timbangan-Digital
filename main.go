package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Buka untuk semua origin (Frontend React & ESP32)
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
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
