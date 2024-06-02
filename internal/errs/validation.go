package errs

import (
	"fmt"
)

type ValidationError struct {
	Text string
	err  error
}

func NewValidationError(text string) ValidationError {
	return ValidationError{
		Text: fmt.Sprintf("validation error: %v", text),
	}
}

func (e ValidationError) Error() string {
	return e.Text
}

func (e ValidationError) Wrap(err error) error {
	e.err = fmt.Errorf("%v: %w", e.Text, err)
	e.Text = e.err.Error()

	return e
}

func (e ValidationError) Unwrap() error {
	return e.err
}
