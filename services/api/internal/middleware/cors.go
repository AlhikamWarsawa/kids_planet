package middleware

import (
	"net/url"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

const (
	allowedMethods = "GET, POST, OPTIONS"
	allowedHeaders = "Authorization, Content-Type"
)

func CORS() fiber.Handler {
	allowedOrigins := buildAllowedOrigins()

	return func(c *fiber.Ctx) error {
		origin := normalizeOrigin(c.Get(fiber.HeaderOrigin))
		if origin == "" {
			if c.Method() == fiber.MethodOptions {
				return c.SendStatus(fiber.StatusNoContent)
			}
			return c.Next()
		}

		if _, ok := allowedOrigins[origin]; !ok {
			return utils.Fail(c, utils.ErrForbidden())
		}

		c.Append(fiber.HeaderVary, fiber.HeaderOrigin)
		c.Set(fiber.HeaderAccessControlAllowOrigin, origin)
		c.Set(fiber.HeaderAccessControlAllowCredentials, "true")
		c.Set(fiber.HeaderAccessControlAllowMethods, allowedMethods)
		c.Set(fiber.HeaderAccessControlAllowHeaders, allowedHeaders)

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

func buildAllowedOrigins() map[string]struct{} {
	allowed := map[string]struct{}{
		"http://localhost":      {},
		"http://localhost:5173": {},
	}

	if origin := normalizeOrigin(os.Getenv("APP_ORIGIN")); origin != "" {
		allowed[origin] = struct{}{}
	}

	return allowed
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	origin = strings.TrimSuffix(origin, "/")
	if origin == "" {
		return ""
	}

	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}
	if strings.TrimSpace(u.Scheme) == "" || strings.TrimSpace(u.Host) == "" {
		return ""
	}
	if u.Path != "" && u.Path != "/" {
		return ""
	}

	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
}
