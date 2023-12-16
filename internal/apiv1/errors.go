package apiv1

import (
	"fmt"
	"net/http"
)

const (
	invalidRequestCode      = 100
	internalServerErrorCode = 101

	ErrorLogLevelInfo    = "info"
	ErrorLogLevelWarning = "warning"
	ErrorLogLevelError   = "error"
)

type Error struct {
	statusCode int
	message    string
	code       int
	err        error
	logLevel   string
}

func (e *Error) StatusCode() int {
	return e.statusCode
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Error() string {
	if e.err == nil {
		return e.message
	}

	return fmt.Sprintf("%s, err: %s", e.message, e.err)
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) LogLevel() string {
	return e.logLevel
}

func NewInvalidRequestError(text string, err error) *Error {
	return &Error{
		statusCode: http.StatusBadRequest,
		message:    text,
		code:       invalidRequestCode,
		err:        err,
		logLevel:   ErrorLogLevelInfo,
	}
}

func NewInvalidRequestErrorInvalidParameter(param string, err error) *Error {
	return NewInvalidRequestError(fmt.Sprintf("invalid request parameter %q", param), err)
}

func NewInvalidRequestErrorMissingRequiredParameter(param string) *Error {
	return NewInvalidRequestError(fmt.Sprintf("required parameter %q is missing", param), nil)
}

func NewServerError(err error) *Error {
	return &Error{
		statusCode: http.StatusInternalServerError,
		message:    "internal server error",
		code:       internalServerErrorCode,
		err:        err,
		logLevel:   ErrorLogLevelError,
	}
}