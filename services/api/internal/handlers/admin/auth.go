package admin

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type AuthHandler struct {
	cfg      config.Config
	userRepo *repos.UserRepo
}

func NewAuthHandler(cfg config.Config, userRepo *repos.UserRepo) *AuthHandler {
	return &AuthHandler{cfg: cfg, userRepo: userRepo}
}

type adminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type adminLoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req adminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		return utils.Fail(c, utils.ErrBadRequest("email and password are required"))
	}

	u, err := h.userRepo.FindByEmail(c.Context(), req.Email)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) || errors.Is(err, repos.ErrInvalidEmail) {
			return utils.Fail(c, utils.ErrUnauthorized())
		}
		return utils.Fail(c, utils.ErrInternal())
	}

	if u.Role != "admin" {
		return utils.Fail(c, utils.ErrForbidden())
	}

	if !u.PasswordHash.Valid {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	if err := utils.ComparePassword(u.PasswordHash.String, req.Password); err != nil {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	token, expiresIn, err := utils.GenerateAdminToken(h.cfg.JWT, u)
	if err != nil {
		return utils.Fail(c, utils.ErrInternal())
	}

	return utils.Success(c, adminLoginResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
	})
}
