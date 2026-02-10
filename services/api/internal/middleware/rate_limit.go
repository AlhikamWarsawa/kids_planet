package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/clients"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

const (
	leaderboardSubmitRateLimit  = 30
	leaderboardSubmitRateWindow = time.Minute
)

func RateLimitLeaderboardSubmit(valkey *clients.Valkey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if valkey == nil {
			return c.Next()
		}

		key := leaderboardSubmitRateKey(c)
		if key == "" {
			return c.Next()
		}

		count, err := valkey.IncrWithTTL(c.Context(), key, leaderboardSubmitRateWindow)
		if err != nil {
			return c.Next()
		}

		if count > int64(leaderboardSubmitRateLimit) {
			return utils.Fail(c, utils.ErrRateLimited("rate limit exceeded"))
		}

		return c.Next()
	}
}

func leaderboardSubmitRateKey(c *fiber.Ctx) string {
	sessionID := strings.TrimSpace(localString(c, LocalPlaySessionID))
	if sessionID != "" {
		return "rl:leaderboard:submit:session:" + sessionID
	}

	token := strings.TrimSpace(bearerToken(c.Get("Authorization")))
	if token != "" {
		return "rl:leaderboard:submit:token:" + hashString(token)
	}

	ip := strings.TrimSpace(c.IP())
	if ip != "" {
		return "rl:leaderboard:submit:ip:" + hashString(ip)
	}

	return ""
}

func bearerToken(auth string) string {
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

func hashString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func localString(c *fiber.Ctx, key string) string {
	v := c.Locals(key)
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return ""
	}
}
