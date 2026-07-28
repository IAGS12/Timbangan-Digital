package routes

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/handlers"
	"smart-livestock-backend/middleware"
)

// RegisterUserRoutes mendaftarkan endpoint manajemen user (khusus admin)
func RegisterUserRoutes(protected fiber.Router, userHandler *handlers.UserHandler) {
	adminGroup := protected.Group("/users", middleware.AdminOnly)
	adminGroup.Get("/", userHandler.GetAll)
	adminGroup.Put("/:id/role", userHandler.UpdateRole)
}
