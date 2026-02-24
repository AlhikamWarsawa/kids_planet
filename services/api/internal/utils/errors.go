package utils

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	RequestIDHeader   = "X-Request-ID"
	RequestIDLocalKey = "request_id"

	CodeBadRequest              = "BAD_REQUEST"
	CodeUnauthorized            = "UNAUTHORIZED"
	CodeForbidden               = "FORBIDDEN"
	CodeResourceNotFound        = "RESOURCE_NOT_FOUND"
	CodeNotFound                = CodeResourceNotFound
	CodeInternal                = "INTERNAL_ERROR"
	CodeRateLimited             = "RATE_LIMITED"
	CodeInvalidZip              = "INVALID_ZIP"
	CodeZipTooLarge             = "ZIP_TOO_LARGE"
	CodeInvalidZipPath          = "INVALID_ZIP_PATH"
	CodeZipTooLargeUncompressed = "ZIP_TOO_LARGE_UNCOMPRESSED"
	CodeZipTooManyFiles         = "ZIP_TOO_MANY_FILES"
	CodeInvalidFileType         = "INVALID_FILE_TYPE"
	CodeMissingIndexHTML        = "MISSING_INDEX_HTML"
)

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id,omitempty"`
	HTTPStatus int    `json:"-"`
}

type AppError = APIError

func (e APIError) Error() string {
	return e.Code + ": " + e.Message
}

func normalizeMessage(msg string, fallback string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return fallback
	}
	return msg
}

func NewBadRequest(msg string) APIError {
	return APIError{
		Code:       CodeBadRequest,
		Message:    normalizeMessage(msg, "invalid input"),
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewNotFound(msg string) APIError {
	return APIError{
		Code:       CodeResourceNotFound,
		Message:    normalizeMessage(msg, "not found"),
		HTTPStatus: http.StatusNotFound,
	}
}

func NewInternal(msg string) APIError {
	return APIError{
		Code:       CodeInternal,
		Message:    normalizeMessage(msg, "internal error"),
		HTTPStatus: http.StatusInternalServerError,
	}
}

func ErrBadRequest(msg string) AppError {
	return NewBadRequest(msg)
}

func ErrUnauthorized() AppError {
	return AppError{
		Code:       CodeUnauthorized,
		Message:    "unauthorized",
		HTTPStatus: http.StatusUnauthorized,
	}
}

func ErrForbidden() AppError {
	return AppError{
		Code:       CodeForbidden,
		Message:    "forbidden",
		HTTPStatus: http.StatusForbidden,
	}
}

func ErrNotFound(msg string) AppError {
	return NewNotFound(msg)
}

func ErrInternal() AppError {
	return NewInternal("")
}

func ErrRateLimited(msg string) AppError {
	return AppError{
		Code:       CodeRateLimited,
		Message:    normalizeMessage(msg, "rate limited"),
		HTTPStatus: http.StatusTooManyRequests,
	}
}

func ErrInvalidZip(msg string) AppError {
	return AppError{
		Code:       CodeInvalidZip,
		Message:    normalizeMessage(msg, "invalid zip file"),
		HTTPStatus: http.StatusBadRequest,
	}
}

func ErrZipTooLarge(maxBytes int64) AppError {
	msg := "zip too large"
	if maxBytes > 0 {
		msg = fmt.Sprintf("zip too large (max %d bytes)", maxBytes)
	}
	return AppError{
		Code:       CodeZipTooLarge,
		Message:    msg,
		HTTPStatus: http.StatusRequestEntityTooLarge,
	}
}

func ErrInvalidZipPath(msg string) AppError {
	return AppError{
		Code:       CodeInvalidZipPath,
		Message:    normalizeMessage(msg, "zip entry path is invalid"),
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

func ErrZipTooLargeUncompressed(maxBytes int64) AppError {
	msg := "zip too large after extraction"
	if maxBytes > 0 {
		msg = fmt.Sprintf("zip too large after extraction (max %d bytes)", maxBytes)
	}
	return AppError{
		Code:       CodeZipTooLargeUncompressed,
		Message:    msg,
		HTTPStatus: http.StatusBadRequest,
	}
}

func ErrZipTooManyFiles(maxFiles int) AppError {
	msg := "zip contains too many files"
	if maxFiles > 0 {
		msg = fmt.Sprintf("zip contains too many files (max %d)", maxFiles)
	}
	return AppError{
		Code:       CodeZipTooManyFiles,
		Message:    msg,
		HTTPStatus: http.StatusBadRequest,
	}
}

func ErrInvalidFileType(ext string) AppError {
	ext = strings.TrimSpace(strings.ToLower(ext))
	if ext == "" {
		ext = "unknown"
	}
	return AppError{
		Code:       CodeInvalidFileType,
		Message:    fmt.Sprintf("invalid file type: %s", ext),
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

func ErrMissingIndexHTML() AppError {
	return AppError{
		Code:       CodeMissingIndexHTML,
		Message:    "index.html must exist at zip root",
		HTTPStatus: http.StatusBadRequest,
	}
}

func RequestIDFromContext(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}

	switch v := c.Locals(RequestIDLocalKey).(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	default:
		return ""
	}
}

func WithRequestID(c *fiber.Ctx, appErr APIError) APIError {
	if strings.TrimSpace(appErr.RequestID) != "" {
		return appErr
	}
	appErr.RequestID = RequestIDFromContext(c)
	return appErr
}

func statusFromCode(code string) int {
	switch strings.TrimSpace(code) {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeResourceNotFound:
		return http.StatusNotFound
	case CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func WriteError(c *fiber.Ctx, appErr APIError) error {
	if c == nil {
		return nil
	}

	if strings.TrimSpace(appErr.Code) == "" {
		appErr = NewInternal("")
	}
	if strings.TrimSpace(appErr.Message) == "" {
		if appErr.HTTPStatus >= http.StatusInternalServerError {
			appErr.Message = "internal error"
		} else {
			appErr.Message = "request failed"
		}
	}
	if appErr.HTTPStatus == 0 {
		appErr.HTTPStatus = statusFromCode(appErr.Code)
	}

	appErr = WithRequestID(c, appErr)
	if appErr.RequestID != "" {
		c.Set(RequestIDHeader, appErr.RequestID)
	}

	level := "warn"
	if appErr.HTTPStatus >= http.StatusInternalServerError {
		level = "error"
	}
	if appErr.HTTPStatus < http.StatusBadRequest {
		level = "info"
	}

	requestID := appErr.RequestID
	if requestID == "" {
		requestID = "-"
	}

	log.Printf(
		"level=%s request_id=%s method=%s path=%s status=%d code=%s msg=%q",
		level,
		requestID,
		c.Method(),
		c.OriginalURL(),
		appErr.HTTPStatus,
		appErr.Code,
		appErr.Message,
	)

	return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
		"error": fiber.Map{
			"code":       appErr.Code,
			"message":    appErr.Message,
			"request_id": appErr.RequestID,
		},
	})
}

func BadRequest(c *fiber.Ctx, msg string) error {
	return WriteError(c, NewBadRequest(msg))
}

func NotFound(c *fiber.Ctx, msg string) error {
	return WriteError(c, NewNotFound(msg))
}

func Internal(c *fiber.Ctx, msg string) error {
	return WriteError(c, NewInternal(msg))
}
