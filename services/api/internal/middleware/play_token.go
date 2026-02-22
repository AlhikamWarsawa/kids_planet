package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

const (
	LocalPlayGameID    = "play_game_id"
	LocalPlayExp       = "play_exp"
	LocalPlaySessionID = "play_session_id"
	LocalPlaySubject   = "play_sub"
)

type PlayClaims struct {
	GameID    int64  `json:"game_id"`
	SessionID string `json:"session_id"`
	Typ       string `json:"typ"`
	jwt.RegisteredClaims
}

func PlayToken(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := strings.TrimSpace(c.Get("Authorization"))
		if auth == "" {
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(strings.TrimSpace(parts[0])) != "bearer" {
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		claims := &PlayClaims{}
		parser := jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithIssuer(cfg.JWT.Issuer),
		)

		_, err := parser.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte(cfg.JWT.Secret), nil
		})
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) ||
				errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
				errors.Is(err, jwt.ErrTokenMalformed) ||
				errors.Is(err, jwt.ErrTokenInvalidIssuer) ||
				errors.Is(err, jwt.ErrTokenUnverifiable) {
				return utils.Fail(c, utils.ErrUnauthorized())
			}
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		if claims.Typ != "play" {
			return utils.Fail(c, utils.ErrUnauthorized())
		}
		if claims.GameID <= 0 {
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		var exp time.Time
		if claims.ExpiresAt != nil {
			exp = claims.ExpiresAt.Time
		} else {
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		c.Locals(LocalPlayGameID, claims.GameID)
		c.Locals(LocalPlayExp, exp)
		c.Locals(LocalPlaySessionID, strings.TrimSpace(claims.SessionID))
		c.Locals(LocalPlaySubject, strings.TrimSpace(claims.Subject))

		return c.Next()
	}
}
