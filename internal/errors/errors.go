package customerrors

import (
	"fmt"
	"net/http"
)

func NewCustomError(err error, statusCode int) error {
	return &CustomError{
		Err:        err,
		StatusCode: statusCode,
	}
}

func ParseError(err error) int {
	switch e := err.(type) {
	case *CustomError:
		return e.StatusCode
	default:
		return http.StatusInternalServerError
	}
}

type CustomError struct {
	Err        error
	StatusCode int
}

func (err *CustomError) Error() string {
	return fmt.Sprintf("%v", err.Err)
}

func (err *CustomError) Unwrap() error {
	return err.Err
}
