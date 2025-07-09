package kratos

import (
	kratosLog "github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
)

type KratosLog struct {
	logger *zap.Logger
}

func NewKratosLog() *KratosLog {
	return &KratosLog{
		logger: zap.S().Desugar().WithOptions(zap.AddCallerSkip(-2)),
	}
}

const (
	kratosLogMsgKey = "kratos-log"
)

// Log 实现kratos的logger接口.
func (t *KratosLog) Log(level kratosLog.Level, keyvals ...interface{}) error {
	switch level {
	case kratosLog.LevelDebug:
		t.logger.Sugar().Debugw(kratosLogMsgKey, keyvals...)
	case kratosLog.LevelInfo:
		t.logger.Sugar().Infow(kratosLogMsgKey, keyvals...)
	case kratosLog.LevelWarn:
		t.logger.Sugar().Warnw(kratosLogMsgKey, keyvals...)
	case kratosLog.LevelError:
		t.logger.Sugar().Errorw(kratosLogMsgKey, keyvals...)
	case kratosLog.LevelFatal:
		t.logger.Sugar().Fatalw(kratosLogMsgKey, keyvals...)
	}
	return nil
}
