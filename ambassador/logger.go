package main

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger returns a logger
//
// Valid levels are:
//   0 = errors only,
//   1 = include warnings,
//   2 = include informational messages,
//   3 = include debug messages.
func NewLogger(verbosity int) *zap.SugaredLogger {
	var level zapcore.Level
	switch {
	case verbosity >= 3:
		level = zapcore.DebugLevel
	case verbosity == 2:
		level = zapcore.InfoLevel
	case verbosity == 1:
		level = zapcore.WarnLevel
	default:
		level = zapcore.ErrorLevel
	}

	var stdout io.Writer = os.Stdout
	encoder := zapcore.CapitalLevelEncoder
	if x, ok := stdout.(interface{ Fd() uintptr }); ok {
		if isatty.IsTerminal(x.Fd()) {
			encoder = zapcore.CapitalColorLevelEncoder
		}
	}
	cfg := zapcore.EncoderConfig{
		LevelKey:       "level",
		MessageKey:     "msg",
		NameKey:        "logger",
		TimeKey:        "timestamp",
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeLevel:    encoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(stdout),
		level,
	)

	return zap.New(core).Sugar()
}
