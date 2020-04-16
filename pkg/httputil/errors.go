package httputil

import (
	"errors"
	"fmt"
)

var ErrOverflow = errors.New("overflow")

type FieldError struct {
	Name string
	Err  error
}

func NewFieldError(name string, err error) *FieldError {
	return &FieldError{Name: name, Err: err}
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("field %s: %v", e.Name, e.Err)
}

func (e *FieldError) Unwrap() error {
	return e.Err
}
