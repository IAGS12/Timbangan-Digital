package services

import (
	"errors"

	"smart-livestock-backend/models"
	"smart-livestock-backend/repositories"
	"smart-livestock-backend/utils"
)

type CowService struct {
	cowRepo *repositories.CowRepository
}

func NewCowService(cowRepo *repositories.CowRepository) *CowService {
	return &CowService{cowRepo: cowRepo}
}

func (s *CowService) GetAll(status, breed string, page, limit int) ([]models.CowListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.cowRepo.GetAll(status, breed, page, limit)
}

func (s *CowService) GetByID(id int64) (*models.Cow, error) {
	cow, err := s.cowRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("sapi tidak ditemukan")
	}
	return cow, nil
}

func (s *CowService) Create(req models.CowRequest) (int64, error) {
	if err := utils.ValidateRequired(req.CowCode, "cow_code"); err != nil {
		return 0, err
	}
	if err := utils.ValidateRequired(req.Name, "name"); err != nil {
		return 0, err
	}
	if err := utils.ValidateRequired(req.Breed, "breed"); err != nil {
		return 0, err
	}
	if err := utils.ValidateGender(req.Gender); err != nil {
		return 0, err
	}
	if err := utils.ValidateDateNotFuture(req.BirthDate); err != nil {
		return 0, err
	}

	existing, _ := s.cowRepo.FindByCode(req.CowCode)
	if existing != nil {
		return 0, errors.New("cow_code sudah digunakan")
	}

	return s.cowRepo.Create(req)
}

func (s *CowService) Update(id int64, req models.CowRequest) error {
	existing, err := s.cowRepo.FindByID(id)
	if err != nil {
		return errors.New("sapi tidak ditemukan")
	}

	if req.Gender != "" {
		if err := utils.ValidateGender(req.Gender); err != nil {
			return err
		}
	} else {
		req.Gender = existing.Gender
	}

	if req.Status != "" {
		if err := utils.ValidateStatus(req.Status); err != nil {
			return err
		}
	} else {
		req.Status = existing.Status
	}

	if req.Name == "" {
		req.Name = existing.Name
	}
	if req.Breed == "" {
		req.Breed = existing.Breed
	}
	if req.BirthDate == "" && existing.BirthDate != nil {
		req.BirthDate = *existing.BirthDate
	}
	if req.Owner == "" && existing.Owner != nil {
		req.Owner = *existing.Owner
	}

	return s.cowRepo.Update(id, req)
}

func (s *CowService) Delete(id int64) error {
	_, err := s.cowRepo.FindByID(id)
	if err != nil {
		return errors.New("sapi tidak ditemukan")
	}
	return s.cowRepo.SoftDelete(id)
}

// GetCowsForSync — daftar sapi ringkas milik pemilik perangkat, untuk cache di ESP32
func (s *CowService) GetCowsForSync(deviceCode string) ([]models.CowCacheItem, error) {
	if deviceCode == "" {
		return nil, errors.New("device_code wajib diisi")
	}
	return s.cowRepo.GetCowsForDevice(deviceCode)
}

