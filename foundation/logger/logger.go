// Package logger provides a convenience function to constructing a logger
// for use. This is required not just for applications but for testing.
package logger

import (
	"github.com/GLCharge/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// New constructs a Sugared Logger that writes to stdout and
// provides human-readable timestamps.
func New(logLevel string) (*otelzap.Logger, error) {
	level := zapcore.InfoLevel

	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	stdout := zapcore.Lock(os.Stdout)
	stderr := zapcore.Lock(os.Stderr)

	stdoutLevelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= level && l < zapcore.ErrorLevel
	})
	stderrLevelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= level && l >= zapcore.ErrorLevel
	})

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, stdout, stdoutLevelEnabler),
		zapcore.NewCore(encoder, stderr, stderrLevelEnabler),
	)

	logger := otelzap.New(
		zap.New(core),
		otelzap.WithTraceIDField(true),
		otelzap.WithMinLevel(level),
	)

	return logger, nil
}
