package admin

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type CategoriesHandler struct {
	categorySvc *services.CategoryService
}

func NewCategoriesHandler(categorySvc *services.CategoryService) *CategoriesHandler {
	return &CategoriesHandler{categorySvc: categorySvc}
}

type CreateAgeCategoryRequest struct {
	Label  string `json:"label"`
	MinAge int    `json:"min_age"`
	MaxAge int    `json:"max_age"`
}

type UpdateAgeCategoryRequest struct {
	Label  *string `json:"label,omitempty"`
	MinAge *int    `json:"min_age,omitempty"`
	MaxAge *int    `json:"max_age,omitempty"`
}

func (h *CategoriesHandler) ListAge(c *fiber.Ctx) error {
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

	items, appErr := h.categorySvc.ListAgeCategories(context.Background(), q, page, limit)
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, fiber.Map{
		"items": items,
		"page":  page,
		"limit": limit,
	})
}

func (h *CategoriesHandler) CreateAge(c *fiber.Ctx) error {
	var req CreateAgeCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	out, appErr := h.categorySvc.CreateAgeCategory(context.Background(), services.CreateAgeCategoryInput{
		Label:  req.Label,
		MinAge: req.MinAge,
		MaxAge: req.MaxAge,
	})
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, out)
}

func (h *CategoriesHandler) UpdateAge(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer"))
	}

	var req UpdateAgeCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	out, appErr := h.categorySvc.UpdateAgeCategory(context.Background(), id, services.UpdateAgeCategoryInput{
		Label:  req.Label,
		MinAge: req.MinAge,
		MaxAge: req.MaxAge,
	})
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, out)
}

func (h *CategoriesHandler) DeleteAge(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer"))
	}

	if appErr := h.categorySvc.DeleteAgeCategory(context.Background(), id); appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, fiber.Map{"deleted": true})
}

type CreateEducationCategoryRequest struct {
	Name  string  `json:"name"`
	Icon  *string `json:"icon,omitempty"`
	Color *string `json:"color,omitempty"`
}

type UpdateEducationCategoryRequest struct {
	Name  *string `json:"name,omitempty"`
	Icon  *string `json:"icon,omitempty"`
	Color *string `json:"color,omitempty"`
}

func (h *CategoriesHandler) ListEducation(c *fiber.Ctx) error {
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

	items, appErr := h.categorySvc.ListEducationCategories(context.Background(), q, page, limit)
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, fiber.Map{
		"items": items,
		"page":  page,
		"limit": limit,
	})
}

func (h *CategoriesHandler) CreateEducation(c *fiber.Ctx) error {
	var req CreateEducationCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	out, appErr := h.categorySvc.CreateEducationCategory(context.Background(), services.CreateEducationCategoryInput{
		Name:  req.Name,
		Icon:  req.Icon,
		Color: req.Color,
	})
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, out)
}

func (h *CategoriesHandler) UpdateEducation(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer"))
	}

	var req UpdateEducationCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, utils.ErrBadRequest("invalid json body"))
	}

	out, appErr := h.categorySvc.UpdateEducationCategory(context.Background(), id, services.UpdateEducationCategoryInput{
		Name:  req.Name,
		Icon:  req.Icon,
		Color: req.Color,
	})
	if appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, out)
}

func (h *CategoriesHandler) DeleteEducation(c *fiber.Ctx) error {
	idStr := strings.TrimSpace(c.Params("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest("id must be an integer"))
	}

	if appErr := h.categorySvc.DeleteEducationCategory(context.Background(), id); appErr != nil {
		return utils.Fail(c, *appErr)
	}
	return utils.Success(c, fiber.Map{"deleted": true})
}
