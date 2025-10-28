package log

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(env string) *zap.Logger {
	env = strings.ToLower(env)
	cfg := zap.NewProductionConfig()

	if env == "" || env == "development" || env == "dev" {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	l, _ := cfg.Build()
	return l
}
