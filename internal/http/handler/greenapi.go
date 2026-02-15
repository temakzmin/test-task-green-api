package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"green-api/internal/model"
	"green-api/internal/service"
)

type Service interface {
	GetSettings(ctx *gin.Context, req service.CredentialsRequest) (int, []byte, string, *model.APIError)
	GetState(ctx *gin.Context, req service.CredentialsRequest) (int, []byte, string, *model.APIError)
	SendMessage(ctx *gin.Context, req service.SendMessageRequest) (int, []byte, string, *model.APIError)
	SendFileByURL(ctx *gin.Context, req service.SendFileByURLRequest) (int, []byte, string, *model.APIError)
}

type GreenAPIHandler struct {
	service *HandlerService
}

type HandlerService struct {
	core *service.Service
}

func NewGreenAPIHandler(core *service.Service) *GreenAPIHandler {
	return &GreenAPIHandler{service: &HandlerService{core: core}}
}

func (h *GreenAPIHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/settings", h.getSettings)
	router.POST("/state", h.getState)
	router.POST("/send-message", h.sendMessage)
	router.POST("/send-file-by-url", h.sendFileByURL)
}

func (h *GreenAPIHandler) getSettings(c *gin.Context) {
	var req service.CredentialsRequest
	if !bindJSON(c, &req) {
		return
	}

	resp, err := h.service.core.GetSettings(c.Request.Context(), req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	proxyResponse(c, resp.StatusCode, resp.Body, resp.ContentType)
}

func (h *GreenAPIHandler) getState(c *gin.Context) {
	var req service.CredentialsRequest
	if !bindJSON(c, &req) {
		return
	}

	resp, err := h.service.core.GetState(c.Request.Context(), req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	proxyResponse(c, resp.StatusCode, resp.Body, resp.ContentType)
}

func (h *GreenAPIHandler) sendMessage(c *gin.Context) {
	var req service.SendMessageRequest
	if !bindJSON(c, &req) {
		return
	}

	resp, err := h.service.core.SendMessage(c.Request.Context(), req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	proxyResponse(c, resp.StatusCode, resp.Body, resp.ContentType)
}

func (h *GreenAPIHandler) sendFileByURL(c *gin.Context) {
	var req service.SendFileByURLRequest
	if !bindJSON(c, &req) {
		return
	}

	resp, err := h.service.core.SendFileByURL(c.Request.Context(), req)
	if err != nil {
		writeAPIError(c, err)
		return
	}
	proxyResponse(c, resp.StatusCode, resp.Body, resp.ContentType)
}

func bindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		writeAPIError(c, &model.APIError{
			StatusCode: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "invalid JSON payload",
			Details:    err.Error(),
		})
		return false
	}
	return true
}

func proxyResponse(c *gin.Context, status int, body []byte, contentType string) {
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(status, contentType, body)
}

func writeAPIError(c *gin.Context, err *model.APIError) {
	status := err.StatusCode
	if status == 0 {
		status = http.StatusInternalServerError
	}
	c.JSON(status, model.ErrorResponse{Error: *err})
}
