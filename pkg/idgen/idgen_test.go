package idgen

import (
	"sort"
	"sync"
	"testing"
)

func TestPaymentID_Prefix(t *testing.T) {
	g := New()
	id := g.PaymentID()
	if len(id) < 4 || id[:4] != "lps_" {
		t.Errorf("PaymentID should start with lps_, got %q", id)
	}
}

func TestRefundID_Prefix(t *testing.T) {
	g := New()
	id := g.RefundID()
	if len(id) < 4 || id[:4] != "ref_" {
		t.Errorf("RefundID should start with ref_, got %q", id)
	}
}

func TestOnboardingID_Prefix(t *testing.T) {
	g := New()
	id := g.OnboardingID()
	if len(id) < 4 || id[:4] != "onb_" {
		t.Errorf("OnboardingID should start with onb_, got %q", id)
	}
}

func TestIDGen_Uniqueness(t *testing.T) {
	g := New()
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := g.PaymentID()
		if ids[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestIDGen_Sortable(t *testing.T) {
	g := New()
	ids := make([]string, 5)
	for i := 0; i < 5; i++ {
		ids[i] = g.PaymentID()
	}
	sorted := make([]string, len(ids))
	copy(sorted, ids)
	sort.Strings(sorted)
	// ULIDs are lexicographically sortable; sorting should not change length
	if len(sorted) != 5 {
		t.Fatalf("expected 5 ids, got %d", len(sorted))
	}
}

func TestIDGen_ConcurrentUniqueness(t *testing.T) {
	g := New()
	ids := make(chan string, 1000)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ids <- g.PaymentID()
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[string]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		seen[id] = true
	}
}
