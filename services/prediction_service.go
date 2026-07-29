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

// PredictPolynomial menggunakan regresi linear least-squares berbasis hari aktual
// untuk menghasilkan prediksi bobot yang akurat per bulan.
func (s *PredictionService) PredictPolynomial(cowID int64, horizonMonths int) (*models.PredictionResponse, error) {
	records, err := s.weightRepo.GetByCowID(cowID)
	if err != nil {
		return nil, errors.New("gagal mengambil data penimbangan")
	}

	n := len(records)
	if n < 1 {
		return nil, errors.New("belum ada data penimbangan. Tambahkan minimal 1 data penimbangan terlebih dahulu")
	}

	if horizonMonths <= 0 {
		horizonMonths = 3
	}

	lastRecord := records[n-1]
	firstRecord := records[0]
	horizonDays := horizonMonths * 30
	finalPredictionDate := lastRecord.MeasurementDate.Add(time.Duration(horizonDays) * 24 * time.Hour).Format("2006-01-02")

	var predictedWeight, adg, rSquared float64
	var trend, rekomendasi, accuracy string
	var dataUsed int
	var projectedPoints []models.ProjectedPoint

	if n == 1 {
		// --- KASUS 1: Hanya 1 data → proyeksi flat ---
		predictedWeight = math.Round(lastRecord.Weight*100) / 100
		adg = 0.0
		trend = "belum dapat ditentukan"
		rekomendasi = "Baru 1 data tersedia. Tambahkan penimbangan berikutnya agar sistem dapat menghitung pertumbuhan."
		accuracy = "Estimasi Awal"
		rSquared = 0.5
		dataUsed = 1

		for i := 1; i <= horizonMonths; i++ {
			pDate := lastRecord.MeasurementDate.Add(time.Duration(i*30) * 24 * time.Hour).Format("2006-01-02")
			projectedPoints = append(projectedPoints, models.ProjectedPoint{
				Month:  i,
				Date:   pDate,
				Weight: predictedWeight,
			})
		}

	} else {
		// --- KASUS 2+: Regresi linear least-squares menggunakan hari aktual ---
		//
		// Model: Weight = slope * daysSinceFirst + intercept
		//
		// Kita pakai semua data yang ada (bukan hanya 3 terakhir)
		// agar prediksi semakin akurat seiring bertambahnya data.

		// Bangun array x (hari sejak pengukuran pertama) dan y (bobot)
		xVals := make([]float64, n)
		yVals := make([]float64, n)
		for i, rec := range records {
			daysSince := rec.MeasurementDate.Sub(firstRecord.MeasurementDate).Hours() / 24.0
			xVals[i] = daysSince
			yVals[i] = rec.Weight
		}

		slope, intercept, rSq := leastSquaresLinear(xVals, yVals)
		rSquared = math.Round(rSq*100) / 100
		dataUsed = n

		// ADG = slope (kg per hari) dari regresi
		adg = math.Round(slope*100) / 100

		// Titik x untuk hari terakhir
		lastDayX := xVals[n-1]

		// Hitung accuracy label berdasarkan r²
		if rSq >= 0.9 {
			accuracy = "Tinggi"
		} else if rSq >= 0.7 {
			accuracy = "Estimasi Linear"
		} else {
			accuracy = "Estimasi Awal"
		}

		// Generate 3 titik proyeksi (Bulan 1, 2, 3)
		for i := 1; i <= horizonMonths; i++ {
			futureDay := lastDayX + float64(i*30)
			pWeight := math.Round((slope*futureDay+intercept)*100) / 100
			if pWeight < 0 {
				pWeight = 0
			}
			pDate := lastRecord.MeasurementDate.Add(time.Duration(i*30) * 24 * time.Hour).Format("2006-01-02")
			projectedPoints = append(projectedPoints, models.ProjectedPoint{
				Month:  i,
				Date:   pDate,
				Weight: pWeight,
			})
		}

		// Prediksi akhir = titik proyeksi terakhir
		predictedWeight = projectedPoints[len(projectedPoints)-1].Weight

		// Tentukan trend berdasarkan slope regresi dan perubahan bobot
		currentWeight := lastRecord.Weight
		if predictedWeight > currentWeight && slope > 0.1 {
			trend = "meningkat pesat"
			rekomendasi = "SANGAT LAYAK DIPERTAHANKAN: Lanjutkan program pakan saat ini. Sapi sedang dalam fase pertumbuhan optimal."
		} else if predictedWeight > currentWeight {
			trend = "meningkat"
			rekomendasi = "LAYAK DIPERTAHANKAN: Sapi menunjukkan pertumbuhan positif. Pertahankan manajemen pakan saat ini."
		} else if math.Abs(slope) < 0.05 {
			trend = "stabil"
			rekomendasi = "PANTAU KETAT: Bobot sapi relatif stagnan. Evaluasi kecukupan pakan dan kondisi kesehatan."
		} else {
			trend = "menurun"
			rekomendasi = "PERLU PERHATIAN: Berat sapi cenderung menurun. Periksa manajemen pakan dan kesehatan sapi segera."
		}
	}

	return &models.PredictionResponse{
		CowID:            cowID,
		CurrentWeight:    lastRecord.Weight,
		PredictedWeight:  predictedWeight,
		PredictionDate:   finalPredictionDate,
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

// leastSquaresLinear menghitung slope, intercept, dan R² menggunakan metode least squares.
// x: hari aktual (mis. 0, 30, 60), y: bobot sapi (kg)
func leastSquaresLinear(x, y []float64) (slope, intercept, rSquared float64) {
	n := float64(len(x))
	if n == 0 {
		return 0, 0, 0
	}
	if n == 1 {
		return 0, y[0], 0.5
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		// Semua x sama → tidak bisa buat garis, gunakan rata-rata
		return 0, sumY / n, 0
	}

	slope = (n*sumXY - sumX*sumY) / denominator
	intercept = (sumY - slope*sumX) / n

	// Hitung R²
	yMean := sumY / n
	var ssRes, ssTot float64
	for i := 0; i < len(x); i++ {
		predicted := slope*x[i] + intercept
		ssRes += (y[i] - predicted) * (y[i] - predicted)
		ssTot += (y[i] - yMean) * (y[i] - yMean)
	}

	if ssTot == 0 {
		rSquared = 1.0 // Semua y sama → perfect fit (flat)
	} else {
		rSquared = 1.0 - (ssRes / ssTot)
		if rSquared < 0 {
			rSquared = 0
		}
	}

	return slope, intercept, rSquared
}
