package repositories

import (
	"github.com/jmoiron/sqlx"
	"smart-livestock-backend/models"
)

type CowRepository struct {
	db *sqlx.DB
}

func NewCowRepository(db *sqlx.DB) *CowRepository {
	return &CowRepository{db: db}
}

func (r *CowRepository) GetAll(status, breed string, page, limit int) ([]models.CowListItem, int, error) {
	countQuery := "SELECT COUNT(*) FROM cows WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		countQuery += " AND status = ?"
		args = append(args, status)
	}
	if breed != "" {
		countQuery += " AND breed LIKE ?"
		args = append(args, "%"+breed+"%")
	}
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT c.*,
			(SELECT weight FROM weight_records WHERE cow_id = c.id ORDER BY measurement_date DESC LIMIT 1) as last_weight,
			(SELECT adg FROM weight_records WHERE cow_id = c.id ORDER BY measurement_date DESC LIMIT 1) as last_adg,
			(SELECT COUNT(*) FROM weight_records WHERE cow_id = c.id) as weigh_count
		FROM cows c
		WHERE 1=1
	`
	if status != "" {
		query += " AND c.status = ?"
	}
	if breed != "" {
		query += " AND c.breed LIKE ?"
	}
	query += " ORDER BY c.created_at DESC LIMIT ? OFFSET ?"
	offset := (page - 1) * limit
	args = append(args, limit, offset)
	var cows []models.CowListItem
	err = r.db.Select(&cows, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return cows, total, nil
}

func (r *CowRepository) FindByID(id int64) (*models.Cow, error) {
	var cow models.Cow
	err := r.db.Get(&cow, "SELECT * FROM cows WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &cow, nil
}

func (r *CowRepository) FindByCode(code string) (*models.Cow, error) {
	var cow models.Cow
	err := r.db.Get(&cow, "SELECT * FROM cows WHERE cow_code = ?", code)
	if err != nil {
		return nil, err
	}
	return &cow, nil
}

func (r *CowRepository) Create(req models.CowRequest) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO cows (cow_code, name, breed, gender, birth_date, owner) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		req.CowCode, req.Name, req.Breed, req.Gender, req.BirthDate, req.Owner,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *CowRepository) Update(id int64, req models.CowRequest) error {
	_, err := r.db.Exec(
		`UPDATE cows SET name = ?, breed = ?, gender = ?, birth_date = ?, owner = ?, 
		 status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		req.Name, req.Breed, req.Gender, req.BirthDate, req.Owner, req.Status, id,
	)
	return err
}

func (r *CowRepository) SoftDelete(id int64) error {
	_, err := r.db.Exec(
		"UPDATE cows SET status = 'deceased', updated_at = CURRENT_TIMESTAMP WHERE id = ?", id,
	)
	return err
}

func getStandardCowBreeds() []models.CowCacheItem {
	return []models.CowCacheItem{
		{ID: 1,  CowCode: "BALI-01",  Name: "Sapi Bali",       Breed: "Sapi Bali"},
		{ID: 2,  CowCode: "MAD-01",   Name: "Sapi Madura",     Breed: "Sapi Madura"},
		{ID: 3,  CowCode: "PO-01",    Name: "Sapi PO",         Breed: "Peranakan Ongole"},
		{ID: 4,  CowCode: "LIM-01",   Name: "Sapi Limosin",    Breed: "Sapi Limosin"},
		{ID: 5,  CowCode: "SIM-01",   Name: "Sapi Simental",   Breed: "Sapi Simental"},
		{ID: 6,  CowCode: "FH-01",    Name: "Sapi FH",         Breed: "Friesian Holstein"},
		{ID: 7,  CowCode: "BRH-01",   Name: "Sapi Brahman",    Breed: "Sapi Brahman"},
		{ID: 8,  CowCode: "ACEH-01",  Name: "Sapi Aceh",       Breed: "Sapi Aceh"},
		{ID: 9,  CowCode: "PES-01",   Name: "Sapi Pesisir",    Breed: "Sapi Pesisir"},
		{ID: 10, CowCode: "BRG-01",   Name: "Sapi Brangus",    Breed: "Sapi Brangus"},
		{ID: 11, CowCode: "WAG-01",   Name: "Sapi Wagyu",      Breed: "Sapi Wagyu"},
		{ID: 12, CowCode: "ANG-01",   Name: "Sapi Angus",      Breed: "Black Angus"},
		{ID: 13, CowCode: "HER-01",   Name: "Sapi Hereford",   Breed: "Sapi Hereford"},
		{ID: 14, CowCode: "CHA-01",   Name: "Sapi Charolais",  Breed: "Sapi Charolais"},
		{ID: 15, CowCode: "JAB-01",   Name: "Sapi Jabres",     Breed: "Sapi Jabres"},
	}
}

// GetCowsForDevice — ambil daftar sapi ringkas milik pemilik device_code
// Digunakan untuk sync cache ke ESP32
func (r *CowRepository) GetCowsForDevice(deviceCode string) ([]models.CowCacheItem, error) {
	var pairingStatus string
	_ = r.db.Get(&pairingStatus, "SELECT pairing_status FROM devices WHERE device_code = ?", deviceCode)

	// Jika perangkat belum terdaftar/approved, kembalikan 15 Rumpun Sapi Standar
	if pairingStatus != "approved" {
		return getStandardCowBreeds(), nil
	}

	query := `
		SELECT id, cow_code, name, breed
		FROM cows
		WHERE status = 'active'
		ORDER BY cow_code ASC
	`
	var cows []models.CowCacheItem
	_ = r.db.Select(&cows, query)

	if len(cows) == 0 {
		return getStandardCowBreeds(), nil
	}

	return cows, nil
}