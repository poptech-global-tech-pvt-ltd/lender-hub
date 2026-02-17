package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

type Config struct {
	Service string
	Env     string // "local", "dev", "staging", "prod"
}

func New(cfg Config) (*Logger, error) {
	level := zapcore.InfoLevel
	if cfg.Env == "local" || cfg.Env == "dev" {
		level = zapcore.DebugLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)

	base := zap.New(core).
		With(zap.String("service", cfg.Service)).
		With(zap.String("env", cfg.Env))

	return &Logger{Logger: base}, nil
}
