package models

import "time"

type Device struct {
	ID            int64     `db:"id" json:"id"`
	DeviceCode    string    `db:"device_code" json:"device_code"`
	DeviceName    string    `db:"device_name" json:"device_name"`
	UserID        *int64    `db:"user_id" json:"user_id,omitempty"`
	Status        string    `db:"status" json:"status"`
	PairingStatus string    `db:"pairing_status" json:"pairing_status"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type DeviceRequest struct {
	DeviceCode string `json:"device_code"`
	DeviceName string `json:"device_name"`
}

type ClaimDeviceRequest struct {
	DeviceCode string `json:"device_code"`
	DeviceName string `json:"device_name"`
}

// PairingRequest digunakan ESP32 untuk meminta pairing ke server
type PairingRequest struct {
	DeviceCode string `json:"device_code"`
	DeviceName string `json:"device_name"`
}

// ApproveDeviceRequest digunakan web admin untuk approve/reject pairing
type ApproveDeviceRequest struct {
	PairingStatus string `json:"pairing_status"` // "approved" or "rejected"
}
