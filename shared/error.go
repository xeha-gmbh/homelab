package shared

var (
	ErrParse      = ErrorFactory(1)("parse-error")
	ErrApi        = ErrorFactory(2)("api-error")
	ErrOp         = ErrorFactory(3)("op-error")
	ErrDependency = ErrorFactory(4)("dependency-error")
)

func ErrorFactory(code int) func(cause string) *LabError {
	return func(cause string) *LabError {
		return &LabError{
			Cause:    cause,
			ExitCode: code,
		}
	}
}

type LabError struct {
	Cause    string
	ExitCode int
}

func (e *LabError) Error() string {
	return e.Cause
}
