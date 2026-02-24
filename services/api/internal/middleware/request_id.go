package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := strings.TrimSpace(c.Get(utils.RequestIDHeader))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Locals(utils.RequestIDLocalKey, requestID)
		c.Set(utils.RequestIDHeader, requestID)

		return c.Next()
	}
}
