package routes

import (
	"github.com/gofiber/fiber/v2"
	"smart-livestock-backend/handlers"
)

// RegisterDashboardRoutes mendaftarkan endpoint dashboard summary
func RegisterDashboardRoutes(protected fiber.Router, dashboardHandler *handlers.DashboardHandler) {
	protected.Get("/dashboard/summary", dashboardHandler.GetSummary)
	protected.Get("/dashboard/growth", dashboardHandler.GetGrowthTrend)
}
