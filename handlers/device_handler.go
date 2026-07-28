package handlers

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/models"
	"smart-livestock-backend/services"
	"smart-livestock-backend/utils"
)

type DeviceHandler struct {
	deviceService *services.DeviceService
}

func NewDeviceHandler(deviceService *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceService: deviceService}
}

// GetAllDevices GET /api/devices — admin view semua perangkat
func (h *DeviceHandler) GetAllDevices(c *fiber.Ctx) error {
	devices, err := h.deviceService.GetAllDevices()
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil daftar perangkat")
	}
	return utils.Success(c, devices)
}

// GetMyDevices GET /api/devices/me
func (h *DeviceHandler) GetMyDevices(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		userID = 1
	}
	devices, err := h.deviceService.GetUserDevices(userID)
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil daftar perangkat")
	}
	return utils.Success(c, devices)
}

// GetPendingDevices GET /api/devices/pending — ambil perangkat menunggu persetujuan
func (h *DeviceHandler) GetPendingDevices(c *fiber.Ctx) error {
	devices, err := h.deviceService.GetPendingDevices()
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil daftar perangkat pending")
	}
	return utils.Success(c, devices)
}

// RequestPairing POST /api/devices/pair — PUBLIC, dipanggil ESP32
func (h *DeviceHandler) RequestPairing(c *fiber.Ctx) error {
	var req models.PairingRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format data tidak valid")
	}

	if err := h.deviceService.RequestPairing(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, "Permintaan pairing berhasil dikirim, menunggu persetujuan admin", nil)
}

// GetPairingStatus GET /api/devices/pairing-status/:device_code — PUBLIC, dipoll ESP32
func (h *DeviceHandler) GetPairingStatus(c *fiber.Ctx) error {
	deviceCode := c.Params("device_code")
	if deviceCode == "" {
		return utils.BadRequest(c, "device_code tidak boleh kosong")
	}

	status, err := h.deviceService.GetPairingStatus(deviceCode)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success":        true,
			"pairing_status": "unknown",
			"device_code":    deviceCode,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":        true,
		"pairing_status": status,
		"device_code":    deviceCode,
	})
}

// ApprovePairing PUT /api/devices/:id/pairing — Admin approve/reject
func (h *DeviceHandler) ApprovePairing(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return utils.BadRequest(c, "ID perangkat tidak valid")
	}

	var req models.ApproveDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format data tidak valid")
	}

	if err := h.deviceService.ApprovePairing(id, req.PairingStatus); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	// Notify ESP32 over MQTT
	if dev, err := h.deviceService.GetDeviceByID(id); err == nil && dev != nil && services.GlobalMQTT != nil {
		_ = services.GlobalMQTT.PublishPairingStatus(dev.DeviceCode, req.PairingStatus)
	}

	msg := "Perangkat berhasil disetujui"
	if req.PairingStatus == "rejected" {
		msg = "Perangkat berhasil ditolak"
	} else if req.PairingStatus == "unpaired" {
		msg = "Akses perangkat berhasil dicabut (Unpaired)"
	}
	return utils.Success(c, msg)
}

// ClaimDevice POST /api/devices/claim
func (h *DeviceHandler) ClaimDevice(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		userID = 1
	}

	var req models.ClaimDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format data tidak valid")
	}

	if err := h.deviceService.ClaimDevice(req, userID); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, "Perangkat berhasil dihubungkan", nil)
}

// UnlinkDevice DELETE /api/devices/:id
func (h *DeviceHandler) UnlinkDevice(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return utils.BadRequest(c, "ID perangkat tidak valid")
	}

	if err := h.deviceService.DeleteDevice(id); err != nil {
		return utils.InternalError(c, "Gagal menghapus perangkat")
	}

	return utils.Success(c, "Perangkat berhasil dihapus")
}

type RemoteCommand struct {
	Action  string `json:"action"`
	CowCode string `json:"cow_code,omitempty"`
	CowName string `json:"cow_name,omitempty"`
	CowID   int64  `json:"cow_id,omitempty"`
}

// Storage sederhana untuk antrean perintah remote per device_code (untuk polling HTTP ESP32)
var (
	pendingCommandsMutex sync.Mutex
	pendingCommands      = make(map[string]RemoteCommand)
)

// SendCommand POST /api/devices/command — kirim perintah remote dari Web ke ESP32
func (h *DeviceHandler) SendCommand(c *fiber.Ctx) error {
	var req struct {
		DeviceCode string `json:"device_code"`
		Action     string `json:"action"` // "tare", "buzzer", "backlight_toggle", "select_cow", "save_weight", "calibrate"
		CowCode    string `json:"cow_code,omitempty"`
		CowName    string `json:"cow_name,omitempty"`
		CowID      int64  `json:"cow_id,omitempty"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format data tidak valid")
	}
	if req.DeviceCode == "" || req.Action == "" {
		return utils.BadRequest(c, "device_code dan action wajib diisi")
	}

	// Wajib cek status pairing perangkat
	if !h.deviceService.IsDeviceApproved(req.DeviceCode) {
		return utils.Forbidden(c, "Perangkat belum disetujui / pairing tidak aktif")
	}

	cmdObj := RemoteCommand{
		Action:  req.Action,
		CowCode: req.CowCode,
		CowName: req.CowName,
		CowID:   req.CowID,
	}

	// Simpan di antrean pending untuk di-poll oleh ESP32 (HTTP fallback)
	pendingCommandsMutex.Lock()
	pendingCommands[req.DeviceCode] = cmdObj
	pendingCommandsMutex.Unlock()

	// Kirim langsung via MQTT jika terhubung
	if services.GlobalMQTT != nil {
		_ = services.GlobalMQTT.PublishCommand(req.DeviceCode, req.Action, map[string]interface{}{
			"cow_code": req.CowCode,
			"cow_name": req.CowName,
			"cow_id":   req.CowID,
		})
	}

	// Broadcast via WebSocket juga
	services.GlobalWSHub.Broadcast(fiber.Map{
		"type":        "REMOTE_COMMAND",
		"device_code": req.DeviceCode,
		"action":      req.Action,
		"cow_code":    req.CowCode,
		"cow_name":    req.CowName,
		"cow_id":      req.CowID,
	})

	return utils.Success(c, "Perintah remote '"+req.Action+"' berhasil dikirim ke perangkat")
}

// GetPendingCommand GET /api/devices/command/:device_code — PUBLIC, dipoll ESP32
func (h *DeviceHandler) GetPendingCommand(c *fiber.Ctx) error {
	deviceCode := c.Params("device_code")
	if deviceCode == "" {
		return utils.BadRequest(c, "device_code wajib diisi")
	}

	// Jika perangkat belum approved, batalkan perintah
	if !h.deviceService.IsDeviceApproved(deviceCode) {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"action":  "none",
		})
	}

	pendingCommandsMutex.Lock()
	cmd, exists := pendingCommands[deviceCode]
	if exists {
		delete(pendingCommands, deviceCode) // Konsumsi perintah 1x
	}
	pendingCommandsMutex.Unlock()

	if !exists || cmd.Action == "" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"action":  "none",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":  true,
		"action":   cmd.Action,
		"cow_code": cmd.CowCode,
		"cow_name": cmd.CowName,
		"cow_id":   cmd.CowID,
	})
}

// PostLiveWeight POST /api/devices/live — PUBLIC, dipanggil ESP32 untuk streaming berat live + telemetri lengkap
func (h *DeviceHandler) PostLiveWeight(c *fiber.Ctx) error {
	var req struct {
		DeviceCode          string  `json:"device_code"`
		Weight              float64 `json:"weight"`
		CowCode             string  `json:"cow_code"`
		CowName             string  `json:"cow_name"`
		IsLocked            bool    `json:"is_locked"`
		VccVoltage          float64 `json:"vcc_voltage"`
		VoltageStatus       string  `json:"voltage_status"`
		Hx711Status         string  `json:"hx711_status"`
		A12EStatus          string  `json:"a12e_status"`
		RawAdc              int64   `json:"raw_adc"`
		WifiRssi            int     `json:"wifi_rssi"`
		FreeHeap            int64   `json:"free_heap"`
		CpuTemp             float64 `json:"cpu_temp"`
		UptimeSec           int64   `json:"uptime_sec"`

		Hx711NoiseSigma    int64   `json:"hx711_noise_sigma"`
		Hx711NoiseStatus   string  `json:"hx711_noise_status"`
		Hx711SampleRateSps int     `json:"hx711_sample_rate_sps"`
		Hx711ZeroDrift     int64   `json:"hx711_zero_drift"`
		Hx711SnrDb         float64 `json:"hx711_snr_db"`
		Hx711OverallStatus string  `json:"hx711_overall_status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format data tidak valid")
	}

	// Cek apakah perangkat terdaftar dan disetujui (approved)
	if !h.deviceService.IsDeviceApproved(req.DeviceCode) {
		return utils.Forbidden(c, "Perangkat belum disetujui / pairing tidak aktif")
	}

	a12eStatus := req.A12EStatus
	if a12eStatus == "" {
		a12eStatus = req.Hx711Status
	}

	// Broadcast live scale streaming + hardware telemetry ke Web Dashboard via WebSocket
	services.GlobalWSHub.Broadcast(fiber.Map{
		"type":                  "LIVE_WEIGHT",
		"device_code":           req.DeviceCode,
		"weight":                req.Weight,
		"cow_code":              req.CowCode,
		"cow_name":              req.CowName,
		"is_locked":             req.IsLocked,
		"vcc_voltage":           req.VccVoltage,
		"voltage_status":        req.VoltageStatus,
		"a12e_status":           a12eStatus,
		"hx711_status":          a12eStatus,
		"raw_adc":               req.RawAdc,
		"wifi_rssi":             req.WifiRssi,
		"free_heap":             req.FreeHeap,
		"cpu_temp":              req.CpuTemp,
		"uptime_sec":            req.UptimeSec,
		// Hardware Detail
		"hx711_noise_sigma":     req.Hx711NoiseSigma,
		"hx711_noise_status":    req.Hx711NoiseStatus,
		"hx711_sample_rate_sps": req.Hx711SampleRateSps,
		"hx711_zero_drift":      req.Hx711ZeroDrift,
		"hx711_snr_db":          req.Hx711SnrDb,
		"hx711_overall_status":  req.Hx711OverallStatus,
		"timestamp":             time.Now().Unix(),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}


