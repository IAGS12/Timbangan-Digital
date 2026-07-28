package routes

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/handlers"
)


func RegisterCowRoutes(protected fiber.Router, cowHandler *handlers.CowHandler, predictionHandler *handlers.PredictionHandler, weightHandler *handlers.WeightHandler) {
	cowsGroup := protected.Group("/cows")
	cowsGroup.Get("/", cowHandler.GetAll)
	cowsGroup.Get("/:id", cowHandler.GetByID)
	cowsGroup.Post("/", cowHandler.Create)
	cowsGroup.Put("/:id", cowHandler.Update)
	cowsGroup.Delete("/:id", cowHandler.Delete)
	cowsGroup.Get("/:id/prediction", predictionHandler.GetPrediction)
	cowsGroup.Get("/:id/weights", weightHandler.GetCowWeights)
}
