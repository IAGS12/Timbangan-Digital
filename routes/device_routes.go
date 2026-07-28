package routes

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/handlers"
)

// RegisterDeviceRoutes mendaftarkan endpoint manajemen perangkat timbangan (hanya protected)
// NOTE: Public endpoints (pair & pairing-status) sudah didaftarkan langsung di routes.go
func RegisterDeviceRoutes(protected fiber.Router, deviceHandler *handlers.DeviceHandler) {
	devicesGroup := protected.Group("/devices")
	devicesGroup.Get("/", deviceHandler.GetAllDevices)
	devicesGroup.Get("/pending", deviceHandler.GetPendingDevices)
	devicesGroup.Put("/:id/pairing", deviceHandler.ApprovePairing)
	devicesGroup.Post("/claim", deviceHandler.ClaimDevice)
	devicesGroup.Delete("/:id", deviceHandler.UnlinkDevice)
}
