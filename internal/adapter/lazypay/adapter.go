package lazypay

import (
	"lending-hub-service/internal/adapter/lazypay/client"
	"lending-hub-service/internal/adapter/lazypay/config"
	"lending-hub-service/internal/infrastructure/http/executor"
)

// NewAdapter creates the full Lazypay adapter with all gateways
func NewAdapter(cfg *config.LazypayConfig) *client.LazypayClient {
	profileExec := executor.NewProfileExecutor()
	paymentExec := executor.NewPaymentExecutor()
	return client.NewLazypayClient(cfg, profileExec, paymentExec)
}
