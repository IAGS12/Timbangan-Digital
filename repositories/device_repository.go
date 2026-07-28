package repositories

import (
	"github.com/jmoiron/sqlx"
	"smart-livestock-backend/models"
)

type DeviceRepository struct {
	db *sqlx.DB
}

func NewDeviceRepository(db *sqlx.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) GetByUserID(userID int64) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Select(&devices, "SELECT * FROM devices WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// GetAllPending returns all devices with pairing_status = 'pending'
func (r *DeviceRepository) GetAllPending() ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Select(&devices, "SELECT * FROM devices WHERE pairing_status = 'pending' ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// GetAllDevices returns all devices (for admin view)
func (r *DeviceRepository) GetAllDevices() ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Select(&devices, "SELECT * FROM devices ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *DeviceRepository) FindByCode(code string) (*models.Device, error) {
	var device models.Device
	err := r.db.Get(&device, "SELECT * FROM devices WHERE device_code = ?", code)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepository) FindByID(id int64) (*models.Device, error) {
	var device models.Device
	err := r.db.Get(&device, "SELECT * FROM devices WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// RequestPairing dipanggil ESP32 — insert jika baru, update ke pending jika sudah ada
func (r *DeviceRepository) RequestPairing(code, name string) error {
	var exists int
	_ = r.db.Get(&exists, "SELECT COUNT(*) FROM devices WHERE device_code = ?", code)
	if exists > 0 {
		_, err := r.db.Exec(
			"UPDATE devices SET device_name = ?, pairing_status = 'pending', status = 'active' WHERE device_code = ?",
			name, code,
		)
		return err
	}
	_, err := r.db.Exec(
		"INSERT INTO devices (device_code, device_name, status, pairing_status) VALUES (?, ?, 'active', 'pending')",
		code, name,
	)
	return err
}

// UpdatePairingStatus approve / reject oleh admin web
func (r *DeviceRepository) UpdatePairingStatus(id int64, pairingStatus string) error {
	_, err := r.db.Exec(
		"UPDATE devices SET pairing_status = ? WHERE id = ?",
		pairingStatus, id,
	)
	return err
}

// GetPairingStatus query status pairing untuk ESP32 polling
func (r *DeviceRepository) GetPairingStatus(code string) (string, error) {
	var status string
	err := r.db.Get(&status, "SELECT pairing_status FROM devices WHERE device_code = ?", code)
	return status, err
}

func (r *DeviceRepository) ClaimDevice(code string, name string, userID int64) error {
	var exists int
	err := r.db.Get(&exists, "SELECT COUNT(*) FROM devices WHERE device_code = ?", code)
	if err == nil && exists > 0 {
		_, err = r.db.Exec("UPDATE devices SET user_id = ?, device_name = ?, status = 'active' WHERE device_code = ?", userID, name, code)
		return err
	}
	_, err = r.db.Exec("INSERT INTO devices (device_code, device_name, user_id, status, pairing_status) VALUES (?, ?, ?, 'active', 'approved')", code, name, userID)
	return err
}

func (r *DeviceRepository) UnlinkDevice(id int64, userID int64) error {
	_, err := r.db.Exec("DELETE FROM devices WHERE id = ?", id)
	return err
}

func (r *DeviceRepository) DeleteDevice(id int64) error {
	_, err := r.db.Exec("DELETE FROM devices WHERE id = ?", id)
	return err
}
