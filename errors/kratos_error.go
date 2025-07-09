package errors

import (
	kratosErr "github.com/go-kratos/kratos/v2/errors"
)

func ConvertToKratosError(err error) error {
	if !IsKitError(err) {
		return kratosErr.FromError(err)
	}
	code := 500
	reason := ""
	message := err.Error()
	if ierr, ok := err.(*KitError); ok {
		code = ierr.Code()
		reason = ierr.Reason()
		message = ierr.Message()
	}
	kErr := kratosErr.New(code, reason, message)
	return kErr
}

func ConvertToIError(kerr *kratosErr.Error) (error, bool) {
	return New(int(kerr.Code), kerr.Reason, kerr.Message), true
}

func Code(err error) int {
	if ierr, ok := err.(*KitError); ok {
		return ierr.Code()
	}
	kerr := kratosErr.FromError(err)
	return int(kerr.Code)
}

func Message(err error) string {
	if ierr, ok := err.(*KitError); ok {
		return ierr.Message()
	}
	kerr := kratosErr.FromError(err)
	return kerr.Message
}
