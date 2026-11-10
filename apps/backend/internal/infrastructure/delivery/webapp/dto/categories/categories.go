package dto

import (
	"time"

	"market/backend/internal/business/domain"
)

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// UpdateCategoryRequest is a partial update: an absent or null field stays
// as it is, and an empty description clears it.
type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewCategoryResponse(category domain.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		Active:      category.Active,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}
