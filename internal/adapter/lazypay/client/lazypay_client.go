package client

import (
	"lending-hub-service/internal/adapter/lazypay/config"
	"lending-hub-service/internal/adapter/lazypay/signature"
	onbPort "lending-hub-service/internal/domain/onboarding/port"
	orderPort "lending-hub-service/internal/domain/order/port"
	profilePort "lending-hub-service/internal/domain/profile/port"
	refundPort "lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/internal/infrastructure/http/executor"
	"lending-hub-service/pkg/idgen"
	baseLogger "lending-hub-service/pkg/logger"
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
	logger *baseLogger.Logger,
	idgen *idgen.Generator,
) *LazypayClient {
	signer := signature.NewSignatureService(cfg.AccessKey, cfg.SecretKey)
	c := &LazypayClient{
		config:      cfg,
		signer:      signer,
		profileExec: profileExec,
		paymentExec: paymentExec,
	}
	c.profile = NewProfileClient(cfg, signer, profileExec, logger)
	c.onboarding = NewOnboardingClient(cfg, signer, profileExec, logger)
	c.payment = NewPaymentClient(cfg, signer, paymentExec, logger)
	c.refund = NewRefundClient(cfg, signer, paymentExec, logger, idgen)
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
