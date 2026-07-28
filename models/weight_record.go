package models

import "time"

type WeightRecord struct {
	ID              int64     `db:"id" json:"id"`
	CowID           int64     `db:"cow_id" json:"cow_id"`
	Weight          float64   `db:"weight" json:"weight"`
	ADG             *float64  `db:"adg" json:"adg,omitempty"`
	MeasurementDate time.Time `db:"measurement_date" json:"measurement_date"`
	DeviceID        *string   `db:"device_id" json:"device_id,omitempty"`
	OperatorID      *int64    `db:"operator_id" json:"operator_id,omitempty"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}


type WeightRequest struct {
	CowID    int64   `json:"cow_id"`
	CowCode  string  `json:"cow_code,omitempty"`
	Weight   float64 `json:"weight"`
	Date     string  `json:"date,omitempty"`
	DeviceID string  `json:"device_id,omitempty"`
}


type WeightWithCow struct {
	WeightRecord
	CowCode string `db:"cow_code" json:"cow_code"`
	CowName string `db:"cow_name" json:"cow_name"`
}

// BatchWeighingRecord — satu record dalam batch upload dari ESP32 (offline queue)
type BatchWeighingRecord struct {
	CowCode  string  `json:"cow_code"`   // kode sapi dari cache lokal ESP32
	CowID    int64   `json:"cow_id"`     // ID sapi (opsional, jika ESP32 sudah simpan)
	Weight   float64 `json:"weight"`
	Date     string  `json:"date,omitempty"` // RFC3339, opsional (gunakan server time jika kosong)
}

// BatchWeighingRequest — payload POST /api/weighings/batch dari ESP32
type BatchWeighingRequest struct {
	DeviceCode string                `json:"device_code"`
	Records    []BatchWeighingRecord `json:"records"`
}

// BatchWeighingResponse — ringkasan hasil batch upload
type BatchWeighingResponse struct {
	Total     int      `json:"total"`
	Saved     int      `json:"saved"`
	Skipped   int      `json:"skipped"`
	Errors    []string `json:"errors,omitempty"`
}

