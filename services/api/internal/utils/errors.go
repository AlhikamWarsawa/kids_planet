package utils

import "net/http"

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
		Code:       "BAD_REQUEST",
		Message:    msg,
		HTTPStatus: http.StatusBadRequest,
	}
}

func ErrUnauthorized() AppError {
	return AppError{
		Code:       "UNAUTHORIZED",
		Message:    "unauthorized",
		HTTPStatus: http.StatusUnauthorized,
	}
}

func ErrForbidden() AppError {
	return AppError{
		Code:       "FORBIDDEN",
		Message:    "forbidden",
		HTTPStatus: http.StatusForbidden,
	}
}

func ErrNotFound(msg string) AppError {
	if msg == "" {
		msg = "not found"
	}
	return AppError{
		Code:       "NOT_FOUND",
		Message:    msg,
		HTTPStatus: http.StatusNotFound,
	}
}

func ErrInternal() AppError {
	return AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}
}
