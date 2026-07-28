package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/exports"
	"smart-livestock-backend/repositories"
	"smart-livestock-backend/services"
)

type ExportHandler struct {
	cowRepo           *repositories.CowRepository
	weightRepo        *repositories.WeightRepository
	predictionService *services.PredictionService
	decisionService   *services.DecisionService
}

func NewExportHandler(
	cowRepo *repositories.CowRepository,
	weightRepo *repositories.WeightRepository,
	predictionService *services.PredictionService,
	decisionService *services.DecisionService,
) *ExportHandler {
	return &ExportHandler{
		cowRepo:           cowRepo,
		weightRepo:        weightRepo,
		predictionService: predictionService,
		decisionService:   decisionService,
	}
}

// ExportExcel GET /api/export/excel — Unduh laporan Excel (.xlsx) dengan multi-sheet & filter
func (h *ExportHandler) ExportExcel(c *fiber.Ctx) error {
	breedFilter := c.Query("breed")

	// Fetch all cows matching criteria (up to 5000 records)
	cows, _, err := h.cowRepo.GetAll("", breedFilter, 1, 5000)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data sapi dari database"})
	}

	var excelData []exports.ExcelCowData
	for _, cowItem := range cows {
		weighings, _ := h.weightRepo.GetByCowID(cowItem.ID)

		lastW := 0.0
		lastADG := 0.0
		if cowItem.LastWeight != nil {
			lastW = *cowItem.LastWeight
		}
		if cowItem.LastADG != nil {
			lastADG = *cowItem.LastADG
		}

		statusDSS := "Belum Cukup Data (< 3x Timbang)"
		predictedW := 0.0
		if len(weighings) >= 3 {
			pred, predErr := h.predictionService.PredictPolynomial(cowItem.ID, 3)
			if predErr == nil && pred != nil {
				predictedW = pred.PredictedWeight
				evalStatus := h.decisionService.Evaluate(pred)
				switch evalStatus {
				case "LAYAK_DIPERTAHANKAN":
					statusDSS = "LAYAK DIPERTAHANKAN"
				case "PERLU_EVALUASI":
					statusDSS = "PERLU EVALUASI"
				case "TIDAK_LAYAK_DIPERTAHANKAN":
					statusDSS = "TIDAK LAYAK DIPERTAHANKAN"
				default:
					statusDSS = evalStatus
				}
			}
		}

		excelData = append(excelData, exports.ExcelCowData{
			Cow:             cowItem.Cow,
			Weighings:       weighings,
			LastWeight:      lastW,
			LastADG:         lastADG,
			PredictedWeight: predictedW,
			StatusDSS:       statusDSS,
		})
	}

	excelBytes, err := exports.GenerateExcelReport(excelData, breedFilter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat berkas Excel: " + err.Error()})
	}

	fileName := fmt.Sprintf("Laporan_TimbangSapi_IoT_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	return c.Send(excelBytes)
}
