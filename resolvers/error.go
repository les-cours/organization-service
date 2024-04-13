package resolvers

import (
	"errors"
	"fmt"
)

func ErrNotFound(resource string) error {
	errNotFound := fmt.Sprintf("resource %s not found", resource)
	return errors.New(errNotFound)
}

func ErrInvalidInput(field string, reason string) error {
	errInvalidInput := fmt.Sprintf("invalid input provided for field %s: %s", field, reason)
	return errors.New(errInvalidInput)
}

func ErrExistInput(field string) error {
	return ErrInvalidInput(field, " already exists.")
}

func Err(err string) error {
	return errors.New(err)
}

var ErrInternal = errors.New("internal error, please try again later")
