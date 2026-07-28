package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/models"
	"smart-livestock-backend/services"
	"smart-livestock-backend/utils"
)

type CowHandler struct {
	cowService *services.CowService
}

func NewCowHandler(cowService *services.CowService) *CowHandler {
	return &CowHandler{cowService: cowService}
}

// GetAll GET /api/cows?status=active&breed=limousin&page=1&limit=10
func (h *CowHandler) GetAll(c *fiber.Ctx) error {
	status := c.Query("status")
	breed := c.Query("breed")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	cows, total, err := h.cowService.GetAll(status, breed, page, limit)
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil data sapi")
	}

	return utils.Paginated(c, cows, page, limit, total)
}

// GetByID GET /api/cows/:id
func (h *CowHandler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "sync" {
		return h.SyncCows(c)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.BadRequest(c, "ID tidak valid")
	}

	cow, err := h.cowService.GetByID(id)
	if err != nil {
		return utils.NotFound(c, err.Error())
	}

	return utils.Success(c, cow)
}

// Create POST /api/cows
func (h *CowHandler) Create(c *fiber.Ctx) error {
	var req models.CowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format request tidak valid")
	}

	id, err := h.cowService.Create(req)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, "Data sapi berhasil ditambahkan", fiber.Map{
		"id":       id,
		"cow_code": req.CowCode,
	})
}

// Update PUT /api/cows/:id
func (h *CowHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return utils.BadRequest(c, "ID tidak valid")
	}

	var req models.CowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format request tidak valid")
	}

	if err := h.cowService.Update(id, req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Success(c, fiber.Map{"message": "Data sapi berhasil diperbarui"})
}

// Delete DELETE /api/cows/:id
func (h *CowHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return utils.BadRequest(c, "ID tidak valid")
	}

	if err := h.cowService.Delete(id); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Success(c, fiber.Map{"message": "Data sapi berhasil dihapus"})
}

// SyncCows GET /api/cows/sync?device_code=SCALE-ESP32-01
// PUBLIC — dipanggil ESP32 untuk download cache daftar sapi ke NVS/LittleFS
func (h *CowHandler) SyncCows(c *fiber.Ctx) error {
	deviceCode := c.Query("device_code")
	if deviceCode == "" {
		return utils.BadRequest(c, "Parameter device_code wajib diisi")
	}

	cows, err := h.cowService.GetCowsForSync(deviceCode)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":     true,
		"device_code": deviceCode,
		"count":       len(cows),
		"cows":        cows,
	})
}

