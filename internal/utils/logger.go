package utils

import (
	"Sparrow/configs"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() error {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(configs.LogLevel)); err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout", "logs/sparrow.log"}
	config.ErrorOutputPaths = []string{"stderr", "logs/zap.log"}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

func SyncLogger() {
	if Log != nil {
		_ = Log.Sync()
	}
}
