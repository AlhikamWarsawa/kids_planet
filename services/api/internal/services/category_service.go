package services

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type CategoryService struct {
	ageRepo *repos.AgeCategoryRepo
	eduRepo *repos.EducationCategoryRepo
}

func NewCategoryService(ageRepo *repos.AgeCategoryRepo, eduRepo *repos.EducationCategoryRepo) *CategoryService {
	return &CategoryService{
		ageRepo: ageRepo,
		eduRepo: eduRepo,
	}
}

type CreateAgeCategoryInput struct {
	Label  string
	MinAge int
	MaxAge int
}

type UpdateAgeCategoryInput struct {
	Label  *string
	MinAge *int
	MaxAge *int
}

func (s *CategoryService) ListAgeCategories(ctx context.Context, q string, page, limit int) ([]repos.AgeCategory, *utils.AppError) {
	items, err := s.ageRepo.List(ctx, strings.TrimSpace(q), page, limit)
	if err != nil {
		ae := utils.ErrInternal()
		return nil, &ae
	}
	return items, nil
}

func (s *CategoryService) CreateAgeCategory(ctx context.Context, in CreateAgeCategoryInput) (*repos.AgeCategory, *utils.AppError) {
	label := strings.TrimSpace(in.Label)
	if label == "" {
		ae := utils.ErrBadRequest("label is required")
		return nil, &ae
	}
	if len(label) > 50 {
		ae := utils.ErrBadRequest("label is too long (max 50)")
		return nil, &ae
	}
	if in.MinAge < 0 {
		ae := utils.ErrBadRequest("min_age must be >= 0")
		return nil, &ae
	}
	if in.MaxAge < in.MinAge {
		ae := utils.ErrBadRequest("max_age must be >= min_age")
		return nil, &ae
	}

	exists, err := s.ageRepo.LabelExists(ctx, label, nil)
	if err != nil {
		ae := utils.ErrInternal()
		return nil, &ae
	}
	if exists {
		ae := utils.ErrBadRequest("age category label already exists")
		return nil, &ae
	}

	out, err := s.ageRepo.Create(ctx, repos.CreateAgeCategoryInput{
		Label:  label,
		MinAge: in.MinAge,
		MaxAge: in.MaxAge,
	})
	if err != nil {
		if isUniqueViolation(err) {
			ae := utils.ErrBadRequest("age category label already exists")
			return nil, &ae
		}
		ae := utils.ErrInternal()
		return nil, &ae
	}
	return out, nil
}

func (s *CategoryService) UpdateAgeCategory(ctx context.Context, id int64, in UpdateAgeCategoryInput) (*repos.AgeCategory, *utils.AppError) {
	if id <= 0 {
		ae := utils.ErrBadRequest("invalid id")
		return nil, &ae
	}

	var labelPtr *string
	if in.Label != nil {
		v := strings.TrimSpace(*in.Label)
		if v == "" {
			ae := utils.ErrBadRequest("label is required")
			return nil, &ae
		}
		if len(v) > 50 {
			ae := utils.ErrBadRequest("label is too long (max 50)")
			return nil, &ae
		}
		labelPtr = &v

		exID := id
		exists, err := s.ageRepo.LabelExists(ctx, v, &exID)
		if err != nil {
			ae := utils.ErrInternal()
			return nil, &ae
		}
		if exists {
			ae := utils.ErrBadRequest("age category label already exists")
			return nil, &ae
		}
	}

	if in.MinAge != nil {
		if *in.MinAge < 0 {
			ae := utils.ErrBadRequest("min_age must be >= 0")
			return nil, &ae
		}
	}
	if in.MaxAge != nil {
		if in.MinAge != nil && *in.MaxAge < *in.MinAge {
			ae := utils.ErrBadRequest("max_age must be >= min_age")
			return nil, &ae
		}
	}

	out, err := s.ageRepo.Update(ctx, id, repos.UpdateAgeCategoryInput{
		Label:  labelPtr,
		MinAge: in.MinAge,
		MaxAge: in.MaxAge,
	})
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			ae := utils.ErrNotFound("age category not found")
			return nil, &ae
		}
		if isUniqueViolation(err) {
			ae := utils.ErrBadRequest("age category label already exists")
			return nil, &ae
		}
		ae := utils.ErrInternal()
		return nil, &ae
	}
	return out, nil
}

func (s *CategoryService) DeleteAgeCategory(ctx context.Context, id int64) *utils.AppError {
	if id <= 0 {
		ae := utils.ErrBadRequest("invalid id")
		return &ae
	}
	if err := s.ageRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			ae := utils.ErrNotFound("age category not found")
			return &ae
		}
		if isFKViolation(err) {
			ae := utils.ErrBadRequest("age category is in use")
			return &ae
		}
		ae := utils.ErrInternal()
		return &ae
	}
	return nil
}

type CreateEducationCategoryInput struct {
	Name  string
	Icon  *string
	Color *string
}

type UpdateEducationCategoryInput struct {
	Name  *string
	Icon  *string
	Color *string
}

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (s *CategoryService) ListEducationCategories(ctx context.Context, q string, page, limit int) ([]repos.EducationCategory, *utils.AppError) {
	items, err := s.eduRepo.List(ctx, strings.TrimSpace(q), page, limit)
	if err != nil {
		ae := utils.ErrInternal()
		return nil, &ae
	}
	return items, nil
}

func (s *CategoryService) CreateEducationCategory(ctx context.Context, in CreateEducationCategoryInput) (*repos.EducationCategory, *utils.AppError) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		ae := utils.ErrBadRequest("name is required")
		return nil, &ae
	}
	if len(name) > 100 {
		ae := utils.ErrBadRequest("name is too long (max 100)")
		return nil, &ae
	}

	var iconNS sql.NullString
	if in.Icon != nil {
		v := strings.TrimSpace(*in.Icon)
		if v != "" {
			if len(v) > 100 {
				ae := utils.ErrBadRequest("icon is too long (max 100)")
				return nil, &ae
			}
			iconNS = sql.NullString{String: v, Valid: true}
		}
	}

	var colorNS sql.NullString
	if in.Color != nil {
		v := strings.TrimSpace(*in.Color)
		if v != "" {
			if len(v) > 20 {
				ae := utils.ErrBadRequest("color is too long (max 20)")
				return nil, &ae
			}
			if !hexColorRe.MatchString(v) {
				ae := utils.ErrBadRequest("color must be a hex code like #RRGGBB")
				return nil, &ae
			}
			colorNS = sql.NullString{String: v, Valid: true}
		}
	}

	exists, err := s.eduRepo.NameExists(ctx, name, nil)
	if err != nil {
		ae := utils.ErrInternal()
		return nil, &ae
	}
	if exists {
		ae := utils.ErrBadRequest("education category name already exists")
		return nil, &ae
	}

	out, err := s.eduRepo.Create(ctx, repos.CreateEducationCategoryInput{
		Name:  name,
		Icon:  iconNS,
		Color: colorNS,
	})
	if err != nil {
		if isUniqueViolation(err) {
			ae := utils.ErrBadRequest("education category name already exists")
			return nil, &ae
		}
		ae := utils.ErrInternal()
		return nil, &ae
	}
	return out, nil
}

func (s *CategoryService) UpdateEducationCategory(ctx context.Context, id int64, in UpdateEducationCategoryInput) (*repos.EducationCategory, *utils.AppError) {
	if id <= 0 {
		ae := utils.ErrBadRequest("invalid id")
		return nil, &ae
	}

	var namePtr *string
	if in.Name != nil {
		v := strings.TrimSpace(*in.Name)
		if v == "" {
			ae := utils.ErrBadRequest("name is required")
			return nil, &ae
		}
		if len(v) > 100 {
			ae := utils.ErrBadRequest("name is too long (max 100)")
			return nil, &ae
		}
		namePtr = &v

		exID := id
		exists, err := s.eduRepo.NameExists(ctx, v, &exID)
		if err != nil {
			ae := utils.ErrInternal()
			return nil, &ae
		}
		if exists {
			ae := utils.ErrBadRequest("education category name already exists")
			return nil, &ae
		}
	}

	var iconPtr *string
	if in.Icon != nil {
		v := strings.TrimSpace(*in.Icon)
		if v != "" && len(v) > 100 {
			ae := utils.ErrBadRequest("icon is too long (max 100)")
			return nil, &ae
		}
		iconPtr = &v
	}

	var colorPtr *string
	if in.Color != nil {
		v := strings.TrimSpace(*in.Color)
		if v != "" {
			if len(v) > 20 {
				ae := utils.ErrBadRequest("color is too long (max 20)")
				return nil, &ae
			}
			if !hexColorRe.MatchString(v) {
				ae := utils.ErrBadRequest("color must be a hex code like #RRGGBB")
				return nil, &ae
			}
		}
		colorPtr = &v
	}

	out, err := s.eduRepo.Update(ctx, id, repos.UpdateEducationCategoryInput{
		Name:  namePtr,
		Icon:  iconPtr,
		Color: colorPtr,
	})
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			ae := utils.ErrNotFound("education category not found")
			return nil, &ae
		}
		if isUniqueViolation(err) {
			ae := utils.ErrBadRequest("education category name already exists")
			return nil, &ae
		}
		ae := utils.ErrInternal()
		return nil, &ae
	}
	return out, nil
}

func (s *CategoryService) DeleteEducationCategory(ctx context.Context, id int64) *utils.AppError {
	if id <= 0 {
		ae := utils.ErrBadRequest("invalid id")
		return &ae
	}
	if err := s.eduRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			ae := utils.ErrNotFound("education category not found")
			return &ae
		}
		if isFKViolation(err) {
			ae := utils.ErrBadRequest("education category is in use")
			return &ae
		}
		ae := utils.ErrInternal()
		return &ae
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23505") ||
		strings.Contains(msg, "duplicate key value violates unique constraint")
}

func isFKViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23503") ||
		strings.Contains(msg, "violates foreign key constraint")
}
