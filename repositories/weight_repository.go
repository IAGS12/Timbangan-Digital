package repositories

import (
	"github.com/jmoiron/sqlx"
	"smart-livestock-backend/models"
)

type WeightRepository struct {
	db *sqlx.DB
}

func NewWeightRepository(db *sqlx.DB) *WeightRepository {
	return &WeightRepository{db: db}
}


func (r *WeightRepository) Create(cowID int64, weight float64, adg *float64, date, deviceID string, operatorID *int64) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO weight_records (cow_id, weight, adg, measurement_date, device_id, operator_id) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		cowID, weight, adg, date, deviceID, operatorID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}


func (r *WeightRepository) GetByCowID(cowID int64) ([]models.WeightRecord, error) {
	var records []models.WeightRecord
	err := r.db.Select(&records,
		"SELECT * FROM weight_records WHERE cow_id = ? ORDER BY measurement_date ASC",
		cowID,
	)
	if err != nil {
		return nil, err
	}
	return records, nil
}


func (r *WeightRepository) GetLastByCowID(cowID int64) (*models.WeightRecord, error) {
	var record models.WeightRecord
	err := r.db.Get(&record,
		"SELECT * FROM weight_records WHERE cow_id = ? ORDER BY measurement_date DESC LIMIT 1",
		cowID,
	)
	if err != nil {
		return nil, err
	}
	return &record, nil
}


func (r *WeightRepository) GetAll(cowID int64, startDate, endDate string) ([]models.WeightWithCow, error) {
	query := `
		SELECT wr.*, c.cow_code, c.name as cow_name
		FROM weight_records wr
		JOIN cows c ON wr.cow_id = c.id
		WHERE 1=1
	`
	args := []interface{}{}

	if cowID > 0 {
		query += " AND wr.cow_id = ?"
		args = append(args, cowID)
	}
	if startDate != "" {
		query += " AND wr.measurement_date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND wr.measurement_date <= ?"
		args = append(args, endDate)
	}
	query += " ORDER BY wr.measurement_date DESC"

	var records []models.WeightWithCow
	err := r.db.Select(&records, query, args...)
	if err != nil {
		return nil, err
	}
	return records, nil
}


func (r *WeightRepository) CountAll() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM weight_records")
	return count, err
}


func (r *WeightRepository) GetAverageWeight() (float64, error) {
	var avg float64
	err := r.db.Get(&avg, `
		SELECT COALESCE(AVG(wr.weight), 0)
		FROM weight_records wr
		INNER JOIN (
			SELECT cow_id, MAX(measurement_date) as max_date
			FROM weight_records
			GROUP BY cow_id
		) latest ON wr.cow_id = latest.cow_id AND wr.measurement_date = latest.max_date
		JOIN cows c ON wr.cow_id = c.id
		WHERE c.status = 'active'
	`)
	return avg, err
}

type MonthlyGrowth struct {
	MonthGroup string  `db:"month_group"`
	AvgWeight  float64 `db:"avg_weight"`
	AvgADG     float64 `db:"avg_adg"`
}

func (r *WeightRepository) GetGrowthTrend(limitMonths int) ([]MonthlyGrowth, error) {
	query := `
		SELECT 
			strftime('%Y-%m', measurement_date) as month_group,
			COALESCE(AVG(weight), 0) as avg_weight,
			COALESCE(AVG(adg), 0) as avg_adg
		FROM weight_records
		GROUP BY month_group
		ORDER BY month_group DESC
		LIMIT ?
	`
	var results []MonthlyGrowth
	err := r.db.Select(&results, query, limitMonths)
	if err != nil {
		return nil, err
	}
	return results, nil
}

