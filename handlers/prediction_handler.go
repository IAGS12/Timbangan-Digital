package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/services"
	"smart-livestock-backend/utils"
)

type PredictionHandler struct {
	predictionService *services.PredictionService
	decisionService   *services.DecisionService
}

func NewPredictionHandler(predictionService *services.PredictionService, decisionService *services.DecisionService) *PredictionHandler {
	return &PredictionHandler{
		predictionService: predictionService,
		decisionService:   decisionService,
	}
}

// GetPrediction GET /api/cows/:id/prediction?horizon=30
func (h *PredictionHandler) GetPrediction(c *fiber.Ctx) error {
	cowID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return utils.BadRequest(c, "ID tidak valid")
	}

	horizonDays, _ := strconv.Atoi(c.Query("horizon", "30"))
	horizonMonths := horizonDays / 30
	if horizonMonths < 1 {
		horizonMonths = 1
	}

	prediction, err := h.predictionService.PredictPolynomial(cowID, horizonMonths)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	// Evaluasi rekomendasi
	prediction.Recommendation = h.decisionService.Evaluate(prediction)
	prediction.Reason = h.decisionService.GetReason(prediction)

	return utils.Success(c, prediction)
}
