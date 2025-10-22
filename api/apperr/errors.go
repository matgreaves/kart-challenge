package apperr

type Code string

const (
	CodeValidation Code = "validation"
	CodeNotFound   Code = "missing"
	CodeConstraint Code = "constraint"
)

type Error struct {
	Code Code
	// Cause of the error, visible to external users.
	Cause error
	// cause of the error, not visible to users, ok to log
	source error
}

// Error implements [error.Error].
func (ae Error) Error() string {
	return string(ae.Code) + ": " + ae.Cause.Error()
}

func (ae Error) Unwrap() error {
	return ae.Cause
}

func NewError(code Code, cause error) Error {
	return Error{
		Code:  code,
		Cause: cause,
	}
}
