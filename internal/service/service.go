package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"green-api/internal/greenapi"
	"green-api/internal/model"
)

type GreenAPIClient interface {
	GetSettings(ctx context.Context, idInstance, apiTokenInstance string) (greenapi.Response, error)
	GetStateInstance(ctx context.Context, idInstance, apiTokenInstance string) (greenapi.Response, error)
	SendMessage(ctx context.Context, idInstance, apiTokenInstance, chatID, message string) (greenapi.Response, error)
	SendFileByURL(ctx context.Context, idInstance, apiTokenInstance, chatID, urlFile, fileName string) (greenapi.Response, error)
}

type Service struct {
	client   GreenAPIClient
	validate *validator.Validate
}

type CredentialsRequest struct {
	IDInstance       string `json:"idInstance" validate:"required"`
	APITokenInstance string `json:"apiTokenInstance" validate:"required"`
}

type SendMessageRequest struct {
	CredentialsRequest
	ChatID  string `json:"chatId" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type SendFileByURLRequest struct {
	CredentialsRequest
	ChatID  string `json:"chatId" validate:"required"`
	URLFile string `json:"urlFile" validate:"required,url"`
}

var digitsOnly = regexp.MustCompile(`^\d+$`)

func New(client GreenAPIClient) *Service {
	validate := validator.New()
	return &Service{client: client, validate: validate}
}

func (s *Service) GetSettings(ctx context.Context, req CredentialsRequest) (greenapi.Response, *model.APIError) {
	if err := s.validateCredentials(req); err != nil {
		return greenapi.Response{}, err
	}

	resp, callErr := s.client.GetSettings(ctx, strings.TrimSpace(req.IDInstance), strings.TrimSpace(req.APITokenInstance))
	if callErr != nil {
		return greenapi.Response{}, mapUpstreamError(callErr)
	}
	return resp, nil
}

func (s *Service) GetState(ctx context.Context, req CredentialsRequest) (greenapi.Response, *model.APIError) {
	if err := s.validateCredentials(req); err != nil {
		return greenapi.Response{}, err
	}

	resp, callErr := s.client.GetStateInstance(ctx, strings.TrimSpace(req.IDInstance), strings.TrimSpace(req.APITokenInstance))
	if callErr != nil {
		return greenapi.Response{}, mapUpstreamError(callErr)
	}
	return resp, nil
}

func (s *Service) SendMessage(ctx context.Context, req SendMessageRequest) (greenapi.Response, *model.APIError) {
	if err := s.validate.Struct(req); err != nil {
		return greenapi.Response{}, validationError(err)
	}
	if err := s.validateCredentials(req.CredentialsRequest); err != nil {
		return greenapi.Response{}, err
	}

	normalizedChatID, err := NormalizeChatID(req.ChatID)
	if err != nil {
		return greenapi.Response{}, invalidInput("chatId", err.Error())
	}

	resp, callErr := s.client.SendMessage(
		ctx,
		strings.TrimSpace(req.IDInstance),
		strings.TrimSpace(req.APITokenInstance),
		normalizedChatID,
		strings.TrimSpace(req.Message),
	)
	if callErr != nil {
		return greenapi.Response{}, mapUpstreamError(callErr)
	}
	return resp, nil
}

func (s *Service) SendFileByURL(ctx context.Context, req SendFileByURLRequest) (greenapi.Response, *model.APIError) {
	if err := s.validate.Struct(req); err != nil {
		return greenapi.Response{}, validationError(err)
	}
	if err := s.validateCredentials(req.CredentialsRequest); err != nil {
		return greenapi.Response{}, err
	}

	normalizedChatID, err := NormalizeChatID(req.ChatID)
	if err != nil {
		return greenapi.Response{}, invalidInput("chatId", err.Error())
	}

	urlFile := strings.TrimSpace(req.URLFile)
	if err := validateURLFile(urlFile); err != nil {
		return greenapi.Response{}, invalidInput("urlFile", err.Error())
	}

	fileName, err := ExtractFileName(urlFile)
	if err != nil {
		return greenapi.Response{}, invalidInput("urlFile", err.Error())
	}

	resp, callErr := s.client.SendFileByURL(
		ctx,
		strings.TrimSpace(req.IDInstance),
		strings.TrimSpace(req.APITokenInstance),
		normalizedChatID,
		urlFile,
		fileName,
	)
	if callErr != nil {
		return greenapi.Response{}, mapUpstreamError(callErr)
	}
	return resp, nil
}

func (s *Service) validateCredentials(req CredentialsRequest) *model.APIError {
	if err := s.validate.Struct(req); err != nil {
		return validationError(err)
	}
	return nil
}

func NormalizeChatID(raw string) (string, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return "", fmt.Errorf("chatId is required")
	}

	if strings.HasSuffix(candidate, "@c.us") {
		number := strings.TrimSuffix(candidate, "@c.us")
		if number == "" || !digitsOnly.MatchString(number) {
			return "", fmt.Errorf("chatId must contain only digits before @c.us")
		}
		return candidate, nil
	}

	if !digitsOnly.MatchString(candidate) {
		return "", fmt.Errorf("chatId must contain only digits")
	}

	return candidate + "@c.us", nil
}

func validateURLFile(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("urlFile must be a valid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("urlFile must use http or https")
	}
	return nil
}

func ExtractFileName(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url")
	}

	fileName := path.Base(parsed.Path)
	if fileName == "." || fileName == "/" || fileName == "" {
		return "", fmt.Errorf("cannot extract fileName from urlFile")
	}
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		return "", fmt.Errorf("invalid fileName extracted from urlFile")
	}

	decoded, unescapeErr := url.PathUnescape(fileName)
	if unescapeErr == nil {
		fileName = decoded
	}

	if len(fileName) > 255 {
		return "", fmt.Errorf("fileName is too long")
	}
	return fileName, nil
}

func validationError(err error) *model.APIError {
	return &model.APIError{
		StatusCode: 400,
		Code:       "validation_error",
		Message:    "invalid request payload",
		Details:    err.Error(),
	}
}

func invalidInput(field, message string) *model.APIError {
	return &model.APIError{
		StatusCode: 400,
		Code:       "validation_error",
		Message:    "invalid request payload",
		Details: map[string]string{
			"field":   field,
			"message": message,
		},
	}
}

func mapUpstreamError(err error) *model.APIError {
	statusCode := 502
	if errors.Is(err, context.DeadlineExceeded) {
		statusCode = 504
	}
	if errors.Is(err, greenapi.ErrCircuitBreakerOpen) {
		statusCode = 503
	}

	return &model.APIError{
		StatusCode: statusCode,
		Code:       "upstream_error",
		Message:    err.Error(),
	}
}
