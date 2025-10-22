package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/matgreaves/kart-challenge/api/apperr"
)

const (
	ErrCodeValidation = "validation"
	ErrCodeInternal   = "internal"
	ErrCodeNotFound   = "not found"
	ErrCodeConstraint = "constraint"
	ErrCodeBadRequest = "bad request"
)

// ErrInternal is the default error returned if no more specific error can be matched.
var ErrInternal = ServerError{
	Code:    ErrCodeInternal,
	Message: "internal server error",
}

// ServerError represents a response to a client of a request that failed.
type ServerError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ServerError implements [error.Error]
func (se ServerError) Error() string {
	return fmt.Sprintf("%s: %s", se.Code, se.Message)
}

// StatucCode maps a s to a numeric http status code
func (s ServerError) StatusCode() int {
	switch s.Code {
	case ErrCodeValidation:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeConstraint:
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}

func appErrToServer(err error) ServerError {
	serr := ErrInternal
	var se apperr.Error

	if errors.As(err, &se) {
		switch se.Code {
		case apperr.CodeValidation:
			serr.Code = ErrCodeValidation
			serr.Message = se.Cause.Error()
		case apperr.CodeConstraint:
			serr.Code = ErrCodeConstraint
			serr.Message = se.Cause.Error()
		case apperr.CodeNotFound:
			serr.Code = ErrCodeNotFound
			// note: This doesn't quite match the behaviour of the demo server which
			// doesn't return a body at all for missing resources. To keep this endpoint
			// consistent with the behaviour of createOrder I've kept to the same error handling pattern.
			serr.Message = se.Cause.Error()
		}
	}

	return serr
}
