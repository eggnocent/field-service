package models

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

type Field struct {
	ID            uint           `gorm:"primarKey;autoIncrement"`
	UUID          uuid.UUID      `gorm:"type:uuid;not null"`
	Code          string         `gorm:"varchar(15);not null"`
	Name          string         `gorm:"varchar(255);not null"`
	PricePerHour  int            `gorm:"type:int;not null"`
	Images        pq.StringArray `gorm:"type:text[];not null"`
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	DeletedAt     *time.Time
	FieldSchedule []FieldSchedule `gorm:"foreignkey:field_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
