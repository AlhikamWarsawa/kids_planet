package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()
		duration := time.Since(start).Milliseconds()

		log.Printf(
			"method=%s path=%s status=%d duration_ms=%d",
			c.Method(),
			c.OriginalURL(),
			status,
			duration,
		)

		return err
	}
}
