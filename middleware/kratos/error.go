package kratos

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/reaburoa/micro-kit/errors"
)

func ClientIErrorMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			reply, err = handler(ctx, req)
			if err != nil {
				// kratos error convert to ierror
				// if ierrors.IsIError(err) {
				// 	return
				// }
				// err, _ = ierrors.ConvertToIError(errors.FromError(err))
			}
			return
		}
	}
}

func ServerErrorMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			reply, err = handler(ctx, req)
			if err != nil {
				err = errors.ConvertToKratosError(err)
			}
			return
		}
	}
}
