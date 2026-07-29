package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"smart-livestock-backend/models"
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

// GetProfile GET /api/profile — Ambil profil user yang sedang login
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return utils.BadRequest(c, "Sesi tidak valid")
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.NotFound(c, "User tidak ditemukan")
	}

	return utils.Success(c, fiber.Map{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

// UpdateProfile PUT /api/profile — Update username dan/atau password
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return utils.BadRequest(c, "Sesi tidak valid")
	}

	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format request tidak valid")
	}

	// Ambil data user lama untuk verifikasi
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.NotFound(c, "User tidak ditemukan")
	}

	updated := false

	// --- Update Username ---
	newUsername := strings.TrimSpace(req.Username)
	if newUsername != "" && newUsername != user.Username {
		if len(newUsername) < 3 {
			return utils.BadRequest(c, "Username minimal 3 karakter")
		}
		// Cek apakah username sudah dipakai user lain
		exists, err := h.userRepo.UsernameExists(newUsername, userID)
		if err != nil {
			return utils.InternalError(c, "Gagal validasi username")
		}
		if exists {
			return utils.BadRequest(c, "Username '"+newUsername+"' sudah digunakan oleh akun lain")
		}
		if err := h.userRepo.UpdateUsername(userID, newUsername); err != nil {
			return utils.InternalError(c, "Gagal mengubah username")
		}
		user.Username = newUsername
		updated = true
	}

	// --- Update Password ---
	if req.NewPassword != "" {
		// Harus menyertakan password lama untuk verifikasi
		if req.OldPassword == "" {
			return utils.BadRequest(c, "Password lama wajib diisi untuk mengganti password")
		}
		// Verifikasi password lama
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
			return utils.BadRequest(c, "Password lama tidak sesuai")
		}
		if len(req.NewPassword) < 6 {
			return utils.BadRequest(c, "Password baru minimal 6 karakter")
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return utils.InternalError(c, "Gagal memproses password baru")
		}
		if err := h.userRepo.UpdatePassword(userID, string(hashed)); err != nil {
			return utils.InternalError(c, "Gagal mengubah password")
		}
		updated = true
	}

	if !updated {
		return utils.BadRequest(c, "Tidak ada perubahan yang dilakukan")
	}

	return utils.Success(c, fiber.Map{
		"message":  "Profil berhasil diperbarui",
		"username": user.Username,
		"id":       user.ID,
		"role":     user.Role,
	})
}
