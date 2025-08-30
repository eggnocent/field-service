package repository

import (
	"context"
	"errors"
	error2 "field-service/common/error"
	errConstant "field-service/constants/error"
	error3 "field-service/constants/error/time"
	"field-service/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TimeRepository struct {
	db *gorm.DB
}

type ITimeRepository interface {
	FindAll(context.Context) ([]models.Time, error)
	FindByUUID(context.Context, string) (*models.Time, error)
	FindByID(context.Context, int) (*models.Time, error)
	Create(context.Context, *models.Time) (*models.Time, error)
}

func NewTimeRepository(db *gorm.DB) ITimeRepository {
	return &TimeRepository{db: db}
}

func (r *TimeRepository) FindAll(ctx context.Context) ([]models.Time, error) {
	var times []models.Time
	err := r.db.WithContext(ctx).Find(&times).Error
	if err != nil {
		return nil, error2.WrapError(errConstant.ErrSQLError)
	}

	return times, nil
}

func (r *TimeRepository) FindByUUID(ctx context.Context, uuid string) (*models.Time, error) {
	var times models.Time
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&times).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, error2.WrapError(error3.ErrTimeNotFound)
		}
		return nil, error2.WrapError(errConstant.ErrSQLError)
	}

	return &times, nil
}

func (r *TimeRepository) FindByID(ctx context.Context, id int) (*models.Time, error) {
	var times models.Time
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&times).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, error2.WrapError(error3.ErrTimeNotFound)
		}
		return nil, error2.WrapError(errConstant.ErrSQLError)
	}
	return &times, nil
}

func (r *TimeRepository) Create(ctx context.Context, time *models.Time) (*models.Time, error) {
	time.UUID = uuid.New()
	err := r.db.WithContext(ctx).Create(&time).Error
	if err != nil {
		return nil, error2.WrapError(errConstant.ErrSQLError)
	}

	return time, nil
}
