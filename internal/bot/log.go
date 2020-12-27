package bot

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger returns a logger
func NewLogger(debug bool, format string, withTime bool) *zap.Logger {
	errorLog := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	infoLog := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if debug {
			return lvl < zapcore.ErrorLevel
		}
		return lvl < zapcore.ErrorLevel && lvl >= zapcore.InfoLevel
	})

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	var encoder func(cfg zapcore.EncoderConfig) zapcore.Encoder
	var encoderConfig func() zapcore.EncoderConfig

	if format == "console" {
		encoder = zapcore.NewConsoleEncoder
	} else {
		encoder = zapcore.NewJSONEncoder
	}

	if debug {
		encoderConfig = zap.NewDevelopmentEncoderConfig
	} else {
		encoderConfig = zap.NewProductionEncoderConfig
	}

	c := encoderConfig()
	if !withTime {
		c.TimeKey = ""
	}

	core := zapcore.NewTee(
		zapcore.NewCore(encoder(c), consoleErrors, errorLog),
		zapcore.NewCore(encoder(c), consoleDebugging, infoLog),
	)

	return zap.New(core)
}
