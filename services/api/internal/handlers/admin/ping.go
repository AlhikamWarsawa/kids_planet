package admin

import (
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type PingHandler struct{}

func NewPingHandler() *PingHandler { return &PingHandler{} }

func (h *PingHandler) Get(c *fiber.Ctx) error {
	return utils.Success(c, fiber.Map{"ok": true})
}
