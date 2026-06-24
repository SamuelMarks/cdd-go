package models

// AppError represents a standard application error with an HTTP status code and a custom message.
type AppError struct {
	Message string
	Code    int
}

// Error returns the custom message, implementing the standard error interface.
func (e *AppError) Error() string { return e.Message }

// NewNotImplementedError creates a new AppError with a 501 Not Implemented status code.
func NewNotImplementedError() error { return &AppError{Message: "Not implemented", Code: 501} }
