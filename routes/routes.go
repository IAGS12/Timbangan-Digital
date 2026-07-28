package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"smart-livestock-backend/config"
	"smart-livestock-backend/handlers"
	"smart-livestock-backend/middleware"
	"smart-livestock-backend/repositories"
	"smart-livestock-backend/services"
)

// SetupRoutes melakukan inisialisasi semua layer dan mendaftarkan route
func SetupRoutes(app *fiber.App) {
	// ==========================================
	// 1. Inisialisasi Repositories
	// ==========================================
	userRepo := repositories.NewUserRepository(config.DB)
	cowRepo := repositories.NewCowRepository(config.DB)
	weightRepo := repositories.NewWeightRepository(config.DB)
	deviceRepo := repositories.NewDeviceRepository(config.DB)

	// ==========================================
	// 2. Inisialisasi Services
	// ==========================================
	authService := services.NewAuthService(userRepo)
	cowService := services.NewCowService(cowRepo)
	weightService := services.NewWeightService(weightRepo, cowRepo)
	predictionService := services.NewPredictionService(weightRepo)
	decisionService := services.NewDecisionService()
	deviceService := services.NewDeviceService(deviceRepo)

	// ==========================================
	// 3. Inisialisasi Handlers
	// ==========================================
	authHandler := handlers.NewAuthHandler(authService)
	cowHandler := handlers.NewCowHandler(cowService)
	weightHandler := handlers.NewWeightHandler(weightService, deviceService)
	predictionHandler := handlers.NewPredictionHandler(predictionService, decisionService)
	dashboardHandler := handlers.NewDashboardHandler(cowRepo, weightRepo)
	userHandler := handlers.NewUserHandler(userRepo)
	deviceHandler := handlers.NewDeviceHandler(deviceService)
	exportHandler := handlers.NewExportHandler(cowRepo, weightRepo, predictionService, decisionService)

	// ==========================================
	// 4. Mendaftarkan Endpoint (Routes)
	// ==========================================

	// WebSocket Endpoint (Public Real-Time Push Stream)
	app.Use("/ws", handlers.UpgradeToWebSocket)
	app.Get("/ws", websocket.New(handlers.HandleWebSocket))

	// Public Endpoints (ESP32 & Auth — Bypass AuthMiddleware)
	app.Get("/api/cows", cowHandler.GetAll)
	app.Post("/api/weighings", weightHandler.AddWeight)
	app.Post("/api/weighings/batch", weightHandler.BatchUpload)
	app.Get("/api/cows/sync", cowHandler.SyncCows)
	app.Post("/api/devices/pair", deviceHandler.RequestPairing)
	app.Get("/api/devices/pairing-status/:device_code", deviceHandler.GetPairingStatus)
	app.Post("/api/devices/command", deviceHandler.SendCommand)
	app.Get("/api/devices/command/:device_code", deviceHandler.GetPendingCommand)
	app.Post("/api/devices/live", deviceHandler.PostLiveWeight)
	api := app.Group("/api")
	RegisterAuthRoutes(api, authHandler)       // Public: ESP32 poll status pairing
	api.Get("/export/excel", exportHandler.ExportExcel) // Public fallback for Excel export
	api.Post("/admin/reset-db", handlers.ResetDatabase)  // Dev/Debug: Reset seluruh database

	// Endpoint Terproteksi (Wajib Login) — HARUS didaftarkan SETELAH route public di atas
	protected := app.Group("/api", middleware.AuthMiddleware)
	RegisterCowRoutes(protected, cowHandler, predictionHandler, weightHandler)
	protected.Get("/weighings", weightHandler.GetHistory)
	protected.Get("/export/excel", exportHandler.ExportExcel)
	RegisterDashboardRoutes(protected, dashboardHandler)
	RegisterUserRoutes(protected, userHandler)
	RegisterDeviceRoutes(protected, deviceHandler)
}
