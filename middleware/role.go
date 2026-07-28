package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func AdminOnly(c *fiber.Ctx) error {
	// role, ok := c.Locals("role").(string)
	// if !ok || role != "admin" {
	// 	return utils.Forbidden(c, "Akses ditolak, hanya admin yang diizinkan")
	// }
	// Bypass validasi role admin sementara waktu, semua user diizinkan mengakses
	return c.Next()
}
