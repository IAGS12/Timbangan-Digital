package repositories

import (
	"github.com/jmoiron/sqlx"
	"smart-livestock-backend/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(username, email, password, role string) (int64, error) {
	result, err := r.db.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)",
		username, email, password, role)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Select(&users, "SELECT id, username, email, role, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) UpdateRole(id int64, role string) error {
	_, err := r.db.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
	return err
}