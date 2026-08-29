package errors

import (
	stderrors "errors"
	"fmt"
)

type KitError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
	Cause   error  `json:"-"`
}

func (e *KitError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return fmt.Sprintf("code: %d, reason: %s, message: %s, caused_by: %v", e.Code, e.Reason, e.Message, e.Cause)
	}
	return fmt.Sprintf("code: %d, reason: %s, message: %s", e.Code, e.Reason, e.Message)
}

func (e *KitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
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

func Wrap(err error, code int, reason, message string) error {
	if err == nil {
		return New(code, message, reason)
	}
	return &KitError{
		Code:    code,
		Message: message,
		Reason:  reason,
		Cause:   err,
	}
}

func Errorf(code int, format string, args ...interface{}) error {
	return &KitError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func IsKitError(err error) bool {
	if err == nil {
		return false
	}
	var kitErr *KitError
	return stderrors.As(err, &kitErr)
}
