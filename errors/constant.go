package errors

var (
	InternalError = New(500, "Internal Server Error", "Internal Server Error")
)
