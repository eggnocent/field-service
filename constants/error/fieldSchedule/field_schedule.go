package error

import "errors"

var (
	ErrFieldScheduleNotFound = errors.New("field schedule not found")
	ErrFieldScheduleExists   = errors.New("field schedule exists")
)

var FieldScheduleErrors = []error{
	ErrFieldScheduleNotFound,
	ErrFieldScheduleExists,
}
