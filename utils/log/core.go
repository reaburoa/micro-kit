package log

import (
	"fmt"
	"os"
	"time"

	"github.com/welltop-cn/common/cloud/config"
	"github.com/welltop-cn/common/protos"
	"github.com/welltop-cn/common/utils/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(options ...zap.Option) {
	zap.ReplaceGlobals(initLog(getLoggerConf(), options...))
}

func NewLogger(conf *protos.Logger, options ...zap.Option) *zap.Logger {
	return initLog(conf, options...)
}

func initLog(conf *protos.Logger, options ...zap.Option) *zap.Logger {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(Level2ZapLevle(Level(conf.Level)))

	lumberJackLogger := getLumberJackLogger(conf)
	writerSyncer := []zapcore.WriteSyncer{
		zapcore.AddSync(os.Stdout),        // 输出到stdout
		zapcore.AddSync(lumberJackLogger), // 输出到文件
	}
	multiWriter := zapcore.NewMultiWriteSyncer(writerSyncer...)
	core := zapcore.NewCore(
		getEncoder(), // 设置编码器
		multiWriter,  // 设置日志打印方式
		atomicLevel,  // 日志级别
	)
	ops := make([]zap.Option, 0, len(options))
	ops = append(ops,
		zap.AddCaller(),   // 开启开发模式，堆栈跟踪
		zap.Development(), // 开启文件及行号
		zap.Fields(zap.String("service", env.ServiceName())), // 设置初始化字段
		zap.AddCallerSkip(1), // 默认跳过一层
	)
	ops = append(ops, options...)
	logger := zap.New(core, ops...)

	return logger
}

func getLoggerConf() *protos.Logger {
	var logC protos.Logger
	err := config.Get("logger").Scan(&logC)
	if err != nil {
		return &protos.Logger{
			Level: "info",
		}
	}
	return &logC
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.MessageKey = "body"
	encoderConfig.EncodeName = zapcore.FullNameEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.NameKey = ""
	return zapcore.NewJSONEncoder(encoderConfig)
}

// 使用 lumberjack 库设置log归档、切分
func getLumberJackLogger(logConf *protos.Logger) *lumberjack.Logger {
	jackLogger := &lumberjack.Logger{
		LocalTime: true,
	}
	logFilename := fmt.Sprintf("%s.%s", "logs/app.log", time.Now().Format("20060102"))
	if logConf.Path != "" && logConf.Filename != "" {
		logFilename = fmt.Sprintf("%s/%s.%s", logConf.Path, logConf.Filename, time.Now().Format("20060102"))
	}
	jackLogger.Filename = logFilename

	if logConf.MaxSize > 0 {
		jackLogger.MaxSize = int(logConf.MaxSize)
	}
	if logConf.MaxAge > 0 {
		jackLogger.MaxAge = int(logConf.MaxAge)
	}
	if logConf.BackupNums > 0 {
		jackLogger.MaxBackups = int(logConf.BackupNums)
	}
	if logConf.IsCompress {
		jackLogger.Compress = logConf.IsCompress
	}

	return jackLogger
}
