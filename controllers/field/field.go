package controllers

import (
	"errors"
	error2 "field-service/common/error"
	"field-service/common/response"
	"field-service/domain/dto"
	fieldService "field-service/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type FieldController struct {
	service fieldService.IServiceRegistry
}

type IFieldController interface {
	GetAllWithPagination(*gin.Context)
	GetAllWithoutPagination(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Update(*gin.Context)
	Delete(*gin.Context)
}

func NewFieldController(service fieldService.IServiceRegistry) IFieldController {
	return &FieldController{service: service}
}

func (f *FieldController) GetAllWithPagination(c *gin.Context) {
	var params dto.FieldRequestParam
	err := c.ShouldBindQuery(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := error2.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusBadRequest,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     c,
		})
		return
	}

	result, err := f.service.GetField().GetAllWithPagination(c, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (f *FieldController) GetAllWithoutPagination(c *gin.Context) {
	fmt.Println("DEBUG: f.service is nil?", f.service == nil)
	fmt.Println("DEBUG: f.service.GetField() is nil?", f.service.GetField() == nil)

	result, err := f.service.GetField().GetAllWithoutPagination(c)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (f *FieldController) GetByUUID(c *gin.Context) {
	result, err := f.service.GetField().GetByUUID(c, c.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (f *FieldController) Create(c *gin.Context) {
	var request dto.FieldRequest

	// Binding form-data
	err := c.ShouldBindWith(&request, binding.FormMultipart)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// Validasi manual semua field kecuali images
	validate := validator.New()
	err = validate.Var(request.Name, "required")
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  errors.New("name is required"),
			Gin:  c,
		})
		return
	}

	err = validate.Var(request.Code, "required")
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  errors.New("code is required"),
			Gin:  c,
		})
		return
	}

	err = validate.Var(request.PricePerHour, "required")
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  errors.New("pricePerHour is required"),
			Gin:  c,
		})
		return
	}

	// ✅ Validasi manual untuk images (hindari panic)
	if len(request.Images) == 0 {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  errors.New("at least one image is required"),
			Gin:  c,
		})
		return
	}

	for i, image := range request.Images {
		if image.Filename == "" {
			response.HttpResponse(response.ParamHTTPResp{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("image[%d] is missing a filename", i),
				Gin:  c,
			})
			return
		}
	}

	// Lanjut ke service
	result, err := f.service.GetField().Create(c, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (f *FieldController) Update(c *gin.Context) {
	var request dto.UpdateFieldRequest

	logrus.Infof("[CONTROLLER] Update Field: Binding JSON request")

	err := c.ShouldBindWith(&request, binding.FormMultipart)
	if err != nil {
		logrus.Errorf("[CONTROLLER] Failed to bind JSON: %v", err)
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	logrus.Infof("[CONTROLLER] Validating request: %+v", request)

	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
		logrus.Errorf("[CONTROLLER] Validation failed: %v", err)
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := error2.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusBadRequest,
			Err:     err,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     c,
		})
		return
	}

	logrus.Infof("[CONTROLLER] Calling service update for UUID: %s", c.Param("uuid"))

	result, err := f.service.GetField().Update(c, c.Param("uuid"), &request)
	if err != nil {
		logrus.Errorf("[CONTROLLER] Service update failed: %v", err)
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	logrus.Infof("[CONTROLLER] Update successful. Result: %+v", result)

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (f *FieldController) Delete(c *gin.Context) {
	err := f.service.GetField().Delete(c, c.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  c,
	})
}
