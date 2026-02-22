package public

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type HistoryHandler struct {
	historySvc *services.HistoryService
}

func NewHistoryHandler(historySvc *services.HistoryService) *HistoryHandler {
	return &HistoryHandler{historySvc: historySvc}
}

func (h *HistoryHandler) List(c *fiber.Ctx) error {
	pageStr := strings.TrimSpace(c.Query("page", "1"))
	limitStr := strings.TrimSpace(c.Query("limit", "10"))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return utils.Fail(c, utils.ErrBadRequest("page must be an integer >= 1"))
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		return utils.Fail(c, utils.ErrBadRequest("limit must be an integer between 1 and 100"))
	}

	playerID, ok := getPlayerID(c)
	if !ok {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	resp, svcErr := h.historySvc.ListPlayerHistory(c.Context(), playerID, page, limit)
	if svcErr != nil {
		return failFromServiceErr(c, svcErr)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func getPlayerID(c *fiber.Ctx) (string, bool) {
	v := c.Locals(middleware.LocalPlayerID)
	switch t := v.(type) {
	case string:
		playerID := strings.TrimSpace(t)
		return playerID, playerID != ""
	case []byte:
		playerID := strings.TrimSpace(string(t))
		return playerID, playerID != ""
	default:
		return "", false
	}
}
