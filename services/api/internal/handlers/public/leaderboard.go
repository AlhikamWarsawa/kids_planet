package public

import (
	"context"
	"strconv"
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

func (h *LeaderboardHandler) GetTop(c *fiber.Ctx) error {
	gameIDStr := strings.TrimSpace(c.Params("game_id", ""))
	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil || gameID < 1 {
		return utils.Fail(c, utils.ErrBadRequest("game_id must be an integer >= 1"))
	}

	period := strings.TrimSpace(c.Query("period", ""))
	scope := strings.TrimSpace(c.Query("scope", ""))

	limit := 0
	limitStr := strings.TrimSpace(c.Query("limit", ""))
	if limitStr != "" {
		v, err := strconv.Atoi(limitStr)
		if err != nil {
			return utils.Fail(c, utils.ErrBadRequest("limit must be an integer between 1 and 100"))
		}
		limit = v
	}

	ctx := context.Background()

	items, svcErr := h.svc.GetTop(ctx, gameID, period, scope, limit)
	if svcErr != nil {
		return failFromServiceErr(c, svcErr)
	}

	p := strings.ToLower(strings.TrimSpace(period))
	if p == "" {
		p = "daily"
	}
	s := strings.ToLower(strings.TrimSpace(scope))
	if s == "" {
		s = "game"
	}
	if limit <= 0 {
		limit = 10
	}

	return utils.Success(c, models.LeaderboardViewResponse{
		GameID: gameID,
		Period: p,
		Scope:  s,
		Limit:  limit,
		Items:  items,
	})
}
