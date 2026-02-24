package utils

import "github.com/gofiber/fiber/v2"

func Success(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
	})
}

func Fail(c *fiber.Ctx, appErr AppError) error {
	return WriteError(c, appErr)
}
