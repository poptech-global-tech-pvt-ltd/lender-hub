package logger

import "go.uber.org/zap"

func RequestID(id string) zap.Field    { return zap.String("requestId", id) }
func UserID(id string) zap.Field       { return zap.String("userId", id) }
func PaymentID(id string) zap.Field    { return zap.String("paymentId", id) }
func RefundID(id string) zap.Field     { return zap.String("refundId", id) }
func OnboardingID(id string) zap.Field { return zap.String("onboardingId", id) }
func Module(name string) zap.Field     { return zap.String("module", name) }
func ErrorCode(code string) zap.Field  { return zap.String("errorCode", code) }
func HTTPStatus(code int) zap.Field    { return zap.Int("httpStatus", code) }
func DurationMs(ms int64) zap.Field    { return zap.Int64("durationMs", ms) }
func Status(s string) zap.Field        { return zap.String("status", s) }
func Provider(name string) zap.Field   { return zap.String("provider", name) }
func Endpoint(ep string) zap.Field     { return zap.String("endpoint", ep) }
