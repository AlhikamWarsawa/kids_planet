package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

func Recover() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				requestID := utils.RequestIDFromContext(c)
				if requestID == "" {
					requestID = "-"
				}

				log.Printf(
					"level=error request_id=%s method=%s path=%s status=%d msg=%q panic_recovered=%v",
					requestID,
					c.Method(),
					c.OriginalURL(),
					fiber.StatusInternalServerError,
					"panic recovered",
					r,
				)

				_ = utils.Internal(c, "")
				err = nil
			}
		}()

		return c.Next()
	}
}
