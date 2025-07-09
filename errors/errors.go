package errors

import (
	"fmt"
)

type KitError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

func (e *KitError) Error() string {
	return fmt.Sprintf("code: %d, reason: %s, message: %s", e.Code, e.Reason, e.Message)
}

func (e *KitError) GetCode() int {
	if e != nil {
		return e.Code
	}
	return 0
}

func (e *KitError) GetMessage() string {
	if e != nil {
		return e.Message
	}
	return ""
}

func (e *KitError) GetReason() string {
	if e != nil {
		return e.Reason
	}
	return ""
}

func New(code int, message, reason string) *KitError {
	return &KitError{
		Code:    code,
		Message: message,
		Reason:  reason,
	}
}

func Errorf(code int, format string, args ...interface{}) error {
	return &KitError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func IsKitError(err error) bool {
	if _, ok := err.(*KitError); ok {
		return true
	}
	return false
}
