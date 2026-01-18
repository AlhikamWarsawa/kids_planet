package utils

import "github.com/gofiber/fiber/v2"

func Success(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
	})
}

func Fail(c *fiber.Ctx, appErr AppError) error {
	if appErr.Code == "" {
		appErr = ErrInternal()
	}
	if appErr.HTTPStatus == 0 {
		appErr.HTTPStatus = fiber.StatusInternalServerError
	}

	return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	})
}
