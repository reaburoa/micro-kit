package kratos

import (
	kratosLog "github.com/go-kratos/kratos/v2/log"
	"github.com/reaburoa/micro-kit/utils/log"
)

type KratosLog struct {
}

func NewKratosLog() *KratosLog {
	return &KratosLog{}
}

const (
	kratosLogMsgKey = "kratos-log"
)

// Log 实现kratos的logger接口.
func (t *KratosLog) Log(level kratosLog.Level, keyvals ...interface{}) error {
	switch level {
	case kratosLog.LevelDebug:
		log.Debugw(kratosLogMsgKey, keyvals...)
	case kratosLog.LevelInfo:
		log.Infow(kratosLogMsgKey, keyvals...)
	case kratosLog.LevelWarn:
		log.Warnw(kratosLogMsgKey, keyvals...)
	case kratosLog.LevelError:
		log.Errorw(kratosLogMsgKey, keyvals...)
	case kratosLog.LevelFatal:
		log.Fatalw(kratosLogMsgKey, keyvals...)
	}
	return nil
}
