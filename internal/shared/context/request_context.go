package context

import (
	"context"
)

// RequestContext carries per-request metadata end-to-end
type RequestContext struct {
	RequestID string // UUID from X-Request-ID header or generated
	UserID    string // set by handler from request body (not middleware)
	Platform  string // from x-platform header: WEB, ANDROID, IOS
	DeviceID  string // from x-device-id header
	UserIP    string // from x-user-ip header
	Source    string // set by handler from request body: PDP, CART, CHECKOUT, CX
}

type ctxKey struct{}

// WithRequestContext stores RequestContext in context.Context
func WithRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	return context.WithValue(ctx, ctxKey{}, rc)
}

// FromContext retrieves RequestContext; returns empty struct if not found
func FromContext(ctx context.Context) *RequestContext {
	if rc, ok := ctx.Value(ctxKey{}).(*RequestContext); ok {
		return rc
	}
	return &RequestContext{}
}
