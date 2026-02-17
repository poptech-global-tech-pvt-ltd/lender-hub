package idgen

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// IDGenerator generates prefixed ULID-based identifiers
type IDGenerator struct {
	entropy io.Reader
	mu      sync.Mutex
}

// NewIDGenerator creates a new ID generator with crypto/rand entropy
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		entropy: rand.Reader,
	}
}

// Generate creates a new ID with the given prefix
// Format: <PREFIX>_<ULID>
// Example: PAY_01ARZ3NDEKTSV4RRFFQ69G5FAV
func (g *IDGenerator) Generate(prefix string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := ulid.MustNew(ulid.Timestamp(time.Now()), g.entropy)
	return fmt.Sprintf("%s_%s", prefix, id.String())
}

// GenerateWithTime creates an ID with a specific timestamp (for testing)
func (g *IDGenerator) GenerateWithTime(prefix string, t time.Time) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := ulid.MustNew(ulid.Timestamp(t), g.entropy)
	return fmt.Sprintf("%s_%s", prefix, id.String())
}

// Validate checks if an ID follows the expected format
func (g *IDGenerator) Validate(id, expectedPrefix string) bool {
	if len(id) < len(expectedPrefix)+2 {
		return false
	}

	// Check prefix
	if id[:len(expectedPrefix)] != expectedPrefix {
		return false
	}

	// Check separator
	if id[len(expectedPrefix)] != '_' {
		return false
	}

	// Check ULID format (26 chars after prefix and separator)
	ulidPart := id[len(expectedPrefix)+1:]
	if len(ulidPart) != 26 {
		return false
	}

	// Try to parse ULID
	_, err := ulid.Parse(ulidPart)
	return err == nil
}

// ExtractTimestamp extracts the timestamp from a ULID-based ID
func (g *IDGenerator) ExtractTimestamp(id string) (time.Time, error) {
	// Find separator
	sepIdx := -1
	for i, c := range id {
		if c == '_' {
			sepIdx = i
			break
		}
	}

	if sepIdx == -1 {
		return time.Time{}, fmt.Errorf("invalid ID format: no separator")
	}

	ulidPart := id[sepIdx+1:]
	parsed, err := ulid.Parse(ulidPart)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid ULID: %w", err)
	}

	return ulid.Time(parsed.Time()), nil
}

// Convenience methods for common ID types

func (g *IDGenerator) PaymentID() string {
	return g.Generate(PrefixPayment)
}

func (g *IDGenerator) RefundID() string {
	return g.Generate(PrefixRefund)
}

func (g *IDGenerator) OnboardingID() string {
	return g.Generate(PrefixOnboarding)
}

func (g *IDGenerator) RequestID() string {
	return g.Generate(PrefixRequest)
}

func (g *IDGenerator) WebhookID() string {
	return g.Generate(PrefixWebhook)
}

func (g *IDGenerator) UserProfileID() string {
	return g.Generate(PrefixUserProfile)
}

func (g *IDGenerator) IdempotencyKey() string {
	return g.Generate(PrefixIdempotency)
}
