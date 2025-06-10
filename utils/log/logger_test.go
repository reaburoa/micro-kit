package log

import (
	"context"
	"testing"

	"github.com/welltop-cn/common/cloud/config"
)

func Test_logger(t *testing.T) {
	config.InitConfig()
	InitLogger()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "ttttttttt")
	ctx = context.WithValue(ctx, AttrRequestId, "request_id333")

	//Infow("abc", "this is a debug log", "1222")

	CtxInfof(ctx, "abc %s", "this is a debug log")

}
