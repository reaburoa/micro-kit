package errors

import (
	"fmt"
)

type KitError struct {
	code    int    `json:"code"`
	message string `json:"message"`
	reason  string `json:"reason"`
}

func (e *KitError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.code, e.message)
}

func (e *KitError) Code() int {
	return e.code
}

func (e *KitError) Message() string {
	return e.message
}

func (e *KitError) Reason() string {
	return e.reason
}

func New(code int, message, reason string) *KitError {
	return &KitError{
		code:    code,
		message: message,
		reason:  reason,
	}
}

func Errorf(code int, format string, args ...interface{}) error {
	return &KitError{
		code:    code,
		message: fmt.Sprintf(format, args...),
	}
}

func IsKitError(err error) bool {
	if _, ok := err.(*KitError); ok {
		return true
	}
	return false
}
