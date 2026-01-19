package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

const (
	LocalUserID = "user_id"
	LocalRole   = "role"
)

func AuthJWT(cfg config.Config) fiber.Handler {
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

		parsed, err := utils.ParseToken(cfg.JWT, tokenStr)
		if err != nil {
			if appErr, ok := err.(utils.AppError); ok {
				return utils.Fail(c, appErr)
			}
			return utils.Fail(c, utils.ErrUnauthorized())
		}

		c.Locals(LocalUserID, parsed.UserID)
		c.Locals(LocalRole, parsed.Role)

		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleAny := c.Locals(LocalRole)
		role, _ := roleAny.(string)

		if role != "admin" {
			return utils.Fail(c, utils.ErrForbidden())
		}
		return c.Next()
	}
}
