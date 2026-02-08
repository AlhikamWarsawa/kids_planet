package admin

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type DashboardHandler struct {
	dashboardSvc *services.DashboardService
}

func NewDashboardHandler(dashboardSvc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardSvc: dashboardSvc}
}

func (h *DashboardHandler) Overview(c *fiber.Ctx) error {
	out, appErr := h.dashboardSvc.GetOverview(context.Background())
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, out)
}
