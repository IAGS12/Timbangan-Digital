package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/utils"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.Unauthorized(c, "Token tidak ditemukan")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return utils.Unauthorized(c, "Format token tidak valid")
	}

	claims, err := utils.ParseToken(parts[1])
	if err != nil {
		return utils.Unauthorized(c, "Token tidak valid atau sudah kadaluarsa")
	}

	c.Locals("userID", claims.UserID)
	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)

	return c.Next()
}
