package public

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type SessionsHandler struct {
	cfg        config.Config
	sessionSvc *services.SessionService
}

func NewSessionsHandler(cfg config.Config, sessionSvc *services.SessionService) *SessionsHandler {
	return &SessionsHandler{cfg: cfg, sessionSvc: sessionSvc}
}

func (h *SessionsHandler) Start(c *fiber.Ctx) error {
	var req models.StartSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	if req.GameID <= 0 {
		return utils.Fail(c, utils.ErrBadRequest("game_id is required"))
	}

	sub, appErr := parseOptionalPlayerSub(c, h.cfg)
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}

	resp, appErr := h.sessionSvc.StartSession(c.Context(), req.GameID, sub)
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}

	return utils.Success(c, resp)
}

func parseOptionalPlayerSub(c *fiber.Ctx, cfg config.Config) (string, *utils.AppError) {
	auth := strings.TrimSpace(c.Get("Authorization"))
	if auth == "" {
		return "", nil
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(strings.TrimSpace(parts[0])) != "bearer" {
		e := utils.ErrUnauthorized()
		return "", &e
	}

	tokenStr := strings.TrimSpace(parts[1])
	if tokenStr == "" {
		e := utils.ErrUnauthorized()
		return "", &e
	}

	parsed, err := utils.ParsePlayerToken(cfg.JWT, tokenStr)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			return "", &appErr
		}
		e := utils.ErrUnauthorized()
		return "", &e
	}

	return parsed.PlayerID, nil
}
