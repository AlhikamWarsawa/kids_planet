package public

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

type playerAuthRequest struct {
	Email string `json:"email"`
	PIN   string `json:"pin"`
}

type playerAuthResponse struct {
	Token  string     `json:"token"`
	Player playerInfo `json:"player"`
}

type playerInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req playerAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	email, pin, appErr := parsePlayerAuthRequest(req)
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}

	_, err := h.userRepo.FindPlayerByEmail(c.Context(), email)
	if err == nil {
		return utils.Fail(c, utils.ErrBadRequest("player already exists"))
	}
	if !errors.Is(err, repos.ErrNotFound) {
		return utils.Fail(c, utils.ErrInternal())
	}

	pinHash, err := utils.HashPIN(pin)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("pin must be exactly 6 digits"))
	}

	player, err := h.userRepo.CreatePlayer(c.Context(), email, pinHash)
	if err != nil {
		if errors.Is(err, repos.ErrAlreadyExists) {
			return utils.Fail(c, utils.ErrBadRequest("player already exists"))
		}
		return utils.Fail(c, utils.ErrInternal())
	}

	token, _, err := utils.GeneratePlayerToken(h.cfg.JWT, player)
	if err != nil {
		return utils.Fail(c, utils.ErrInternal())
	}

	return c.Status(fiber.StatusOK).JSON(toPlayerAuthResponse(token, player))
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req playerAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	email, pin, appErr := parsePlayerAuthRequest(req)
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}

	player, err := h.userRepo.FindPlayerByEmail(c.Context(), email)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return utils.Fail(c, utils.ErrUnauthorized())
		}
		return utils.Fail(c, utils.ErrInternal())
	}

	if !player.PinHash.Valid {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	if err := utils.ComparePIN(player.PinHash.String, pin); err != nil {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	token, _, err := utils.GeneratePlayerToken(h.cfg.JWT, player)
	if err != nil {
		return utils.Fail(c, utils.ErrInternal())
	}

	return c.Status(fiber.StatusOK).JSON(toPlayerAuthResponse(token, player))
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func parsePlayerAuthRequest(req playerAuthRequest) (email string, pin string, appErr *utils.AppError) {
	email, err := repos.NormalizeEmail(req.Email)
	if err != nil {
		e := utils.ErrBadRequest("email is invalid")
		return "", "", &e
	}

	pin = strings.TrimSpace(req.PIN)
	if err := utils.ValidatePIN(pin); err != nil {
		e := utils.ErrBadRequest("pin must be exactly 6 digits")
		return "", "", &e
	}

	return email, pin, nil
}

func toPlayerAuthResponse(token string, player *repos.User) playerAuthResponse {
	pid := ""
	if player != nil && player.PublicID.Valid {
		pid = player.PublicID.String
	}

	email := ""
	if player != nil {
		email = player.Email
	}

	return playerAuthResponse{
		Token: token,
		Player: playerInfo{
			ID:    pid,
			Email: email,
		},
	}
}
