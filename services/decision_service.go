package services

import (
	"fmt"
	"smart-livestock-backend/models"
)

const (
	ThresholdADGGood    = 0.3  
	ThresholdADGStagnan = 0.0  
	ThresholdRSquared   = 0.5  
)

type DecisionService struct{}

func NewDecisionService() *DecisionService {
	return &DecisionService{}
}

func (s *DecisionService) Evaluate(prediction *models.PredictionResponse) string {
	adg := prediction.ADG
	rSquared := prediction.RSquared
	trend := prediction.Trend

	if adg < ThresholdADGStagnan && (trend == "menurun" || trend == "menurun / harus dijual") {
		return "TIDAK_LAYAK_DIPERTAHANKAN"
	}

	if adg > ThresholdADGGood && (trend == "meningkat" || trend == "meningkat pesat" || trend == "meningkat (melambat)") && rSquared >= ThresholdRSquared {
		return "LAYAK_DIPERTAHANKAN"
	}

	return "PERLU_EVALUASI"
}

func (s *DecisionService) GetReason(prediction *models.PredictionResponse) string {
	if prediction == nil || prediction.DataPointsUsed < 3 {
		return ""
	}
	// Ambil status keputusan menggunakan metode Evaluate
	decisionStatus := s.Evaluate(prediction)
	adg := prediction.ADG
	predictedWeight := prediction.PredictedWeight
	days := prediction.HorizonDays
	points := prediction.DataPointsUsed

	switch decisionStatus {
	case "LAYAK_DIPERTAHANKAN":
		return fmt.Sprintf("Dari catatan %d kali penimbangan, pertumbuhan sapi ini sangat bagus dan konsisten naik rata-rata +%.2f Kg setiap hari. Jika terus dirawat dengan pakan yang sama, dalam %d hari ke depan beratnya diperkirakan bisa mencapai sekitar %.2f Kg. Sapi ini sangat sehat dan efisien dalam menyerap pakan.", points, adg, days, predictedWeight)
	case "PERLU_EVALUASI":
		return fmt.Sprintf("Dari catatan %d kali penimbangan, kenaikan berat sapi ini tergolong lambat (rata-rata +%.2f Kg per hari). Dalam %d hari ke depan beratnya diperkirakan hanya mencapai sekitar %.2f Kg. Pertumbuhannya kurang stabil, sebaiknya periksa kualitas pakan atau cek kondisi fisiknya.", points, adg, days, predictedWeight)
	default:
		return fmt.Sprintf("Dari catatan %d kali penimbangan, sapi ini mengalami penurunan berat badan atau macet tumbuh (ADG harian %.2f Kg). Dalam %d hari ke depan beratnya diperkirakan stagnan di sekitar %.2f Kg. Merawat sapi ini lebih lama berisiko merugi karena pemborosan pakan.", points, adg, days, predictedWeight)
	}
}
