package utils

// Create a zap strouctured logger that logs to a file and includes the
// callstack and the log level.

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"os"
)

// NewLogger creates a new logger

func NewLogger() *zap.Logger {

	// Create a file to write logs to

	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {

		panic(err)

	}

	// Create a zapcore encoder

	encoderConfig := zap.NewProductionEncoderConfig()

	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Create a zapcore writer

	core := zapcore.NewCore(encoder, zapcore.AddSync(f), zapcore.DebugLevel)

	// Create a zap logger

	logger := zap.New(core)

	return logger

}
