package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func infoColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.InfoLevel:
		enc.AppendString("\033[32mINFO\033[0m") // green
	default:
		enc.AppendString(level.CapitalString())
	}
}

func InitLogger() {
	if Log != nil {
		return
	}

	cfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    infoColorLevelEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(cfg)
	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.DebugLevel)

	Log = zap.New(core)
}
