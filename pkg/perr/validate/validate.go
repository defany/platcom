package validate

import (
	"encoding/json"
	"errors"

	"github.com/defany/platcom/pkg/perr/codes"
	"github.com/gookit/validate"
)

type ErrorWithDetails struct {
	Code    codes.Code `json:"code"`
	Message string     `json:"message"`
	Details []string   `json:"details,omitempty"`
}

type Error struct {
	Messages []string `json:"error_messages"`
}

func NewError(messages ...string) *Error {
	return &Error{
		Messages: messages,
	}
}

func NewValidate(v any) error {
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
	})

	if r := validate.Struct(v); !r.Validate() {
		return NewError(r.Errors.String())
	}

	return nil
}

func (e *Error) addError(message string) {
	e.Messages = append(e.Messages, message)
}

func (e *Error) Error() string {
	v, err := json.Marshal(e.Messages)
	if err != nil {
		return err.Error()
	}

	return string(v)
}

func (e *Error) ErrorWithDetails() ErrorWithDetails {
	return ErrorWithDetails{
		Code:    codes.InvalidArgument,
		Message: "bad validation",
		Details: e.Messages,
	}
}

func IsValidationError(err error) bool {
	var ve *Error
	return errors.As(err, &ve)
}

func ToValidationError(err error) *Error {
	var ce *Error
	if !IsValidationError(err) {
		return nil
	}

	return ce
}
