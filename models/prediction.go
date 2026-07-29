package models

import "time"

type Prediction struct {
	ID               int64     `db:"id" json:"id"`
	CowID            int64     `db:"cow_id" json:"cow_id"`
	PredictionWeight float64   `db:"prediction_weight" json:"prediction_weight"`
	PredictionDate   string    `db:"prediction_date" json:"prediction_date"`
	HorizonDays      int       `db:"horizon_days" json:"horizon_days"`
	Trend            string    `db:"trend" json:"trend"`
	RSquared         *float64  `db:"r_squared" json:"r_squared,omitempty"`
	Recommendation   string    `db:"recommendation" json:"recommendation"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}


type ProjectedPoint struct {
	Month  int     `json:"month"`
	Date   string  `json:"date"`
	Weight float64 `json:"weight"`
}

type PredictionResponse struct {
	CowID            int64            `json:"cow_id"`
	CurrentWeight    float64          `json:"current_weight"`
	PredictedWeight  float64          `json:"predicted_weight"`
	PredictionDate   string           `json:"prediction_date"`
	HorizonDays      int              `json:"horizon_days"`
	RSquared         float64          `json:"r_squared"`
	AccuracyCategory string           `json:"accuracy_category"`
	Trend            string           `json:"trend"`
	ADG              float64          `json:"adg"`
	Recommendation   string           `json:"recommendation"`
	DataPointsUsed   int              `json:"data_points_used"`
	Reason           string           `json:"reason"`
	ProjectedPoints  []ProjectedPoint `json:"projected_points"`
}

