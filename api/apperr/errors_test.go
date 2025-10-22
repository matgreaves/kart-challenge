package apperr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	cause := errors.New("hit the fan")
	code := CodeValidation
	s := Error{Code: code, Cause: cause}.Error()
	assert.Equal(t, string(code)+": "+cause.Error(), s)
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("hit the fan")
	assert.Equal(t, cause, Error{Cause: cause}.Unwrap())
}
