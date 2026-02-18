package client

import (
	"context"
	"fmt"

	lpConstants "lending-hub-service/internal/adapter/lazypay/constants"
	sharedContext "lending-hub-service/internal/shared/context"
	baseLogger "lending-hub-service/pkg/logger"
	"go.uber.org/zap"
)

// logRequest logs the outgoing Lazypay request
func logLazypayRequest(logger *baseLogger.Logger, ctx context.Context, method, url string, headers map[string]string, body []byte) {
	rc := sharedContext.FromContext(ctx)

	// Mask sensitive headers
	maskedHeaders := make(map[string]string)
	for k, v := range headers {
		if k == lpConstants.HeaderSignature || k == lpConstants.HeaderAccessKey {
			maskedHeaders[k] = maskString(v, 4)
		} else {
			maskedHeaders[k] = v
		}
	}

	// Log request
	logger.Info("Lazypay request",
		baseLogger.Module("lazypay"),
		baseLogger.RequestID(rc.RequestID),
		baseLogger.Provider("lazypay"),
		baseLogger.HTTPStatus(0), // 0 for request
		baseLogger.Status("request"),
		baseLogger.Endpoint(fmt.Sprintf("%s %s", method, url)),
		zap.Any("headers", maskedHeaders),
		zap.String("body", string(body)),
	)
}

// logResponse logs the incoming Lazypay response
func logLazypayResponse(logger *baseLogger.Logger, ctx context.Context, url string, statusCode int, body []byte, err error) {
	rc := sharedContext.FromContext(ctx)

	if err != nil {
		logger.Error("Lazypay request failed",
			baseLogger.Module("lazypay"),
			baseLogger.RequestID(rc.RequestID),
			baseLogger.Provider("lazypay"),
			baseLogger.ErrorCode(err.Error()),
			baseLogger.Status("error"),
			baseLogger.Endpoint(url),
		)
		return
	}

	// Truncate large response bodies for logging
	bodyStr := string(body)
	if len(bodyStr) > 1000 {
		bodyStr = bodyStr[:1000] + "... (truncated)"
	}

	logger.Info("Lazypay response",
		baseLogger.Module("lazypay"),
		baseLogger.RequestID(rc.RequestID),
		baseLogger.Provider("lazypay"),
		baseLogger.HTTPStatus(statusCode),
		baseLogger.Status("response"),
		baseLogger.Endpoint(url),
		zap.String("body", bodyStr),
	)
}

// maskString masks all but the first and last N characters
func maskString(s string, visibleChars int) string {
	if len(s) <= visibleChars*2 {
		return "***"
	}
	return s[:visibleChars] + "***" + s[len(s)-visibleChars:]
}
