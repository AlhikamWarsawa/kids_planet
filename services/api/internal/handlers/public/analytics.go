package public

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

const (
	analyticsEventsPerMinute = 60
	analyticsRateWindow      = time.Minute
)

type AnalyticsHandler struct {
	cfg     config.Config
	repo    *repos.AnalyticsRepo
	limiter *analyticsRateLimiter
}

func NewAnalyticsHandler(cfg config.Config, repo *repos.AnalyticsRepo) *AnalyticsHandler {
	return &AnalyticsHandler{
		cfg:     cfg,
		repo:    repo,
		limiter: newAnalyticsRateLimiter(analyticsEventsPerMinute, analyticsRateWindow),
	}
}

func (h *AnalyticsHandler) TrackEvent(c *fiber.Ctx) error {
	var req models.AnalyticsEventRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	playToken := strings.TrimSpace(req.PlayToken)
	if playToken == "" {
		return utils.Fail(c, utils.ErrBadRequest("play_token is required"))
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return utils.Fail(c, utils.ErrBadRequest("name is required"))
	}

	if len(req.Data) > 0 && !json.Valid(req.Data) {
		return utils.Fail(c, utils.ErrBadRequest("data must be valid json"))
	}

	claims, err := parsePlayToken(playToken, h.cfg)
	if err != nil {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	sessionID := strings.TrimSpace(claims.SessionID)
	if sessionID == "" {
		sessionID = strings.TrimSpace(claims.Subject)
	}
	if sessionID == "" {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	if !h.limiter.Allow(sessionID) {
		return utils.Fail(c, utils.ErrRateLimited("rate limit exceeded"))
	}

	playerID := getPlayerIDFromClaims(claims)
	dataStr, err := composeAnalyticsData(req.Data, playerID)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("data must be valid json"))
	}

	if err := h.repo.InsertAnalyticsEvent(
		c.Context(),
		sessionID,
		claims.GameID,
		name,
		dataStr,
		c.IP(),
		c.Get("User-Agent"),
	); err != nil {
		return utils.Fail(c, utils.ErrInternal())
	}

	return utils.Success(c, models.AnalyticsEventResponse{Ok: true})
}

func getPlayerIDFromClaims(claims *analyticsPlayClaims) string {
	if claims == nil {
		return ""
	}

	playerID := strings.TrimSpace(claims.Subject)
	if playerID == "" {
		return ""
	}
	if _, err := uuid.Parse(playerID); err != nil {
		return ""
	}
	return playerID
}

func composeAnalyticsData(raw json.RawMessage, playerID string) (*string, error) {
	playerID = strings.TrimSpace(playerID)
	hasRaw := len(raw) > 0

	if !hasRaw && playerID == "" {
		return nil, nil
	}

	if !hasRaw && playerID != "" {
		encoded, err := json.Marshal(map[string]string{"player_id": playerID})
		if err != nil {
			return nil, err
		}
		v := string(encoded)
		return &v, nil
	}

	if playerID == "" {
		v := string(raw)
		return &v, nil
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}

	if obj, ok := parsed.(map[string]any); ok {
		obj["player_id"] = playerID
		encoded, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		v := string(encoded)
		return &v, nil
	}

	encoded, err := json.Marshal(map[string]any{
		"player_id": playerID,
		"payload":   parsed,
	})
	if err != nil {
		return nil, err
	}
	v := string(encoded)
	return &v, nil
}

type analyticsRateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	items  map[string]*rateBucket
}

type rateBucket struct {
	count   int
	resetAt time.Time
}

func newAnalyticsRateLimiter(limit int, window time.Duration) *analyticsRateLimiter {
	return &analyticsRateLimiter{
		limit:  limit,
		window: window,
		items:  make(map[string]*rateBucket),
	}
}

// Allow enforces a naive in-memory limit (max N events per window per session).
func (l *analyticsRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	bucket, ok := l.items[key]
	if !ok || now.After(bucket.resetAt) {
		l.items[key] = &rateBucket{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true
	}

	if bucket.count >= l.limit {
		return false
	}

	bucket.count++
	return true
}

type analyticsPlayClaims struct {
	GameID    int64  `json:"game_id"`
	SessionID string `json:"session_id"`
	Typ       string `json:"typ"`
	jwt.RegisteredClaims
}

func parsePlayToken(tokenStr string, cfg config.Config) (*analyticsPlayClaims, error) {
	claims := &analyticsPlayClaims{}
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
