package routes

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/handlers"
)


func RegisterAuthRoutes(api fiber.Router, authHandler *handlers.AuthHandler) {
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
}
