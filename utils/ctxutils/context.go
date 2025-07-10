package ctxutils

type ContextKey string

const (
	CtxUserIpKey ContextKey = "UserClientIp"
	CtxUserIdKey ContextKey = "AuthUerId"
)
