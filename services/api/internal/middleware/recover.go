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
				log.Printf(
					"panic_recovered=%v method=%s path=%s",
					r,
					c.Method(),
					c.OriginalURL(),
				)

				_ = utils.Fail(c, utils.ErrInternal())
				err = nil
			}
		}()

		return c.Next()
	}
}
