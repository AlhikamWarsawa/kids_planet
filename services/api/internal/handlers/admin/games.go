package admin

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type GamesHandler struct {
	gameSvc *services.GameService
}

func NewGamesHandler(gameSvc *services.GameService) *GamesHandler {
	return &GamesHandler{gameSvc: gameSvc}
}

func (h *GamesHandler) List(c *fiber.Ctx) error {
	status := strings.TrimSpace(c.Query("status"))
	q := strings.TrimSpace(c.Query("q"))

	page := 1
	limit := 24

	if v := strings.TrimSpace(c.Query("page")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return utils.Fail(c, utils.ErrBadRequest("page must be an integer"))
		}
		page = n
	}
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return utils.Fail(c, utils.ErrBadRequest("limit must be an integer"))
		}
		limit = n
	}

	out, err := h.gameSvc.ListAdminGames(context.Background(), services.AdminListGamesInput{
		Status: status,
		Q:      q,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			return utils.Fail(c, appErr)
		}
		return utils.Fail(c, utils.ErrInternal())
	}
	return utils.Success(c, out)
}

func (h *GamesHandler) Create(c *fiber.Ctx) error {
	var req models.CreateGameRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	userAny := c.Locals(middleware.LocalUserID)
	userID, ok := userAny.(int64)
	if !ok || userID <= 0 {
		return utils.Fail(c, utils.ErrInternal())
	}

	out, err := h.gameSvc.CreateAdminGame(context.Background(), userID, req)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			return utils.Fail(c, appErr)
		}
		return utils.Fail(c, utils.ErrInternal())
	}
	return utils.Success(c, out)
}

func (h *GamesHandler) Update(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer"))
	}

	var req models.UpdateGameRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	out, err := h.gameSvc.UpdateAdminGame(context.Background(), id, req)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			return utils.Fail(c, appErr)
		}
		return utils.Fail(c, utils.ErrInternal())
	}
	return utils.Success(c, out)
}

func (h *GamesHandler) Publish(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer"))
	}

	out, err := h.gameSvc.PublishAdminGame(context.Background(), id)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			return utils.Fail(c, appErr)
		}
		return utils.Fail(c, utils.ErrInternal())
	}
	return utils.Success(c, out)
}

func (h *GamesHandler) Unpublish(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer"))
	}

	out, err := h.gameSvc.UnpublishAdminGame(context.Background(), id)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			return utils.Fail(c, appErr)
		}
		return utils.Fail(c, utils.ErrInternal())
	}
	return utils.Success(c, out)
}

func (h *GamesHandler) Upload(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer"))
	}
	if id < 1 {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer >= 1"))
	}

	fh, err := c.FormFile("file")
	if err != nil || fh == nil {
		return utils.Fail(c, utils.ErrBadRequest("file is required"))
	}

	f, err := fh.Open()
	if err != nil {
		return utils.Fail(c, utils.ErrInternal())
	}
	defer func() { _ = f.Close() }()

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		return utils.Fail(c, utils.ErrInternal())
	}

	out, upErr := h.gameSvc.UploadAdminGameZip(
		context.Background(),
		id,
		fh.Filename,
		rs,
		fh.Size,
		fh.Header.Get("Content-Type"),
	)
	if upErr != nil {
		if appErr, ok := upErr.(utils.AppError); ok {
			return utils.Fail(c, appErr)
		}
		return utils.Fail(c, utils.ErrInternal())
	}

	return utils.Success(c, out)
}
