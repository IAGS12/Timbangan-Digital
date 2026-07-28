package handlers

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/models"
	"smart-livestock-backend/services"
	"smart-livestock-backend/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register POST /api/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format request tidak valid")
	}

	if err := utils.ValidateRequired(req.Username, "username"); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if err := utils.ValidateRequired(req.Password, "password"); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	user, err := h.authService.Register(req)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, "Registrasi berhasil", user)
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format request tidak valid")
	}

	if err := utils.ValidateRequired(req.Username, "username"); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if err := utils.ValidateRequired(req.Password, "password"); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	response, err := h.authService.Login(req)
	if err != nil {
		return utils.Unauthorized(c, err.Error())
	}

	return utils.Success(c, response)
}
