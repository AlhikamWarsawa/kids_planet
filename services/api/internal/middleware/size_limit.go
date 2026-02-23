package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

const MaxRequestBodyBytes = 50 * 1024 * 1024

func SizeLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodPost {
			return c.Next()
		}

		contentLength := c.Request().Header.ContentLength()
		if contentLength > MaxRequestBodyBytes {
			return utils.Fail(c, utils.ErrZipTooLarge(MaxRequestBodyBytes))
		}

		return c.Next()
	}
}
