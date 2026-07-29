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
	err := r.db.Get(&user, "SELECT * FROM users WHERE LOWER(username) = LOWER(?)", username)
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

func (r *UserRepository) FindByID(id int64) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUsername(id int64, username string) error {
	_, err := r.db.Exec("UPDATE users SET username = ? WHERE id = ?", username, id)
	return err
}

func (r *UserRepository) UpdatePassword(id int64, hashedPassword string) error {
	_, err := r.db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, id)
	return err
}

func (r *UserRepository) UsernameExists(username string, excludeID int64) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM users WHERE LOWER(username) = LOWER(?) AND id != ?", username, excludeID)
	return count > 0, err
}