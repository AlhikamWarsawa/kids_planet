package utils

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	CodeBadRequest              = "BAD_REQUEST"
	CodeUnauthorized            = "UNAUTHORIZED"
	CodeForbidden               = "FORBIDDEN"
	CodeNotFound                = "NOT_FOUND"
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

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e AppError) Error() string {
	return e.Code + ": " + e.Message
}

func ErrBadRequest(msg string) AppError {
	if msg == "" {
		msg = "invalid input"
	}
	return AppError{
		Code:       CodeBadRequest,
		Message:    msg,
		HTTPStatus: http.StatusBadRequest,
	}
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
	if msg == "" {
		msg = "not found"
	}
	return AppError{
		Code:       CodeNotFound,
		Message:    msg,
		HTTPStatus: http.StatusNotFound,
	}
}

func ErrInternal() AppError {
	return AppError{
		Code:       CodeInternal,
		Message:    "internal error",
		HTTPStatus: http.StatusInternalServerError,
	}
}

func ErrRateLimited(msg string) AppError {
	if msg == "" {
		msg = "rate limited"
	}
	return AppError{
		Code:       CodeRateLimited,
		Message:    msg,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

func ErrInvalidZip(msg string) AppError {
	if msg == "" {
		msg = "invalid zip file"
	}
	return AppError{
		Code:       CodeInvalidZip,
		Message:    msg,
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
	if msg == "" {
		msg = "zip entry path is invalid"
	}
	return AppError{
		Code:       CodeInvalidZipPath,
		Message:    msg,
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
