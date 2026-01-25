package public

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type LeaderboardHandler struct {
	svc *services.LeaderboardService
}

func NewLeaderboardHandler(svc *services.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{svc: svc}
}

func (h *LeaderboardHandler) Submit(c *fiber.Ctx) error {
	var req models.SubmitScoreRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	tokenGameID, ok := getTokenGameID(c)
	if !ok {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	guestID := strings.TrimSpace(c.Get("X-Guest-Id"))

	ctx := context.Background()

	resp, appErr := h.svc.SubmitScore(
		ctx,
		tokenGameID,
		guestID,
		req,
		"",
		"",
		"",
	)
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}

	return utils.Success(c, resp)
}

func getTokenGameID(c *fiber.Ctx) (int64, bool) {
	v := c.Locals(middleware.LocalPlayGameID)
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case float64:
		return int64(t), true
	default:
		return 0, false
	}
}
