package admin

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type MeHandler struct {
	userRepo *repos.UserRepo
}

func NewMeHandler(userRepo *repos.UserRepo) *MeHandler {
	return &MeHandler{userRepo: userRepo}
}

func (h *MeHandler) Get(c *fiber.Ctx) error {
	uidAny := c.Locals(middleware.LocalUserID)
	uid, ok := uidAny.(int64)
	if !ok || uid <= 0 {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	u, err := h.userRepo.FindByID(ctx, uid)
	if err != nil {
		if err == repos.ErrNotFound {
			return utils.Fail(c, utils.ErrUnauthorized())
		}
		return utils.Fail(c, utils.ErrInternal())
	}

	if u.Role != "admin" {
		return utils.Fail(c, utils.ErrForbidden())
	}
	if u.Status != "" && u.Status != "active" {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	return utils.Success(c, fiber.Map{
		"id":    u.ID,
		"email": u.Email,
		"role":  u.Role,
	})
}
