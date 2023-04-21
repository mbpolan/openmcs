package logger

import (
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

type Options struct {
	LogLevel string
}

// Setup prepares the logging infrastructure for the server.
func Setup(opts Options) error {
	level, err := zap.ParseAtomicLevel(opts.LogLevel)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = level
	cfg.Encoding = "console"

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	log = logger.Sugar()
	return nil
}

// Debugf logs a debug message.
func Debugf(fmt string, args ...any) {
	log.Debugf(fmt, args)
}

// Infof logs an informational message.
func Infof(fmt string, args ...any) {
	log.Infof(fmt, args)
}

// Errorf logs an error message.
func Errorf(fmt string, args ...any) {
	log.Errorf(fmt, args)
}

// Fatalf logs a fatal error message and exits the program.
func Fatalf(fmt string, args ...any) {
	log.Fatalf(fmt, args)
}
