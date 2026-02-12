package client

import (
	"lending-hub-service/internal/adapter/lazypay/config"
	"lending-hub-service/internal/adapter/lazypay/signature"
	"lending-hub-service/internal/infrastructure/http/executor"
	onbPort "lending-hub-service/internal/domain/onboarding/port"
	orderPort "lending-hub-service/internal/domain/order/port"
	profilePort "lending-hub-service/internal/domain/profile/port"
	refundPort "lending-hub-service/internal/domain/refund/port"
)

// LazypayClient aggregates all sub-clients and returns gateway implementations
type LazypayClient struct {
	config      *config.LazypayConfig
	signer      *signature.SignatureService
	profileExec executor.HttpExecutor
	paymentExec executor.HttpExecutor
	profile     *ProfileClient
	onboarding  *OnboardingClient
	payment     *PaymentClient
	refund      *RefundClient
}

// NewLazypayClient creates a new Lazypay client with all sub-clients
func NewLazypayClient(
	cfg *config.LazypayConfig,
	profileExec executor.HttpExecutor,
	paymentExec executor.HttpExecutor,
) *LazypayClient {
	signer := signature.NewSignatureService(cfg.AccessKey, cfg.SecretKey)
	c := &LazypayClient{
		config:      cfg,
		signer:      signer,
		profileExec: profileExec,
		paymentExec: paymentExec,
	}
	c.profile = NewProfileClient(cfg, signer, profileExec)
	c.onboarding = NewOnboardingClient(cfg, signer, profileExec)
	c.payment = NewPaymentClient(cfg, signer, paymentExec)
	c.refund = NewRefundClient(cfg, signer, paymentExec)
	return c
}

// Gateway accessors — return interface implementations
func (c *LazypayClient) ProfileGateway() profilePort.ProfileGateway {
	return c.profile
}

func (c *LazypayClient) OnboardingGateway() onbPort.OnboardingGateway {
	return c.onboarding
}

func (c *LazypayClient) OrderGateway() orderPort.OrderGateway {
	return c.payment
}

func (c *LazypayClient) RefundGateway() refundPort.RefundGateway {
	return c.refund
}
