package models

import "time"

type GameStatus string

const (
	GameStatusDraft    GameStatus = "draft"
	GameStatusActive   GameStatus = "active"
	GameStatusArchived GameStatus = "archived"
)

type AdminEducationCategoryDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type AdminGameDTO struct {
	ID                   int64                       `json:"id"`
	Title                string                      `json:"title"`
	Slug                 string                      `json:"slug"`
	Status               GameStatus                  `json:"status"`
	Thumbnail            string                      `json:"thumbnail,omitempty"`
	GameURL              string                      `json:"game_url,omitempty"`
	AgeCategoryID        int64                       `json:"age_category_id"`
	EducationCategoryIDs []int64                     `json:"education_category_ids,omitempty"`
	EducationCategories  []AdminEducationCategoryDTO `json:"education_categories,omitempty"`
	Free                 bool                        `json:"free"`
	CreatedAt            time.Time                   `json:"created_at"`
	UpdatedAt            time.Time                   `json:"updated_at"`
}

type CreateGameRequest struct {
	Title                string  `json:"title"`
	Slug                 string  `json:"slug"`
	Thumbnail            string  `json:"thumbnail,omitempty"`
	GameURL              string  `json:"game_url,omitempty"`
	AgeCategoryID        int64   `json:"age_category_id"`
	EducationCategoryIDs []int64 `json:"education_category_ids,omitempty"`
	Free                 *bool   `json:"free,omitempty"`
}

type UpdateGameRequest struct {
	Title                *string  `json:"title,omitempty"`
	Slug                 *string  `json:"slug,omitempty"`
	Thumbnail            *string  `json:"thumbnail,omitempty"`
	GameURL              *string  `json:"game_url,omitempty"`
	AgeCategoryID        *int64   `json:"age_category_id,omitempty"`
	EducationCategoryIDs *[]int64 `json:"education_category_ids,omitempty"`
	Free                 *bool    `json:"free,omitempty"`
}

type SetGameStatusRequest struct {
	Status GameStatus `json:"status"`
}
