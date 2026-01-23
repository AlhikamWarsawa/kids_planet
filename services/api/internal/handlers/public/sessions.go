package public

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type SessionsHandler struct {
	sessionSvc *services.SessionService
}

func NewSessionsHandler(sessionSvc *services.SessionService) *SessionsHandler {
	return &SessionsHandler{sessionSvc: sessionSvc}
}

func (h *SessionsHandler) Start(c *fiber.Ctx) error {
	var req models.StartSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	if req.GameID <= 0 {
		return utils.Fail(c, utils.ErrBadRequest("game_id is required"))
	}

	sub := strings.TrimSpace("")

	resp, appErr := h.sessionSvc.StartSession(c.Context(), req.GameID, sub)
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}

	return utils.Success(c, resp)
}
