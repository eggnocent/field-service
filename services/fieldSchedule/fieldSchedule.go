package services

import (
	"context"
	"field-service/common/util"
	"field-service/constants"
	errFieldSchedule "field-service/constants/error/fieldSchedule"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type FieldScheduleService struct {
	repository repositories.IRepositoryRegistry
}

type IFieldScheduleService interface {
	GetAllWithPagination(context.Context, *dto.FieldScheduleRequestParam) (*util.PaginationResult, error)
	GetAllByFieldIDandDate(context.Context, string, string) ([]dto.FieldScheduleForBookingRespose, error)
	GetByUUID(context.Context, string) (*dto.FieldScheduleResponse, error)
	GenerateScheduleForOneMonth(context.Context, *dto.GenerateFieldScheduleForOneMonthRequest) error
	Create(context.Context, *dto.FieldScheduleRequest) error
	Update(context.Context, string, *dto.UpdateFieldScheduleRequest) (*dto.FieldScheduleResponse, error)
	UpdateStatus(context.Context, *dto.UpdateStatusFieldScheduleRequest) error
	Delete(context.Context, string) error
}

func NewFieldScheduleService(repository repositories.IRepositoryRegistry) IFieldScheduleService {
	return &FieldScheduleService{repository: repository}
}

func (f *FieldScheduleService) GetAllWithPagination(ctx context.Context, param *dto.FieldScheduleRequestParam) (*util.PaginationResult, error) {
	fieldSchedule, total, err := f.repository.GetFieldSchedule().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	fieldScheduleResults := make([]*dto.FieldScheduleResponse, 0, len(fieldSchedule))
	for _, schedule := range fieldSchedule {
		fieldScheduleResults = append(fieldScheduleResults, &dto.FieldScheduleResponse{
			UUID:         schedule.UUID,
			FieldName:    schedule.Field.Name,
			Date:         schedule.Date.Format("2006-01-02"),
			PricePerHour: schedule.Field.PricePerHour,
			Status:       schedule.Status.GetStatusString(),
			Time:         fmt.Sprintf("%s - %s", schedule.Time.StartTime, schedule.Time.EndTime),
			CreatedAt:    schedule.CreatedAt,
			UpdatedAt:    schedule.UpdatedAt,
		})

	}
	pagination := &util.PaginationParams{
		Count: total,
		Page:  param.Page,
		Data:  fieldScheduleResults,
	}

	response := util.GeneratePagination(*pagination)
	return &response, nil
}

func (f *FieldScheduleService) convertMonthName(inputDate string) string {
	date, err := time.Parse("2006-01-02", inputDate)
	if err != nil {
		return ""
	}

	indonesianMonth := map[string]string{
		"Jan": "Jan",
		"Feb": "Feb",
		"Mar": "Mar",
		"Apr": "Apr",
		"May": "Mei",
		"Jun": "Jun",
		"Jul": "Jul",
		"Aug": "Agu",
		"Sep": "Sep",
		"Oct": "Okt",
		"Nov": "Nov",
		"Dec": "Des",
	}

	formattedDate := date.Format("02 Jan")
	day := formattedDate[:3]
	inputDate = formattedDate[:3]

	formattedDate = fmt.Sprintf("%s %s", day, indonesianMonth[inputDate])

	return formattedDate
}

func (f *FieldScheduleService) GetAllByFieldIDandDate(ctx context.Context, uuid, date string) ([]dto.FieldScheduleForBookingRespose, error) {
	field, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	fieldSchedule, err := f.repository.GetFieldSchedule().FindAllByFieldIDAndDate(ctx, int(field.ID), date)
	if err != nil {
		return nil, err
	}

	fieldScheduleResult := make([]dto.FieldScheduleForBookingRespose, 0, len(fieldSchedule))
	for _, schedule := range fieldSchedule {
		pricePerHour := float64(schedule.Field.PricePerHour)
		fieldScheduleResult = append(fieldScheduleResult, dto.FieldScheduleForBookingRespose{
			UUID:         schedule.UUID,
			Date:         f.convertMonthName(schedule.Date.Format("2006-01-02")),
			Time:         schedule.Time.StartTime,
			Status:       schedule.Status.GetStatusString(),
			PricePerHour: util.RupiahFormat(&pricePerHour),
		})
	}

	return fieldScheduleResult, nil
}

func (f *FieldScheduleService) GetByUUID(ctx context.Context, uuid string) (*dto.FieldScheduleResponse, error) {
	fieldSchedule, err := f.repository.GetFieldSchedule().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	response := dto.FieldScheduleResponse{
		UUID:         fieldSchedule.UUID,
		FieldName:    fieldSchedule.Field.Name,
		PricePerHour: fieldSchedule.Field.PricePerHour,
		Date:         fieldSchedule.Date.Format(time.DateOnly),
		Status:       fieldSchedule.Status.GetStatusString(),
		Time:         fmt.Sprintf("%s - %s", fieldSchedule.Time.StartTime, fieldSchedule.Time.EndTime),
		CreatedAt:    fieldSchedule.CreatedAt,
		UpdatedAt:    fieldSchedule.UpdatedAt,
	}

	return &response, nil
}

func (f *FieldScheduleService) Create(ctx context.Context, request *dto.FieldScheduleRequest) error {
	// apakah ada data lapangan untuk menambahkan data field schedulenya, jika ada
	field, err := f.repository.GetField().FindByUUID(ctx, request.FieldID)
	if err != nil {
		return err
	}

	// lanjut kesini, looping data waktu yang akan dimasukan ke jadwal lapangannya, karena lebih dari 1 time nya makanya pake looping
	fieldSchedules := make([]models.FieldSchedule, 0, len(request.TimeIDs))
	dateParsed, _ := time.Parse(time.DateOnly, request.Date)
	for _, timeID := range request.TimeIDs {
		scheduleTime, err := f.repository.GetTime().FindByUUID(ctx, timeID)
		// ada pengecekan, apakah jadwal di table time nya itu ada atau tidak, jika ada
		if err != nil {
			return err
		}

		// masukan kesini untuk mengecek berdasarkan date dan time untuk jadwal lapangan nya, apakah jadwal lapangan berdasarkan waktu dan tanggal yang diinputkan itu ada?
		schedule, err := f.repository.GetFieldSchedule().FindByDateAndTimeID(ctx, request.Date, int(scheduleTime.ID), int(field.ID))
		if err != nil {
			return err
		}

		// jika ada, jangan di create kembali
		if schedule != nil {
			return errFieldSchedule.ErrFieldScheduleExists
		}

		// jika masih kosong, maka akan kesini
		fieldSchedules = append(fieldSchedules, models.FieldSchedule{
			UUID:    uuid.New(),
			FieldID: field.ID,
			TimeID:  uint(scheduleTime.ID),
			Date:    dateParsed,
			Status:  constants.Available,
		})
	}

	// lalu insert kan
	err = f.repository.GetFieldSchedule().Create(ctx, fieldSchedules)
	if err != nil {
		return err
	}

	return nil
}

func (f *FieldScheduleService) GenerateScheduleForOneMonth(ctx context.Context, request *dto.GenerateFieldScheduleForOneMonthRequest) error {
	field, err := f.repository.GetField().FindByUUID(ctx, request.FieldID)
	if err != nil {
		return err
	}

	times, err := f.repository.GetTime().FindAll(ctx)
	if err != nil {
		return err
	}

	numberOfDays := 30
	fieldSchedules := make([]models.FieldSchedule, 0, numberOfDays)
	now := time.Now().Add(time.Duration(1) * 24 * time.Hour)
	for i := 0; i < numberOfDays; i++ {
		currentDate := now.AddDate(0, 0, i)
		for _, item := range times {
			schedule, err := f.repository.GetFieldSchedule().FindByDateAndTimeID(ctx, currentDate.Format(time.DateOnly), int(item.ID), int(field.ID))
			if err != nil {
				return err
			}

			if schedule != nil {
				return errFieldSchedule.ErrFieldScheduleExists
			}

			fieldSchedules = append(fieldSchedules, models.FieldSchedule{
				UUID:    item.UUID,
				FieldID: field.ID,
				TimeID:  uint(item.ID),
				Date:    currentDate,
				Status:  constants.Available,
			})
		}
	}

	err = f.repository.GetFieldSchedule().Create(ctx, fieldSchedules)
	if err != nil {
		return err
	}

	return nil
}

func (f *FieldScheduleService) Update(ctx context.Context, uuid string, request *dto.UpdateFieldScheduleRequest) (*dto.FieldScheduleResponse, error) {
	fieldSchedule, err := f.repository.GetFieldSchedule().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	scheduleTime, err := f.repository.GetTime().FindByUUID(ctx, request.TimeID)
	if err != nil {
		return nil, err
	}

	isTimeExists, err := f.repository.GetFieldSchedule().FindByDateAndTimeID(ctx, request.Date, int(scheduleTime.ID), int(fieldSchedule.ID))
	if err != nil {
		return nil, err
	}

	if isTimeExists != nil && request.Date != fieldSchedule.Date.Format(time.DateOnly) {
		checkDate, err := f.repository.GetFieldSchedule().FindByDateAndTimeID(ctx, request.Date, int(scheduleTime.ID), int(fieldSchedule.ID))
		if err != nil {
			return nil, err
		}

		if checkDate != nil {
			return nil, errFieldSchedule.ErrFieldScheduleExists
		}
	}

	dateParse, _ := time.Parse(time.DateOnly, request.Date)
	fieldResult, err := f.repository.GetFieldSchedule().Update(ctx, uuid, &models.FieldSchedule{
		Date:   dateParse,
		TimeID: uint(scheduleTime.ID),
	})

	if err != nil {
		return nil, err
	}

	response := dto.FieldScheduleResponse{
		UUID:         fieldResult.UUID,
		FieldName:    fieldResult.Field.Name,
		Date:         fieldResult.Date.Format(time.DateOnly),
		PricePerHour: fieldResult.Field.PricePerHour,
		Status:       fieldSchedule.Status.GetStatusString(),
		Time:         fmt.Sprintf("%s - %s", scheduleTime.StartTime, scheduleTime.EndTime),
		CreatedAt:    fieldResult.CreatedAt,
		UpdatedAt:    fieldResult.UpdatedAt,
	}

	return &response, nil
}

func (f *FieldScheduleService) UpdateStatus(ctx context.Context, request *dto.UpdateStatusFieldScheduleRequest) error {
	for _, item := range request.FieldScheduleIDs {
		_, err := f.repository.GetFieldSchedule().FindByUUID(ctx, item)
		if err != nil {
			return err
		}
		err = f.repository.GetFieldSchedule().UpdateStatus(ctx, constants.Booked, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FieldScheduleService) Delete(ctx context.Context, uuid string) error {
	_, err := f.repository.GetFieldSchedule().FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	err = f.repository.GetFieldSchedule().Delete(ctx, uuid)
	if err != nil {
		return err
	}

	return nil
}
