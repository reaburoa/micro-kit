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

// Log 实现kratos的logger接口.
func (t *KratosLog) Log(level kratosLog.Level, keyvals ...interface{}) error {
	switch level {
	case kratosLog.LevelDebug:
		log.Debug(keyvals...)
	case kratosLog.LevelInfo:
		log.Info(keyvals...)
	case kratosLog.LevelWarn:
		log.Warn(keyvals...)
	case kratosLog.LevelError:
		log.Error(keyvals...)
	case kratosLog.LevelFatal:
		log.Fatal(keyvals...)
	}
	return nil
}
