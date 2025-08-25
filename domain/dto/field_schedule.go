package dto

import (
	"field-service/constants"
	"github.com/google/uuid"
	"time"
)

type FieldScheduleRequest struct {
	FieldID string   `json:"field_id" validate:"required"`
	Date    string   `json:"date" validate:"required"`
	TimeIDs []string `json:"time_ids" validate:"required"`
}

type GenerateFieldScheduleForOneMonthRequest struct {
	FieldID string `json:"field_id" validate:"required"`
}

type UpdateFieldScheduleRequest struct {
	Date   string `json:"date" validate:"required"`
	TimeID string `json:"time_id" validate:"required"`
}

type UpdateStatusFieldScheduleRequest struct {
	FieldScheduleIDs []string `json:"field_schedule_ids" validate:"required"`
}

type FieldScheduleResponse struct {
	UUID         uuid.UUID                         `json:"uuid"`
	FieldName    string                            `json:"field_name"`
	PricePerHour int                               `json:"price_per_hour"`
	Date         string                            `json:"date"`
	Status       constants.FieldScheduleStatusName `json:"status"`
	Time         string                            `json:"time"`
	CreatedAt    *time.Time                        `json:"created_at"`
	UpdatedAt    *time.Time                        `json:"updated_at"`
}

type FieldScheduleForBookingRespose struct {
	UUID         uuid.UUID                         `json:"uuid"`
	PricePerHour string                            `json:"price_per_hour"`
	Date         string                            `json:"date"`
	Status       constants.FieldScheduleStatusName `json:"status"`
	Time         string                            `json:"time"`
}

type FieldScheduleRequestParam struct {
	Page       int     `form:"page" validate:"required"`
	Limit      int     `form:"limit" validate:"required"`
	SortColumn *string `form:"sortColumn" validate:"required"`
	SortOrder  *string `form:"sortOrder" validate:"required"`
}

type FieldScheduleByFieldIDandDateRequestParam struct {
	Date string `form:"date" validate:"required"`
}
