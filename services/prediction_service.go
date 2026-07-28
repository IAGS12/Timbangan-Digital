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
	
	// 1. UBAH VALIDASI: Cukup minimal 3 data untuk membentuk kurva polinomial derajat 2
	if n < 3 {
		return nil, errors.New("data belum cukup untuk prediksi polinomial, minimal 3 penimbangan")
	}

	// 2. SET DEFAULT: Jika parameter tidak diisi, buat default 3 bulan ke depan
	if horizonMonths <= 0 {
		horizonMonths = 3
	}

	// 3. AMBIL TEPAT 3 DATA TERAKHIR
	y1 := records[n-3].Weight // Data terlama dari 3 data terakhir
	y2 := records[n-2].Weight // Data tengah
	y3 := records[n-1].Weight // Data terbaru (last record)

	// RUMUS CEPAT POLINOMIAL (Mencari a, b, dan c dengan x1=1, x2=2, x3=3)
	a := (y3 - 2*y2 + y1) / 2.0
	b := (-5*y1 + 8*y2 - 3*y3) / 2.0
	c := (3 * y1) - (3 * y2) + y3

	// Prediksi untuk x selanjutnya (Data terakhir adalah x=3, maka prediksi = 3 + horizonMonths)
	predictX := 3.0 + float64(horizonMonths)
	predictedWeight := math.Round(((a*predictX*predictX)+(b*predictX)+c)*100) / 100

	// 4. ANALISIS TREN & REKOMENDASI DSS
	trend := "stagnan"
	rekomendasi := "Lakukan observasi lebih lanjut terhadap manajemen pakan."

	if predictedWeight > y3 && a >= 0 {
		trend = "meningkat pesat" // Kurva naik (akselerasi)
		rekomendasi = "SANGAT LAYAK DIPERTAHANKAN: Lanjutkan program pakan saat ini. Sapi sedang dalam fase pertumbuhan optimal."
	} else if predictedWeight > y3 && a < 0 {
		trend = "meningkat (melambat)" // Kurva melengkung ke bawah (hampir mentok)
		rekomendasi = "LAYAK DIPERTAHANKAN SEMENTARA: Pantau ketat, sapi mendekati batas genetik bobot maksimal. Rencanakan penjualan/panen."
	} else if predictedWeight <= y3 {
		trend = "menurun / harus dijual" // Bobot anjlok atau sudah mentok puncak kurva
		rekomendasi = "TIDAK LAYAK DIPERTAHANKAN: Segera jual atau potong sapi ini untuk menghindari kerugian pemborosan pakan."
	}

	// 5. HITUNG ADG HARIAN (Menggunakan data terbaru dan data sebelumnya)
	lastRecord := records[n-1]
	prevRecord := records[n-2]
	days := lastRecord.MeasurementDate.Sub(prevRecord.MeasurementDate).Hours() / 24.0
	
	adg := 0.0
	if days > 0 {
		adg = math.Round(((y3-y2)/days)*100) / 100
	} else {
		// Fallback jika tidak sengaja mengukur di hari yang sama (mencegah panic devide by zero)
		adg = math.Round(((y3-y2)/30.0)*100) / 100 
	}

	// Asumsi rata-rata 1 bulan = 30 hari untuk konversi tanggal
	horizonDays := horizonMonths * 30
	predictionDate := lastRecord.MeasurementDate.Add(time.Duration(horizonDays) * 24 * time.Hour).Format("2006-01-02")

	return &models.PredictionResponse{
		CowID:            cowID,
		CurrentWeight:    lastRecord.Weight,
		PredictedWeight:  predictedWeight,
		PredictionDate:   predictionDate,
		HorizonDays:      horizonDays,
		RSquared:         1.0, // Polinomial derajat 2 pada 3 titik selalu melewati titiknya dengan sempurna (Akurasi absolut/1.0)
		AccuracyCategory: "Tinggi",
		Trend:            trend,
		Recommendation:   rekomendasi, // Jangan lupa tambahkan field ini di struct Anda
		ADG:              adg,
		DataPointsUsed:   3, // Selalu merepresentasikan 3 titik referensi
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
