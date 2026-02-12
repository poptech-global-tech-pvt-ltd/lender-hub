package logging

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

// WithLogger stores logger in context
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext retrieves logger from context; returns noop if not found
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return l
	}
	noop := zap.NewNop()
	return &Logger{zap: noop}
}

// ═══════════════════════════════════════════
// Helper field constructors for common context
// ═══════════════════════════════════════════

// RequestID creates a requestId field
func RequestID(id string) zap.Field {
	return zap.String("requestId", id)
}

// UserID creates a userId field
func UserID(id string) zap.Field {
	return zap.String("userId", id)
}

// Module creates a module field
func Module(name string) zap.Field {
	return zap.String("module", name)
}

// PaymentID creates a paymentId field
func PaymentID(id string) zap.Field {
	return zap.String("paymentId", id)
}

// RefundID creates a refundId field
func RefundID(id string) zap.Field {
	return zap.String("refundId", id)
}

// Provider creates a provider field
func Provider(name string) zap.Field {
	return zap.String("provider", name)
}

// Endpoint creates an endpoint field
func Endpoint(ep string) zap.Field {
	return zap.String("endpoint", ep)
}

// DurationMs creates a durationMs field
func DurationMs(ms int64) zap.Field {
	return zap.Int64("durationMs", ms)
}

// Status creates a status field
func Status(s string) zap.Field {
	return zap.String("status", s)
}

// ErrorCode creates an errorCode field
func ErrorCode(code string) zap.Field {
	return zap.String("errorCode", code)
}

// HTTPStatus creates an httpStatus field
func HTTPStatus(code int) zap.Field {
	return zap.Int("httpStatus", code)
}
