package services

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"smart-livestock-backend/models"
	"smart-livestock-backend/repositories"
	"smart-livestock-backend/utils"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.User, error) {
	existing, _ := s.userRepo.FindUsername(req.Username)
	if existing != nil {
		return nil, errors.New("Username sudah digunakan")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}

	id, err := s.userRepo.Create(req.Username, req.Email, string(hashedPassword), "peternak")
	if err != nil {
		return nil, errors.New("gagal menyimpan user")
	}
	user := &models.User{
		ID:       id,
		Username: req.Username,
		Role:     "peternak",
	}
	return user, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.FindUsername(req.Username)
	if err != nil {
		return nil, errors.New("username atau password salah")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("username atau password salah")
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil

}
