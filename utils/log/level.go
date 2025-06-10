package log

import (
	"go.uber.org/zap/zapcore"
)

type Level string

const (
	DebugLevel  Level = "debug"
	InfoLevel   Level = "info"
	WarnLevel   Level = "warn"
	ErrorLevel  Level = "error"
	DPanicLevel Level = "dpanic"
	PanicLevel  Level = "panic"
	FatalLevel  Level = "fatal"
)

var levelZapMap = map[Level]zapcore.Level{
	DebugLevel:  zapcore.DebugLevel,
	InfoLevel:   zapcore.InfoLevel,
	WarnLevel:   zapcore.WarnLevel,
	ErrorLevel:  zapcore.ErrorLevel,
	DPanicLevel: zapcore.DPanicLevel,
	PanicLevel:  zapcore.PanicLevel,
	FatalLevel:  zapcore.FatalLevel,
}

func Level2ZapLevle(l Level) zapcore.Level {
	if zapl, ok := levelZapMap[l]; ok {
		return zapl
	}
	return zapcore.InfoLevel
}
