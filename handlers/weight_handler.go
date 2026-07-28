package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/models"
	"smart-livestock-backend/services"
	"smart-livestock-backend/utils"
)

type WeightHandler struct {
	weightService *services.WeightService
	deviceService *services.DeviceService
}

func NewWeightHandler(weightService *services.WeightService, deviceService *services.DeviceService) *WeightHandler {
	return &WeightHandler{weightService: weightService, deviceService: deviceService}
}

// AddWeight POST /api/weighings
func (h *WeightHandler) AddWeight(c *fiber.Ctx) error {
	var req models.WeightRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format request tidak valid")
	}

	// Cek apakah device sudah di-approve (pairing)
	if req.DeviceID != "" {
		if !h.deviceService.IsDeviceApproved(req.DeviceID) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Perangkat belum terdaftar atau belum disetujui. Lakukan pairing terlebih dahulu melalui menu ESP32.",
			})
		}
	}

	// Ambil operator ID dari token jika ada (opsional untuk ESP32)
	var operatorID *int64
	if uid, ok := c.Locals("userID").(int64); ok {
		operatorID = &uid
	}

	record, err := h.weightService.AddWeight(req, operatorID)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	// Broadcast real-time event ke WebSocket listener (Web Dashboard Live Mirroring)
	services.GlobalWSHub.Broadcast(fiber.Map{
		"type":        "NEW_WEIGHT_RECORD",
		"device_code": req.DeviceID,
		"data":        record,
	})

	return utils.Created(c, "Data penimbangan berhasil disimpan", record)
}

// GetHistory GET /api/weighings?cow_id=1&start_date=2026-01-01&end_date=2026-07-18
func (h *WeightHandler) GetHistory(c *fiber.Ctx) error {
	cowID, _ := strconv.ParseInt(c.Query("cow_id", "0"), 10, 64)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	records, err := h.weightService.GetHistory(cowID, startDate, endDate)
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil riwayat penimbangan")
	}

	return utils.Success(c, records)
}

// GetCowWeights GET /api/cows/:id/weights
func (h *WeightHandler) GetCowWeights(c *fiber.Ctx) error {
	cowID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return utils.BadRequest(c, "ID tidak valid")
	}

	records, err := h.weightService.GetCowWeights(cowID)
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil data penimbangan")
	}

	return utils.Success(c, records)
}

// BatchUpload POST /api/weighings/batch
// PUBLIC — dipanggil ESP32 untuk upload offline queue sekaligus
func (h *WeightHandler) BatchUpload(c *fiber.Ctx) error {
	var req models.BatchWeighingRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format request tidak valid")
	}

	if req.DeviceCode == "" {
		return utils.BadRequest(c, "device_code wajib diisi")
	}
	if len(req.Records) == 0 {
		return utils.BadRequest(c, "records tidak boleh kosong")
	}
	if len(req.Records) > 100 {
		return utils.BadRequest(c, "Maksimal 100 record per batch")
	}

	// Cek device approved sebelum batch upload
	if !h.deviceService.IsDeviceApproved(req.DeviceCode) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Perangkat belum disetujui. Lakukan pairing terlebih dahulu.",
		})
	}

	result := h.weightService.AddBatch(req)

	// Broadcast WebSocket sync notification ke Web Dashboard
	services.GlobalWSHub.Broadcast(fiber.Map{
		"type":        "SYNC_NOTIF",
		"device_code": req.DeviceCode,
		"saved_count": result.Saved,
		"skipped":     result.Skipped,
		"total":       result.Total,
		"message":     "Sync otomatis dari ESP32 selesai",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Batch upload selesai",
		"data":    result,
	})
}

