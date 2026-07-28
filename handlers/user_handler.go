package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/repositories"
	"smart-livestock-backend/utils"
)

type UserHandler struct {
	userRepo *repositories.UserRepository
}

func NewUserHandler(userRepo *repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// GetAll GET /api/users (admin only)
func (h *UserHandler) GetAll(c *fiber.Ctx) error {
	users, err := h.userRepo.GetAll()
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil data user")
	}
	return utils.Success(c, users)
}

// UpdateRole PUT /api/users/:id/role (admin only)
func (h *UserHandler) UpdateRole(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return utils.BadRequest(c, "ID tidak valid")
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.BadRequest(c, "Format request tidak valid")
	}

	if err := utils.ValidateRole(body.Role); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.userRepo.UpdateRole(id, body.Role); err != nil {
		return utils.InternalError(c, "Gagal mengubah role")
	}

	return utils.Success(c, fiber.Map{"message": "Role berhasil diubah"})
}
