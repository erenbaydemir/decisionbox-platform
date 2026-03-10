package errors

import "fmt"

// AppError is a structured error with HTTP status code.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true for server errors (5xx).
func (e *AppError) IsRetryable() bool {
	return e.Code >= 500
}

func NotFound(msg string) *AppError {
	return &AppError{Code: 404, Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: 401, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: 403, Message: msg}
}

func BadRequest(msg string) *AppError {
	return &AppError{Code: 400, Message: msg}
}

func Internal(msg string, err error) *AppError {
	return &AppError{Code: 500, Message: msg, Err: err}
}
