package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"smart-livestock-backend/models"
	"smart-livestock-backend/repositories"
	"smart-livestock-backend/utils"
)

type WeightService struct {
	weightRepo *repositories.WeightRepository
	cowRepo    *repositories.CowRepository
}

func NewWeightService(weightRepo *repositories.WeightRepository, cowRepo *repositories.CowRepository) *WeightService {
	return &WeightService{weightRepo: weightRepo, cowRepo: cowRepo}
}

func (s *WeightService) AddWeight(req models.WeightRequest, operatorID *int64) (*models.WeightRecord, error) {
	if err := utils.ValidateWeight(req.Weight); err != nil {
		return nil, err
	}

	var cow *models.Cow
	var err error
	if req.CowID > 0 {
		cow, err = s.cowRepo.FindByID(req.CowID)
	}
	if (err != nil || cow == nil) && req.CowCode != "" {
		cow, err = s.cowRepo.FindByCode(req.CowCode)
		if cow != nil {
			req.CowID = cow.ID
		}
	}

	if err != nil || cow == nil {
		// Jika sapi belum ada di DB (misal karena baru saja dikosongkan),
		// buat otomatis sapi default agar penimbangan dari ESP32 langsung tersimpan
		code := req.CowCode
		if code == "" {
			code = "SP-001"
		}
		// Cek dulu apakah cow_code sudah ada sebelum insert
		existingCow, _ := s.cowRepo.FindByCode(code)
		if existingCow != nil {
			req.CowID = existingCow.ID
		} else {
			name := "Sapi " + code
			breed := "Sapi Ternak"

			newID, errCreate := s.cowRepo.Create(models.CowRequest{
				CowCode: code,
				Name:    name,
				Breed:   breed,
				Gender:  "jantan",
			})
			if errCreate == nil {
				req.CowID = newID
			} else {
				return nil, fmt.Errorf("gagal membuat sapi default: %v", errCreate)
			}
		}
	}

	loc, errLoc := time.LoadLocation("Asia/Jakarta")
	now := time.Now()
	if errLoc == nil {
		now = now.In(loc)
	}
	// Simpan sebagai RFC3339 dengan offset +07:00 agar browser bisa parsing timezone dengan benar
	measureDate := now.Format(time.RFC3339)
	if req.Date != "" {
		if err := utils.ValidateDateNotFuture(req.Date); err != nil {
			return nil, err
		}
		measureDate = req.Date
	}

	var adg *float64
	lastRecord, err := s.weightRepo.GetLastByCowID(req.CowID)
	if err == nil && lastRecord != nil {
		days := daysBetween(lastRecord.MeasurementDate, parseDate(measureDate))
		if days > 0 {
			adgValue := (req.Weight - lastRecord.Weight) / float64(days)
			adgValue = math.Round(adgValue*100) / 100
			adg = &adgValue
		}
	}

	id, err := s.weightRepo.Create(req.CowID, req.Weight, adg, measureDate, req.DeviceID, operatorID)
	if err != nil {
		return nil, errors.New("gagal menyimpan data penimbangan")
	}

	record := &models.WeightRecord{
		ID:              id,
		CowID:           req.CowID,
		Weight:          req.Weight,
		ADG:             adg,
		MeasurementDate: parseDate(measureDate),
	}

	// Fetch detail cow untuk broadcast WebSocket
	cowCode := ""
	cowName := ""
	if cowInfo, err := s.cowRepo.FindByID(req.CowID); err == nil && cowInfo != nil {
		cowCode = cowInfo.CowCode
		cowName = cowInfo.Name
	}

	// Broadcast event secara Instant (<50ms) ke seluruh Browser Web yang terhubung
	GlobalWSHub.Broadcast(map[string]interface{}{
		"event": "NEW_WEIGHING",
		"data": map[string]interface{}{
			"id":               id,
			"cow_id":           req.CowID,
			"cow_code":         cowCode,
			"cow_name":         cowName,
			"weight":           req.Weight,
			"adg":              adg,
			"measurement_date": parseDate(measureDate).Format(time.RFC3339),
			"device_id":        req.DeviceID,
		},
	})

	return record, nil
}

func (s *WeightService) GetHistory(cowID int64, startDate, endDate string) ([]models.WeightWithCow, error) {
	return s.weightRepo.GetAll(cowID, startDate, endDate)
}

func (s *WeightService) GetCowWeights(cowID int64) ([]models.WeightRecord, error) {
	return s.weightRepo.GetByCowID(cowID)
}

// AddBatch — proses batch upload dari ESP32 offline queue
// POST /api/weighings/batch
func (s *WeightService) AddBatch(req models.BatchWeighingRequest) models.BatchWeighingResponse {
	resp := models.BatchWeighingResponse{Total: len(req.Records)}

	loc, _ := time.LoadLocation("Asia/Jakarta")

	for _, rec := range req.Records {
		// Resolve cow_id: dari cow_code jika cow_id tidak dikirim
		cowID := rec.CowID
		if cowID == 0 && rec.CowCode != "" {
			cow, err := s.cowRepo.FindByCode(rec.CowCode)
			if err != nil || cow == nil {
				resp.Skipped++
				resp.Errors = append(resp.Errors, "cow_code tidak ditemukan: "+rec.CowCode)
				continue
			}
			cowID = cow.ID
		}
		if cowID == 0 {
			resp.Skipped++
			resp.Errors = append(resp.Errors, "cow_id atau cow_code wajib diisi")
			continue
		}

		// Waktu pengukuran
		measureDate := time.Now().In(loc).Format(time.RFC3339)
		if rec.Date != "" {
			measureDate = rec.Date
		}

		// Hitung ADG
		var adg *float64
		if lastRecord, err := s.weightRepo.GetLastByCowID(cowID); err == nil && lastRecord != nil {
			days := daysBetween(lastRecord.MeasurementDate, parseDate(measureDate))
			if days > 0 {
				adgVal := math.Round(((rec.Weight - lastRecord.Weight) / days) * 100) / 100
				adg = &adgVal
			}
		}

		_, err := s.weightRepo.Create(cowID, rec.Weight, adg, measureDate, req.DeviceCode, nil)
		if err != nil {
			resp.Skipped++
			resp.Errors = append(resp.Errors, "gagal simpan untuk cow_id "+string(rune(cowID)))
			continue
		}
		resp.Saved++
	}

	// Broadcast WebSocket setelah batch selesai
	GlobalWSHub.Broadcast(map[string]interface{}{
		"event": "BATCH_SYNC_COMPLETE",
		"data": map[string]interface{}{
			"device_code": req.DeviceCode,
			"saved":       resp.Saved,
			"total":       resp.Total,
		},
	})

	return resp
}


func daysBetween(a, b time.Time) float64 {
	duration := b.Sub(a)
	return duration.Hours() / 24
}

func parseDate(dateStr string) time.Time {
	loc, errLoc := time.LoadLocation("Asia/Jakarta")
	if errLoc != nil {
		loc = time.Local
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, loc)
	if err == nil {
		return t
	}
	t, err = time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t.In(loc)
	}
	t, err = time.ParseInLocation("2006-01-02", dateStr, loc)
	if err == nil {
		return t
	}
	return time.Now().In(loc)
}
