package routes

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/handlers"
	"smart-livestock-backend/middleware"
)

// RegisterUserRoutes mendaftarkan endpoint manajemen user (khusus admin)
func RegisterUserRoutes(protected fiber.Router, userHandler *handlers.UserHandler) {
	// Endpoint Profil — tersedia untuk semua user yang login
	protected.Get("/profile", userHandler.GetProfile)
	protected.Put("/profile", userHandler.UpdateProfile)

	// Endpoint Admin
	adminGroup := protected.Group("/users", middleware.AdminOnly)
	adminGroup.Get("/", userHandler.GetAll)
	adminGroup.Put("/:id/role", userHandler.UpdateRole)
}
