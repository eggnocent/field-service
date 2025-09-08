package controllers

import (
	"field-service/common/response"
	"field-service/domain/dto"
	"field-service/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type TimeController struct {
	service services.IServiceRegistry
}

type ITimeController interface {
	GetAll(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
}

func NewTimeController(service services.IServiceRegistry) ITimeController {
	return &TimeController{service: service}
}

func (t *TimeController) GetAll(c *gin.Context) {
	result, err := t.service.GetTime().GetAll(c)
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

func (t *TimeController) GetByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	result, err := t.service.GetTime().GetByUUID(c, uuid)
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

func (t *TimeController) Create(c *gin.Context) {
	logrus.Info("[CONTROLLER] Start creating new time slot")

	var request dto.TimeRequest

	// 1. Binding JSON
	logrus.Info("[CONTROLLER] Binding JSON body to dto.TimeRequest")
	err := c.ShouldBindJSON(&request)
	if err != nil {
		logrus.Errorf("[CONTROLLER] Failed to bind JSON: %v", err)
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	logrus.Infof("[CONTROLLER] JSON binding successful. Request: %+v", request)

	// 2. Validasi struct
	logrus.Info("[CONTROLLER] Validating request struct")
	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
		logrus.Errorf("[CONTROLLER] Validation failed: %v", err)
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	logrus.Info("[CONTROLLER] Validation successful")

	// 3. Panggil service untuk simpan
	logrus.Info("[CONTROLLER] Calling service.GetTime().Create()")
	result, err := t.service.GetTime().Create(c, &request)
	if err != nil {
		logrus.Errorf("[CONTROLLER] Service failed to create time: %v", err)
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	logrus.Infof("[CONTROLLER] Time successfully created. Response: %+v", result)

	// 4. Kirim response OK
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}
