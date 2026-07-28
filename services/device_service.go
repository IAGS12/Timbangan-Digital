package services

import (
	"errors"
	"smart-livestock-backend/models"
	"smart-livestock-backend/repositories"
)

type DeviceService struct {
	deviceRepo *repositories.DeviceRepository
}

func NewDeviceService(deviceRepo *repositories.DeviceRepository) *DeviceService {
	return &DeviceService{deviceRepo: deviceRepo}
}

func (s *DeviceService) GetUserDevices(userID int64) ([]models.Device, error) {
	return s.deviceRepo.GetByUserID(userID)
}

func (s *DeviceService) GetAllDevices() ([]models.Device, error) {
	return s.deviceRepo.GetAllDevices()
}

func (s *DeviceService) GetPendingDevices() ([]models.Device, error) {
	return s.deviceRepo.GetAllPending()
}

// RequestPairing dipanggil ESP32 untuk mendaftarkan diri ke server
func (s *DeviceService) RequestPairing(req models.PairingRequest) error {
	if req.DeviceCode == "" {
		return errors.New("device_code wajib diisi")
	}
	name := req.DeviceName
	if name == "" {
		name = "Timbangan " + req.DeviceCode
	}
	return s.deviceRepo.RequestPairing(req.DeviceCode, name)
}

// GetPairingStatus digunakan ESP32 untuk polling status pairing
func (s *DeviceService) GetPairingStatus(deviceCode string) (string, error) {
	status, err := s.deviceRepo.GetPairingStatus(deviceCode)
	if err != nil {
		return "unknown", err
	}
	return status, nil
}

// ApprovePairing digunakan admin web untuk approve / reject perangkat
func (s *DeviceService) ApprovePairing(deviceID int64, pairingStatus string) error {
	if pairingStatus != "approved" && pairingStatus != "rejected" && pairingStatus != "unpaired" {
		return errors.New("status harus 'approved', 'rejected', atau 'unpaired'")
	}
	return s.deviceRepo.UpdatePairingStatus(deviceID, pairingStatus)
}

// IsDeviceApproved cek apakah device_code sudah approved (dipakai di weight handler)
func (s *DeviceService) IsDeviceApproved(deviceCode string) bool {
	status, err := s.deviceRepo.GetPairingStatus(deviceCode)
	if err != nil {
		return false
	}
	return status == "approved"
}

func (s *DeviceService) ClaimDevice(req models.ClaimDeviceRequest, userID int64) error {
	if req.DeviceCode == "" {
		return errors.New("kode perangkat wajib diisi")
	}
	name := req.DeviceName
	if name == "" {
		name = "Timbangan " + req.DeviceCode
	}

	// Validasi Maksimal 1 Device per User
	userDevices, err := s.deviceRepo.GetByUserID(userID)
	if err == nil && len(userDevices) >= 1 {
		// Pastikan bukan perangkat yang sama yang sedang di-claim ulang
		if userDevices[0].DeviceCode != req.DeviceCode {
			return errors.New("batas maksimal tercapai: Anda hanya dapat menghubungkan 1 perangkat timbangan. Hapus perangkat lama terlebih dahulu")
		}
	}

	return s.deviceRepo.ClaimDevice(req.DeviceCode, name, userID)
}

func (s *DeviceService) UnlinkDevice(id int64, userID int64) error {
	return s.deviceRepo.DeleteDevice(id)
}

func (s *DeviceService) DeleteDevice(id int64) error {
	return s.deviceRepo.DeleteDevice(id)
}
