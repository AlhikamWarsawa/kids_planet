package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

func Logging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		if err != nil {
			if hErr := c.App().Config().ErrorHandler(c, err); hErr != nil {
				return hErr
			}
		}

		status := c.Response().StatusCode()
		if status == 0 {
			status = fiber.StatusOK
		}
		duration := time.Since(start).Milliseconds()

		level := "info"
		if status >= fiber.StatusInternalServerError {
			level = "error"
		} else if status >= fiber.StatusBadRequest {
			level = "warn"
		}

		requestID := utils.RequestIDFromContext(c)
		if requestID == "" {
			requestID = "-"
		}

		log.Printf(
			"level=%s request_id=%s method=%s path=%s status=%d duration_ms=%d",
			level,
			requestID,
			c.Method(),
			c.OriginalURL(),
			status,
			duration,
		)

		return nil
	}
}
