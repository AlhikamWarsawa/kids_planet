package public

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type LeaderboardHandler struct {
	cfg config.Config
	svc *services.LeaderboardService
}

func NewLeaderboardHandler(cfg config.Config, svc *services.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{cfg: cfg, svc: svc}
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

	tokenPlayerID := strings.TrimSpace(getTokenPlayerID(c))
	guestID := strings.TrimSpace(c.Get("X-Guest-Id"))
	sessionID := strings.TrimSpace(getTokenSessionID(c))

	resp, appErr := h.svc.SubmitScore(
		c.Context(),
		tokenGameID,
		tokenPlayerID,
		guestID,
		req,
		sessionID,
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

func getTokenSessionID(c *fiber.Ctx) string {
	v := c.Locals(middleware.LocalPlaySessionID)
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

func getTokenPlayerID(c *fiber.Ctx) string {
	v := c.Locals(middleware.LocalPlaySubject)
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
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
		if err != nil || v < 1 || v > 100 {
			return utils.Fail(c, utils.ErrBadRequest("limit must be an integer between 1 and 100"))
		}
		limit = v
	}

	items, svcErr := h.svc.GetTop(c.Context(), gameID, period, scope, limit)
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

func (h *LeaderboardHandler) GetSelf(c *fiber.Ctx) error {
	gameIDStr := strings.TrimSpace(c.Params("game_id", ""))
	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil || gameID < 1 {
		return utils.Fail(c, utils.ErrBadRequest("game_id must be an integer >= 1"))
	}

	period := strings.TrimSpace(c.Query("period", ""))
	scope := strings.TrimSpace(c.Query("scope", ""))

	member, fromPlayToken, tokenGameID, authErr := h.resolveSelfMember(c)
	if authErr != nil {
		return utils.Fail(c, *authErr)
	}

	scopeNorm := strings.ToLower(strings.TrimSpace(scope))
	if scopeNorm == "" {
		scopeNorm = "game"
	}
	if fromPlayToken && scopeNorm == "game" && tokenGameID > 0 && tokenGameID != gameID {
		return utils.Fail(c, utils.ErrForbidden())
	}

	dto, svcErr := h.svc.GetSelf(c.Context(), gameID, period, scope, member)
	if svcErr != nil {
		return failFromServiceErr(c, svcErr)
	}

	return utils.Success(c, dto)
}

func (h *LeaderboardHandler) resolveSelfMember(c *fiber.Ctx) (member string, fromPlayToken bool, tokenGameID int64, appErr *utils.AppError) {
	auth := strings.TrimSpace(c.Get("Authorization"))
	if auth == "" {
		e := utils.ErrUnauthorized()
		return "", false, 0, &e
	}

	tokenStr := extractBearerToken(auth)
	if tokenStr == "" {
		e := utils.ErrUnauthorized()
		return "", false, 0, &e
	}

	if parsed, err := utils.ParsePlayerToken(h.cfg.JWT, tokenStr); err == nil {
		return "p:" + parsed.PlayerID, false, 0, nil
	}

	playClaims, err := parseSelfPlayToken(tokenStr, h.cfg)
	if err != nil {
		e := utils.ErrUnauthorized()
		return "", false, 0, &e
	}

	playerID := strings.TrimSpace(playClaims.Subject)
	if playerID != "" {
		if _, parseErr := uuid.Parse(playerID); parseErr == nil {
			return "p:" + playerID, true, playClaims.GameID, nil
		}
	}

	sessionID := strings.TrimSpace(playClaims.SessionID)
	if sessionID != "" {
		return "s:" + sessionID, true, playClaims.GameID, nil
	}

	e := utils.ErrUnauthorized()
	return "", false, 0, &e
}

func extractBearerToken(auth string) string {
	auth = strings.TrimSpace(auth)
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if strings.ToLower(strings.TrimSpace(parts[0])) != "bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func parseSelfPlayToken(tokenStr string, cfg config.Config) (*middleware.PlayClaims, error) {
	claims := &middleware.PlayClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(cfg.JWT.Issuer),
	)

	_, err := parser.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims.Typ != "play" {
		return nil, errors.New("invalid token type")
	}
	if claims.GameID <= 0 {
		return nil, errors.New("invalid token game")
	}
	if claims.ExpiresAt == nil {
		return nil, errors.New("invalid token expiry")
	}

	return claims, nil
}
