package public

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type GamesHandler struct {
	gameSvc *services.GameService
}

func NewGamesHandler(gameSvc *services.GameService) *GamesHandler {
	return &GamesHandler{gameSvc: gameSvc}
}

func failFromServiceErr(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case utils.AppError:
		return utils.Fail(c, e)
	case *utils.AppError:
		return utils.Fail(c, *e)
	default:
		return utils.Fail(c, utils.ErrInternal())
	}
}

func (h *GamesHandler) List(c *fiber.Ctx) error {
	ageStr := strings.TrimSpace(c.Query("age_category_id", ""))
	eduStr := strings.TrimSpace(c.Query("education_category_id", ""))
	sort := strings.TrimSpace(c.Query("sort", "newest"))
	pageStr := strings.TrimSpace(c.Query("page", "1"))
	limitStr := strings.TrimSpace(c.Query("limit", "24"))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return utils.Fail(c, utils.ErrBadRequest("page must be an integer >= 1"))
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		return utils.Fail(c, utils.ErrBadRequest("limit must be an integer between 1 and 100"))
	}

	var ageID *int64
	if ageStr != "" {
		v, err := strconv.ParseInt(ageStr, 10, 64)
		if err != nil || v < 1 {
			return utils.Fail(c, utils.ErrBadRequest("age_category_id must be an integer >= 1"))
		}
		ageID = &v
	}

	var eduID *int64
	if eduStr != "" {
		v, err := strconv.ParseInt(eduStr, 10, 64)
		if err != nil || v < 1 {
			return utils.Fail(c, utils.ErrBadRequest("education_category_id must be an integer >= 1"))
		}
		eduID = &v
	}

	dto, svcErr := h.gameSvc.ListPublicGames(c.Context(), services.ListPublicGamesInput{
		AgeCategoryID:       ageID,
		EducationCategoryID: eduID,
		Sort:                strings.ToLower(sort),
		Page:                page,
		Limit:               limit,
	})
	if svcErr != nil {
		return failFromServiceErr(c, svcErr)
	}

	return utils.Success(c, dto)
}

func (h *GamesHandler) Get(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id", ""))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer >= 1"))
	}

	dto, svcErr := h.gameSvc.GetPublicGameByID(c.Context(), id)
	if svcErr != nil {
		return failFromServiceErr(c, svcErr)
	}

	return utils.Success(c, dto)
}
