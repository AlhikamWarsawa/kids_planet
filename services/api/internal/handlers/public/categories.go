package public

import (
	"context"
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

type publicAgeCategoryDTO struct {
	ID     int64  `json:"id"`
	Label  string `json:"label"`
	MinAge int    `json:"min_age"`
	MaxAge int    `json:"max_age"`
}

type publicEducationCategoryDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (h *CategoriesHandler) List(c *fiber.Ctx) error {
	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	categoryType := strings.TrimSpace(strings.ToLower(c.Query("type")))
	switch categoryType {
	case "", "age", "education":
	default:
		return utils.Fail(c, utils.ErrBadRequest("type must be one of: age, education"))
	}

	ages := make([]publicAgeCategoryDTO, 0)
	if categoryType == "" || categoryType == "age" {
		ageItems, ageErr := h.categorySvc.ListAgeCategories(ctx, "", 1, 500)
		if ageErr != nil {
			return utils.Fail(c, *ageErr)
		}

		ages = make([]publicAgeCategoryDTO, 0, len(ageItems))
		for _, item := range ageItems {
			ages = append(ages, publicAgeCategoryDTO{
				ID:     item.ID,
				Label:  item.Label,
				MinAge: item.MinAge,
				MaxAge: item.MaxAge,
			})
		}
	}

	education := make([]publicEducationCategoryDTO, 0)
	if categoryType == "" || categoryType == "education" {
		educationItems, eduErr := h.categorySvc.ListEducationCategories(ctx, "", 1, 500)
		if eduErr != nil {
			return utils.Fail(c, *eduErr)
		}

		education = make([]publicEducationCategoryDTO, 0, len(educationItems))
		for _, item := range educationItems {
			education = append(education, publicEducationCategoryDTO{
				ID:   item.ID,
				Name: item.Name,
			})
		}
	}

	return utils.Success(c, fiber.Map{
		"age_categories":       ages,
		"education_categories": education,
	})
}
