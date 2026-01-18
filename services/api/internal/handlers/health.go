package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type HealthHandler struct {
	cfg config.Config
}

func NewHealthHandler(cfg config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

func (h *HealthHandler) Get(c *fiber.Ctx) error {
	if h.cfg.Env != "prod" && c.Query("fail") == "1" {
		return utils.Fail(c, utils.ErrBadRequest("simulated health error"))
	}

	return utils.Success(c, fiber.Map{
		"ok":      true,
		"service": "api",
		"time":    time.Now().UTC().Format(time.RFC3339),
		"env":     h.cfg.Env,
	})
}
