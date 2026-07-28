package routes

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/handlers"
)

// RegisterWeightRoutes mendaftarkan endpoint penimbangan
func RegisterWeightRoutes(api fiber.Router, protected fiber.Router, weightHandler *handlers.WeightHandler) {
	// Endpoint ESP32 (publik, idealnya pakai API Key)
	api.Post("/weighings", weightHandler.AddWeight)

	// Histori Penimbangan (terproteksi)
	protected.Get("/weighings", weightHandler.GetHistory)
}
