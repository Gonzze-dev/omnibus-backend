package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	errorsService "tesina/backend/internal/errors"
)

// LicensePlateService lee la patente de una foto usando el microservicio de OCR.
type LicensePlateService interface {
	Read(ctx context.Context, filename string, image io.Reader) (string, error)
}

type licensePlateService struct {
	httpClient *http.Client
	ocrURL     string
	apiKey     string
}

func NewLicensePlateService(httpClient *http.Client, ocrURL, apiKey string) *licensePlateService {
	return &licensePlateService{
		httpClient: httpClient,
		ocrURL:     ocrURL,
		apiKey:     apiKey,
	}
}

func (s *licensePlateService) Read(ctx context.Context, filename string, image io.Reader) (string, error) {
	if image == nil {
		return "", errorsService.ErrLicensePlateImageRequired
	}
	if filename == "" {
		filename = "image.jpg"
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errorsService.ErrUpstreamRequest, err)
	}
	if _, err := io.Copy(part, image); err != nil {
		return "", fmt.Errorf("%w: %w", errorsService.ErrUpstreamRequest, err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("%w: %w", errorsService.ErrUpstreamRequest, err)
	}

	endpoint, err := url.JoinPath(s.ocrURL, "read-license-plate")
	if err != nil {
		return "", fmt.Errorf("%w: %w", errorsService.ErrUpstreamRequest, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errorsService.ErrUpstreamRequest, err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("X-API-Key", s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errorsService.ErrUpstreamRequest, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errorsService.ErrUpstreamResponse, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusBadRequest:
		return "", errorsService.ErrLicensePlateImageInvalid
	case http.StatusUnprocessableEntity:
		return "", errorsService.ErrLicensePlateNotRead
	default:
		return "", fmt.Errorf("%w: status %d", errorsService.ErrUpstreamResponse, resp.StatusCode)
	}

	var out struct {
		LicensePlate string `json:"license_plate"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("%w: %w", errorsService.ErrUpstreamResponse, err)
	}

	plate := normalizeLicensePlate(out.LicensePlate)
	if plate == "" {
		return "", errorsService.ErrLicensePlateNotRead
	}
	return plate, nil
}
