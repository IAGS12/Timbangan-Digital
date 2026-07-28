package handlers

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/repositories"
	"smart-livestock-backend/utils"
)

type DashboardHandler struct {
	cowRepo    *repositories.CowRepository
	weightRepo *repositories.WeightRepository
}

func NewDashboardHandler(cowRepo *repositories.CowRepository, weightRepo *repositories.WeightRepository) *DashboardHandler {
	return &DashboardHandler{cowRepo: cowRepo, weightRepo: weightRepo}
}

// GetSummary GET /api/dashboard/summary
func (h *DashboardHandler) GetSummary(c *fiber.Ctx) error {
	// Total sapi aktif
	cows, totalCows, err := h.cowRepo.GetAll("active", "", 1, 1000)
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil data sapi")
	}

	// Total penimbangan
	totalWeighings, err := h.weightRepo.CountAll()
	if err != nil {
		return utils.InternalError(c, "Gagal menghitung penimbangan")
	}

	// Rata-rata berat
	avgWeight, err := h.weightRepo.GetAverageWeight()
	if err != nil {
		avgWeight = 0
	}

	// Cari sapi dengan ADG terbaik
	var bestGrowth fiber.Map
	if len(cows) > 0 {
		var bestADG float64
		var bestCow string
		var bestCode string
		for _, cow := range cows {
			if cow.LastADG != nil && *cow.LastADG > bestADG {
				bestADG = *cow.LastADG
				bestCow = cow.Name
				bestCode = cow.CowCode
			}
		}
		if bestCow != "" {
			bestGrowth = fiber.Map{
				"cow_code": bestCode,
				"cow_name": bestCow,
				"adg":      bestADG,
			}
		}
	}

	summary := fiber.Map{
		"total_cows_active": totalCows,
		"total_weighings":   totalWeighings,
		"average_weight":    avgWeight,
		"best_growth":       bestGrowth,
	}

	return utils.Success(c, summary)
}

// GetGrowthTrend GET /api/dashboard/growth
func (h *DashboardHandler) GetGrowthTrend(c *fiber.Ctx) error {
	// Ambil data untuk 6 bulan terakhir
	trends, err := h.weightRepo.GetGrowthTrend(6)
	if err != nil {
		return utils.InternalError(c, "Gagal mengambil data trend pertumbuhan")
	}

	// Balik array agar urut kronologis (waktu paling lama ke paling baru)
	for i, j := 0, len(trends)-1; i < j; i, j = i+1, j-1 {
		trends[i], trends[j] = trends[j], trends[i]
	}

	var formattedTrends []fiber.Map
	monthMap := map[string]string{
		"01": "Jan", "02": "Feb", "03": "Mar", "04": "Apr", "05": "Mei", "06": "Jun",
		"07": "Jul", "08": "Agu", "09": "Sep", "10": "Okt", "11": "Nov", "12": "Des",
	}

	for _, t := range trends {
		// Parse 'YYYY-MM' -> 'MMM YYYY'
		label := t.MonthGroup
		if len(label) == 7 { // "2026-07"
			year := label[0:4]
			monthStr := label[5:7]
			if mappedMonth, ok := monthMap[monthStr]; ok {
				label = mappedMonth + " " + year
			}
		}

		formattedTrends = append(formattedTrends, fiber.Map{
			"tanggal":       label,
			"beratRataRata": t.AvgWeight,
			"adgRataRata":   t.AvgADG,
		})
	}

	return utils.Success(c, formattedTrends)
}
