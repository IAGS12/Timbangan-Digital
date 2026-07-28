package models

import "time"

type Cow struct {
	ID        int64     `db:"id" json:"id"`
	CowCode   string    `db:"cow_code" json:"cow_code"`
	Name      string    `db:"name" json:"name"`
	Breed     string    `db:"breed" json:"breed"`
	Gender    string    `db:"gender" json:"gender"`
	BirthDate *string   `db:"birth_date" json:"birth_date,omitempty"`
	Owner     *string   `db:"owner" json:"owner,omitempty"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}


type CowRequest struct {
	CowCode   string `json:"cow_code"`
	Name      string `json:"name"`
	Breed     string `json:"breed"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date,omitempty"`
	Owner     string `json:"owner,omitempty"`
	Status    string `json:"status,omitempty"`
}


type CowListItem struct {
	Cow
	LastWeight *float64 `db:"last_weight" json:"last_weight,omitempty"`
	LastADG    *float64 `db:"last_adg" json:"last_adg,omitempty"`
	WeighCount int      `db:"weigh_count" json:"weigh_count"`
}

// CowCacheItem — data minimal dikirim ke ESP32 untuk disimpan di NVS/LittleFS
type CowCacheItem struct {
	ID      int64  `db:"id"       json:"id"`
	CowCode string `db:"cow_code" json:"cow_code"`
	Name    string `db:"name"     json:"name"`
	Breed   string `db:"breed"    json:"breed"`
}
