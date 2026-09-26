package simplecache

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestShouldVisit_FirstTimeTrue(t *testing.T) {
	c := New()
	if !c.ShouldVisit("https://example.com") {
		t.Fatal("expected true on first visit of a URL")
	}
}

func TestShouldVisit_SecondTimeFalse(t *testing.T) {
	c := New()
	c.ShouldVisit("https://example.com")

	if c.ShouldVisit("https://example.com") {
		t.Fatal("expected false when the same URL is visited again")
	}
}

func TestShouldVisit_DifferentURLsAreIndependent(t *testing.T) {
	c := New()
	if !c.ShouldVisit("https://a.com") {
		t.Fatal("expected true for https://a.com")
	}
	if !c.ShouldVisit("https://b.com") {
		t.Fatal("expected true for https://b.com (different URL)")
	}
}

func TestShouldVisit_ConcurrentSafe(t *testing.T) {
	c := New()

	const n = 200
	var wg sync.WaitGroup
	trueCount := atomic.Int32{}

	for range n {
		wg.Go(func() {
			if c.ShouldVisit("https://same-url.com") {
				trueCount.Add(1)
			}
		})
	}
	wg.Wait()

	if trueCount.Load() != 1 {
		t.Fatalf("expected exactly 1 goroutine to see true, got %d", trueCount)
	}
}
