package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

const LocalPlayerID = "player_id"

func AuthPlayerJWT(cfg config.Config) fiber.Handler {
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

		parsed, err := utils.ParsePlayerToken(cfg.JWT, tokenStr)
		if err != nil {
			if appErr, ok := err.(utils.AppError); ok {
				return utils.Fail(c, appErr)
			}
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		if parsed.Role != "player" || strings.TrimSpace(parsed.PlayerID) == "" {
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		c.Locals(LocalPlayerID, parsed.PlayerID)
		return c.Next()
	}
}
