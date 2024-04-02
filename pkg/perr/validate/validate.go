package validate

import (
	"errors"
	"strings"

	"github.com/gookit/validate"
)

type Error struct {
	Messages []string `json:"error_messages"`
}

func NewError(messages ...string) *Error {
	return &Error{
		Messages: messages,
	}
}

func Validate(v any) error {
	if r := validate.Struct(v); !r.Validate() {
		return NewError(r.Errors.String())
	}

	return nil
}

func (e *Error) addError(message string) {
	e.Messages = append(e.Messages, message)
}

func (e *Error) Error() string {
	return strings.Join(e.Messages, "\n")
}

func IsValidationError(err error) bool {
	var ve *Error
	return errors.As(err, &ve)
}
