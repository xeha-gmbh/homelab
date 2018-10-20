package shared

var (
	ErrParse = errorFactory(1)("parse-error")
	ErrApi = errorFactory(2)("api-error")
	ErrOp = errorFactory(3)("op-error")
	ErrDependency = errorFactory(4)("dependency-error")
)

func errorFactory(code int) func(cause string) *LabError {
	return func(cause string) *LabError {
		return &LabError{
			Cause: cause,
			ExitCode: code,
		}
	}
}

type LabError struct {
	Cause 		string
	ExitCode 	int
}

func (e *LabError) Error() string {
	return e.Cause
}
