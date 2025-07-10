package errors

import (
	kratosErr "github.com/go-kratos/kratos/v2/errors"
)

func ConvertToKratosError(err error) error {
	tmpErr := kratosErr.FromError(err)
	code := int(tmpErr.GetCode())
	reason := tmpErr.GetReason()
	message := tmpErr.GetMessage()

	if ierr, ok := err.(*KitError); ok {
		code = ierr.GetCode()
		reason = ierr.GetReason()
		message = ierr.GetMessage()
	}
	kErr := kratosErr.New(code, reason, message)
	return kErr
}

func ConvertToKitError(kerr *kratosErr.Error) (error, bool) {
	return New(int(kerr.Code), kerr.Reason, kerr.Message), true
}

func Code(err error) int {
	if ierr, ok := err.(*KitError); ok {
		return ierr.GetCode()
	}
	kerr := kratosErr.FromError(err)
	return int(kerr.Code)
}

func Message(err error) string {
	if ierr, ok := err.(*KitError); ok {
		return ierr.GetMessage()
	}
	kerr := kratosErr.FromError(err)
	return kerr.Message
}
