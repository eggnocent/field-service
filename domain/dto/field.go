package dto

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type FieldRequest struct {
	Name         string                 `form:"name" validate:"required"`
	Code         string                 `form:"code" validate:"required"`
	PricePerHour int                    `form:"pricePerHour" validate:"required"`
	Images       []multipart.FileHeader `form:"images"`
}

type UpdateFieldRequest struct {
	Name         string                 `form:"name"`
	Code         string                 `form:"code"`
	PricePerHour int                    `form:"pricePerHour"`
	Images       []multipart.FileHeader `form:"images"`
}

type FieldResponse struct {
	UUID         uuid.UUID  `json:"uuid"`
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	PricePerHour int        `json:"pricePerHour"`
	Images       []string   `json:"images"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

type FieldDetailResponse struct {
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	PricePerHour int       `json:"pricePerHour"`
	Images       []string  `json:"images"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type FieldRequestParam struct {
	Page          int     `form:"page"`
	Limit         int     `form:"limit"`
	SortColumn    *string `form:"sortColumn"`
	SortDirection *string `form:"sortDirection"`
	SortOrder     *string `form:"sortOrder"`
}
