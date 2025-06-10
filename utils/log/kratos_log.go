package log

import (
	kratosLog "github.com/go-kratos/kratos/v2/log"
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
		Debug(keyvals...)
	case kratosLog.LevelInfo:
		Info(keyvals...)
	case kratosLog.LevelWarn:
		Warn(keyvals...)
	case kratosLog.LevelError:
		Error(keyvals...)
	case kratosLog.LevelFatal:
		Fatal(keyvals...)
	}
	return nil
}
