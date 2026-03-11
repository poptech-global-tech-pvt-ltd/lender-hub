package idgen

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type Generator struct {
	entropy io.Reader
	mu      sync.Mutex
}

func New() *Generator {
	return &Generator{entropy: rand.Reader}
}

func (g *Generator) generate(prefix string, t time.Time) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := ulid.MustNew(ulid.Timestamp(t), g.entropy)
	return fmt.Sprintf("%s_%s", prefix, id.String())
}

func (g *Generator) Generate(prefix string) string {
	return g.generate(prefix, time.Now())
}

// Convenience methods
func (g *Generator) PaymentID() string      { return g.Generate(PrefixPayment) }
func (g *Generator) RefundID() string       { return g.Generate(PrefixRefund) }
func (g *Generator) OnboardingID() string   { return g.Generate(PrefixOnboarding) }
func (g *Generator) RequestID() string      { return g.Generate(PrefixRequest) }
func (g *Generator) WebhookID() string      { return g.Generate(PrefixWebhook) }
func (g *Generator) UserProfileID() string  { return g.Generate(PrefixUserProfile) }
func (g *Generator) IdempotencyKey() string { return g.Generate(PrefixIdempotency) }
func (g *Generator) LockingID() string      { return g.Generate(PrefixLocking) }
func (g *Generator) MerchantTxnID() string  { return g.Generate(PrefixTransaction) } // merchantTxnId for Lazypay