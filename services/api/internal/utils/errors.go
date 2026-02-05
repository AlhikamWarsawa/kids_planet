package utils

import (
	"fmt"
	"net/http"
)

const (
	CodeBadRequest      = "BAD_REQUEST"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeInternal        = "INTERNAL_ERROR"
	CodeInvalidZip      = "INVALID_ZIP"
	CodeZipTooLarge     = "ZIP_TOO_LARGE"
	CodeMissingIndexHTML = "MISSING_INDEX_HTML"
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

func ErrMissingIndexHTML() AppError {
	return AppError{
		Code:       CodeMissingIndexHTML,
		Message:    "index.html must exist at zip root",
		HTTPStatus: http.StatusBadRequest,
	}
}
