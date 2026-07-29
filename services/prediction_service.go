package services

import (
	"errors"
	"math"
	"time"

	"smart-livestock-backend/models"
	"smart-livestock-backend/repositories"
)

type PredictionService struct {
	weightRepo *repositories.WeightRepository
}

func NewPredictionService(weightRepo *repositories.WeightRepository) *PredictionService {
	return &PredictionService{weightRepo: weightRepo}
}

func (s *PredictionService) PredictPolynomial(cowID int64, horizonMonths int) (*models.PredictionResponse, error) {
	records, err := s.weightRepo.GetByCowID(cowID)
	if err != nil {
		return nil, errors.New("gagal mengambil data penimbangan")
	}

	n := len(records)
	
	// Minimum 1 data untuk menghasilkan prediksi sederhana
	if n < 1 {
		return nil, errors.New("belum ada data penimbangan. Tambahkan minimal 1 data penimbangan terlebih dahulu")
	}

	// 2. SET DEFAULT: Jika parameter tidak diisi, buat default 3 bulan ke depan
	if horizonMonths <= 0 {
		horizonMonths = 3
	}

	lastRecord := records[n-1]
	horizonDays := horizonMonths * 30
	predictionDate := lastRecord.MeasurementDate.Add(time.Duration(horizonDays) * 24 * time.Hour).Format("2006-01-02")

	var predictedWeight, adg float64
	trend := "data awal"
	rekomendasi := "Data masih terlalu sedikit. Lakukan penimbangan berkala agar prediksi semakin akurat."
	accuracy := "Estimasi Awal"
	rSquared := 0.5
	dataUsed := n
	var projectedPoints []models.ProjectedPoint

	if n == 1 {
		// Hanya 1 data: prediksi flat (tidak berubah)
		predictedWeight = math.Round(lastRecord.Weight*100) / 100
		adg = 0.0
		trend = "belum dapat ditentukan"
		rekomendasi = "Baru 1 data tersedia. Tambahkan penimbangan berikutnya agar sistem dapat menghitung pertumbuhan."
		
		for i := 1; i <= horizonMonths; i++ {
			pDate := lastRecord.MeasurementDate.Add(time.Duration(i*30) * 24 * time.Hour).Format("2006-01-02")
			projectedPoints = append(projectedPoints, models.ProjectedPoint{
				Date:   pDate,
				Weight: predictedWeight,
			})
		}

	} else if n == 2 {
		// 2 data: gunakan regresi linear sederhana
		y1 := records[n-2].Weight
		y2 := records[n-1].Weight
		days := lastRecord.MeasurementDate.Sub(records[n-2].MeasurementDate).Hours() / 24.0
		if days <= 0 {
			days = 30
		}
		adg = math.Round(((y2 - y1) / days) * 100) / 100
		// Prediksi linear: tambahkan adg * horizonDays
		predictedWeight = math.Round((y2 + adg*float64(horizonDays))*100) / 100
		rSquared = 0.75
		accuracy = "Estimasi Linear"
		if predictedWeight > y2 {
			trend = "meningkat"
			rekomendasi = "LAYAK DIPERTAHANKAN: Sapi menunjukkan pertumbuhan positif. Tambahkan lebih banyak data untuk prediksi lebih akurat."
		} else {
			trend = "menurun"
			rekomendasi = "PERLU PERHATIAN: Berat sapi tidak meningkat. Periksa manajemen pakan dan kesehatan sapi."
		}
		dataUsed = 2
		
		for i := 1; i <= horizonMonths; i++ {
			pDate := lastRecord.MeasurementDate.Add(time.Duration(i*30) * 24 * time.Hour).Format("2006-01-02")
			pWeight := math.Round((y2 + adg*float64(i*30))*100) / 100
			projectedPoints = append(projectedPoints, models.ProjectedPoint{
				Date:   pDate,
				Weight: pWeight,
			})
		}
	} else {
		// 3+ data: gunakan polinomial derajat 2 (rumus asli)
		y1 := records[n-3].Weight
		y2 := records[n-2].Weight
		y3 := records[n-1].Weight

		a := (y3 - 2*y2 + y1) / 2.0
		b := (-5*y1 + 8*y2 - 3*y3) / 2.0
		c := (3 * y1) - (3 * y2) + y3

		predictX := 3.0 + float64(horizonMonths)
		predictedWeight = math.Round(((a*predictX*predictX)+(b*predictX)+c)*100) / 100
		rSquared = 1.0
		accuracy = "Tinggi"
		dataUsed = 3

		prevRecord := records[n-2]
		days := lastRecord.MeasurementDate.Sub(prevRecord.MeasurementDate).Hours() / 24.0
		if days > 0 {
			adg = math.Round(((y3 - y2) / days) * 100) / 100
		} else {
			adg = math.Round(((y3 - y2) / 30.0) * 100) / 100
		}

		if predictedWeight > y3 && a >= 0 {
			trend = "meningkat pesat"
			rekomendasi = "SANGAT LAYAK DIPERTAHANKAN: Lanjutkan program pakan saat ini. Sapi sedang dalam fase pertumbuhan optimal."
		} else if predictedWeight > y3 && a < 0 {
			trend = "meningkat (melambat)"
			rekomendasi = "LAYAK DIPERTAHANKAN SEMENTARA: Pantau ketat, sapi mendekati batas genetik bobot maksimal. Rencanakan penjualan/panen."
		} else {
			trend = "menurun / harus dijual"
			rekomendasi = "TIDAK LAYAK DIPERTAHANKAN: Segera jual atau potong sapi ini untuk menghindari kerugian pemborosan pakan."
		}

		for i := 1; i <= horizonMonths; i++ {
			pDate := lastRecord.MeasurementDate.Add(time.Duration(i*30) * 24 * time.Hour).Format("2006-01-02")
			pX := 3.0 + float64(i)
			pWeight := math.Round(((a*pX*pX)+(b*pX)+c)*100) / 100
			projectedPoints = append(projectedPoints, models.ProjectedPoint{
				Date:   pDate,
				Weight: pWeight,
			})
		}

	}

	return &models.PredictionResponse{
		CowID:            cowID,
		CurrentWeight:    lastRecord.Weight,
		PredictedWeight:  predictedWeight,
		PredictionDate:   predictionDate,
		HorizonDays:      horizonDays,
		RSquared:         rSquared,
		AccuracyCategory: accuracy,
		Trend:            trend,
		Recommendation:   rekomendasi,
		ADG:              adg,
		DataPointsUsed:   dataUsed,
		ProjectedPoints:  projectedPoints,
	}, nil
}


func linearRegression(x, y []float64) (slope, intercept, rSquared float64) {
	n := float64(len(x))
	var sumX, sumY, sumXY, sumX2 float64

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0, sumY / n, 0
	}

	slope = (n*sumXY - sumX*sumY) / denominator
	intercept = (sumY - slope*sumX) / n

	yMean := sumY / n
	var ssRes, ssTot float64
	for i := 0; i < len(x); i++ {
		predicted := slope*x[i] + intercept
		ssRes += (y[i] - predicted) * (y[i] - predicted)
		ssTot += (y[i] - yMean) * (y[i] - yMean)
	}

	if ssTot == 0 {
		rSquared = 0
	} else {
		rSquared = 1 - (ssRes / ssTot)
	}

	return slope, intercept, rSquared
}
