package log

import (
    "log"
    "go.uber.org/zap"
)

var Sugar *zap.SugaredLogger

func init() {
    // Setup logging
    loggerInstance, err := NewLogger("info")
    if err != nil {
        log.Fatalf("Error setting up logger: %v", err)
    }
    defer loggerInstance.Sync()
    Sugar = loggerInstance.Sugar()
}

func NewLogger(level string) (*zap.Logger, error) {
    var cfg zap.Config
    if level == "production" {
        cfg = zap.NewProductionConfig()
    } else {
        cfg = zap.NewDevelopmentConfig()
    }
    logger, err := cfg.Build()
    if err != nil {
        return nil, err
    }
    return logger, nil
}
