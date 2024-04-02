package perr

import (
	"errors"

	"github.com/defany/platcom/pkg/perr/codes"
)

type Err struct {
	msg  string
	code codes.Code
}

func New(msg string, code codes.Code) *Err {
	return &Err{msg: msg, code: code}
}

func (e *Err) Error() string {
	return e.msg
}

func (e *Err) Code() codes.Code {
	return e.code
}

func IsCommonError(err error) bool {
	var ce *Err
	return errors.As(err, &ce)
}

func ToCommonError(err error) *Err {
	var ce *Err
	if !IsCommonError(err) {
		return nil
	}

	return ce
}
