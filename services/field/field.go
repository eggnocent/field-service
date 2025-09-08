package services

import (
	"bytes"
	"context"
	"field-service/common/gcs"
	"field-service/common/util"
	errConstant "field-service/constants/error"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FieldService struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSlient
}

type IFieldService interface {
	GetAllWithPagination(context.Context, *dto.FieldRequestParam) (*util.PaginationResult, error)
	GetAllWithoutPagination(context.Context) ([]dto.FieldResponse, error)
	GetByUUID(context.Context, string) (*dto.FieldResponse, error)
	Create(context.Context, *dto.FieldRequest) (*dto.FieldResponse, error)
	Update(context.Context, string, *dto.UpdateFieldRequest) (*dto.FieldResponse, error)
	Delete(context.Context, string) error
}

func NewFieldService(repository repositories.IRepositoryRegistry, gcs gcs.IGCSlient) IFieldService {
	return &FieldService{repository: repository, gcs: gcs}
}

func (f *FieldService) GetAllWithPagination(ctx context.Context, param *dto.FieldRequestParam) (*util.PaginationResult, error) {
	fields, total, err := f.repository.GetField().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	fieldResults := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResults = append(fieldResults, dto.FieldResponse{
			UUID:         field.UUID,
			Code:         field.Code,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
			CreatedAt:    field.CreatedAt,
			UpdatedAt:    field.UpdatedAt,
		})
	}

	pagination := &util.PaginationParams{
		Count: total,
		Page:  param.Page,
		Limit: param.Limit,
		Data:  fieldResults,
	}

	response := util.GeneratePagination(*pagination)
	return &response, nil
}

func (f *FieldService) GetAllWithoutPagination(ctx context.Context) ([]dto.FieldResponse, error) {
	fields, err := f.repository.GetField().FindAllWithoutPagination(ctx)
	if err != nil {
		return nil, err
	}

	fieldResults := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResults = append(fieldResults, dto.FieldResponse{
			UUID:         field.UUID,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
		})
	}

	return fieldResults, nil
}

func (f *FieldService) GetByUUID(ctx context.Context, uuid string) (*dto.FieldResponse, error) {
	field, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	fieldResult := dto.FieldResponse{
		UUID:         field.UUID,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}

	return &fieldResult, nil
}

func (f *FieldService) validateUpload(images []multipart.FileHeader) error {
	logrus.Infof("VALIDATE_UPLOAD: received %d image(s)", len(images))

	if len(images) == 0 {
		logrus.Warn("VALIDATE_UPLOAD: images is empty")
		return errConstant.ErrInvalidUploadFile
	}

	for i, image := range images {
		logrus.Infof("VALIDATE_UPLOAD: image[%d] = %+v", i, image)

		// ✅ PERLINDUNGAN TAMBAHAN — meskipun ini bukan pointer
		if reflect.ValueOf(image).IsZero() {
			logrus.Warnf("VALIDATE_UPLOAD: image[%d] is zero value", i)
			return errConstant.ErrInvalidUploadFile
		}

		if image.Filename == "" {
			logrus.Warnf("VALIDATE_UPLOAD: image[%d] has empty filename", i)
			return errConstant.ErrInvalidUploadFile
		}

		if image.Size > 5*1024*1024 {
			logrus.Warnf("VALIDATE_UPLOAD: image[%d] too large: %d bytes", i, image.Size)
			return errConstant.ErrSizeTooLarge
		}
	}
	return nil
}

func (f *FieldService) processAndUploadImage(ctx context.Context, image multipart.FileHeader) (string, error) {
	logrus.Infof("PROCESS_UPLOAD: starting for image %+v", image)

	if image.Filename == "" {
		logrus.Warn("PROCESS_UPLOAD: image filename is empty")
		return "", errConstant.ErrInvalidUploadFile
	}

	file, err := image.Open()
	if err != nil {
		logrus.Errorf("PROCESS_UPLOAD: failed to open image: %v", err)
		return "", err
	}
	defer file.Close()

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
	if err != nil {
		logrus.Errorf("PROCESS_UPLOAD: failed to copy file to buffer: %v", err)
		return "", err
	}

	ext := strings.ToLower(path.Ext(image.Filename))
	name := strings.TrimSuffix(image.Filename, ext)
	safeName := strings.ReplaceAll(name, " ", "-")

	filename := fmt.Sprintf("images/%s-%s%s", time.Now().Format("20060102150405"), safeName, ext)
	logrus.Infof("PROCESS_UPLOAD: uploading file with name %s", filename)

	url, err := f.gcs.UploadFile(ctx, filename, buffer.Bytes())
	if err != nil {
		logrus.Errorf("PROCESS_UPLOAD: failed to upload file to GCS: %v", err)
		return "", err
	}

	logrus.Infof("PROCESS_UPLOAD: uploaded file successfully: %s", url)
	return url, nil
}

func (f *FieldService) uploadImage(ctx context.Context, image []multipart.FileHeader) ([]string, error) {
	logrus.Infof("UPLOAD_IMAGE: received %d image(s)", len(image))

	err := f.validateUpload(image)
	if err != nil {
		logrus.Errorf("UPLOAD_IMAGE: validation failed: %v", err)
		return nil, err
	}

	urls := make([]string, 0, len(image))
	for i, file := range image {
		logrus.Infof("UPLOAD_IMAGE: processing image[%d]: %+v", i, file)

		url, err := f.processAndUploadImage(ctx, file)
		if err != nil {
			logrus.Errorf("UPLOAD_IMAGE: failed to process image[%d]: %v", i, err)
			return nil, err
		}
		logrus.Infof("UPLOAD_IMAGE: image[%d] uploaded to %s", i, url)
		urls = append(urls, url)
	}

	logrus.Infof("UPLOAD_IMAGE: all %d image(s) processed successfully", len(urls))
	return urls, nil
}

func (f *FieldService) Create(ctx context.Context, request *dto.FieldRequest) (*dto.FieldResponse, error) {
	logrus.Infof("CREATE_FIELD: received request: %+v", request)

	imageURL, err := f.uploadImage(ctx, request.Images)
	if err != nil {
		logrus.Errorf("CREATE_FIELD: failed to upload image: %v", err)
		return nil, err
	}

	logrus.Infof("CREATE_FIELD: saving field to database with image URLs: %v", imageURL)

	field, err := f.repository.GetField().Create(ctx, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		PricePerHour: request.PricePerHour,
		Images:       imageURL,
	})
	if err != nil {
		logrus.Errorf("CREATE_FIELD: failed to create field in DB: %v", err)
		return nil, err
	}

	response := &dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}

	logrus.Infof("CREATE_FIELD: field created successfully: %+v", response)
	return response, nil
}

func (f *FieldService) Update(ctx context.Context, uuidParam string, request *dto.UpdateFieldRequest) (*dto.FieldResponse, error) {
	logrus.Infof("[SERVICE] Start updating field UUID: %s", uuidParam)
	logrus.Infof("[SERVICE] Incoming request data: %+v", request)

	// 1. Cek apakah field dengan UUID tersebut ada
	field, err := f.repository.GetField().FindByUUID(ctx, uuidParam)
	if err != nil {
		logrus.Errorf("[SERVICE] Failed to find field by UUID %s: %v", uuidParam, err)
		return nil, err
	}

	// 2. Proses upload image jika ada
	var imageUrls []string
	if request.Images == nil {
		logrus.Infof("[SERVICE] No new images uploaded, using existing images")
		imageUrls = field.Images
	} else {
		logrus.Infof("[SERVICE] Uploading %d images", len(request.Images))
		imageUrls, err = f.uploadImage(ctx, request.Images)
		if err != nil {
			logrus.Errorf("[SERVICE] Failed to upload images: %v", err)
			return nil, err
		}
		logrus.Infof("[SERVICE] Uploaded image URLs: %+v", imageUrls)
	}

	// 3. Panggil repository untuk update
	logrus.Infof("[SERVICE] Calling repository.Update with Code=%s, Name=%s, Price=%.2f",
		request.Code, request.Name, request.PricePerHour)

	fieldResult, err := f.repository.GetField().Update(ctx, uuidParam, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		PricePerHour: request.PricePerHour,
		Images:       imageUrls,
	})

	if err != nil {
		logrus.Errorf("[SERVICE] Repository update failed: %v", err)
		return nil, err
	}

	logrus.Infof("[SERVICE] Update success. Result UUID: %s", fieldResult.UUID)

	uuidParse, _ := uuid.Parse(uuidParam)
	return &dto.FieldResponse{
		UUID:         uuidParse,
		Code:         fieldResult.Code,
		Name:         fieldResult.Name,
		PricePerHour: fieldResult.PricePerHour,
		Images:       fieldResult.Images,
		CreatedAt:    fieldResult.CreatedAt,
		UpdatedAt:    fieldResult.UpdatedAt,
	}, nil
}

func (f *FieldService) Delete(ctx context.Context, uuid string) error {
	_, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	err = f.repository.GetField().Delete(ctx, uuid)
	if err != nil {
		return err
	}

	return nil
}
