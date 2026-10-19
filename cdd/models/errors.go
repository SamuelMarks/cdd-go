// Package models provides shared domain error and model types for generated code and CDD orchestration.
package models

import "fmt"

// ErrorKind categorizes domain, CLI, parser, and protocol errors across cdd-go.
type ErrorKind string

const (
	// ErrKindNotImplemented indicates an unimplemented stub operation (HTTP 501).
	ErrKindNotImplemented ErrorKind = "not_implemented"
	// ErrKindInvalidArgument indicates invalid CLI flags, tool parameters, or configuration.
	ErrKindInvalidArgument ErrorKind = "invalid_argument"
	// ErrKindSpecValidation indicates an OpenAPI 3.2.0, Swagger 2.0, or Arazzo 1.1.0 parse/validation failure.
	ErrKindSpecValidation ErrorKind = "spec_validation"
	// ErrKindASTUnsupported indicates an unsupported Go AST node or schema construct.
	ErrKindASTUnsupported ErrorKind = "ast_unsupported"
	// ErrKindCodegenIO indicates a file I/O or code generation failure.
	ErrKindCodegenIO ErrorKind = "codegen_io"
	// ErrKindProtocol indicates an MCP or JSON-RPC 2.0 protocol error.
	ErrKindProtocol ErrorKind = "protocol"
)

// AppError represents a standard application and domain error with an ErrorKind, HTTP/RPC status code, custom message, and optional underlying cause.
type AppError struct {
	// Kind is the structured domain error category.
	Kind ErrorKind
	// Message is the human-readable error message.
	Message string
	// Code is the HTTP status code or JSON-RPC code associated with the error.
	Code int
	// Cause is the optional underlying wrapped error.
	Cause error
}

// Error returns the formatted error message, implementing the standard error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause error for errors.Is and errors.As compatibility.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewNotImplementedError creates a new AppError with a 501 Not Implemented status code.
// It returns the initialized error.
func NewNotImplementedError() error {
	return &AppError{Kind: ErrKindNotImplemented, Message: "Not implemented", Code: 501}
}

// NewDomainError creates a new structured AppError for a given ErrorKind, message, and optional underlying cause.
// Parameter kind is the ErrorKind category, code is the numeric status code, msg is the description, and cause is the wrapped error.
func NewDomainError(kind ErrorKind, code int, msg string, cause error) *AppError {
	return &AppError{
		Kind:    kind,
		Message: msg,
		Code:    code,
		Cause:   cause,
	}
}
