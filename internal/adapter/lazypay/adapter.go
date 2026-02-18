package lazypay

import (
	"lending-hub-service/internal/adapter/lazypay/client"
	"lending-hub-service/internal/adapter/lazypay/config"
	"lending-hub-service/internal/infrastructure/http/executor"
	baseLogger "lending-hub-service/pkg/logger"
)

// NewAdapter creates the full Lazypay adapter with all gateways
func NewAdapter(cfg *config.LazypayConfig, logger *baseLogger.Logger) *client.LazypayClient {
	profileExec := executor.NewProfileExecutor()
	paymentExec := executor.NewPaymentExecutor()
	return client.NewLazypayClient(cfg, profileExec, paymentExec, logger)
}
