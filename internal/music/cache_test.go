package music

import (
	"context"
	"strings"
	"testing"

	"github.com/taufiq30s/chisa/internal/model"
)

// A zero-value cache must not panic or block: every caller relies on the
// fail-open contract when Redis is not configured.
func TestSearchCacheNilRedis(t *testing.T) {
	var nilCache *searchCache
	if _, ok := nilCache.get(context.Background(), "daft punk"); ok {
		t.Fatal("nil cache must report a miss")
	}
	nilCache.set(context.Background(), "daft punk", []model.Track{{ID: "x"}})

	empty := newSearchCache(nil)
	if _, ok := empty.get(context.Background(), "daft punk"); ok {
		t.Fatal("cache without a client must report a miss")
	}
	empty.set(context.Background(), "daft punk", nil) // must not panic
}

// "Daft Punk", " daft punk " and "DAFT PUNK" are one cache entry, otherwise the
// same search typed differently misses and burns another yt-dlp run.
func TestSearchCacheKeyNormalization(t *testing.T) {
	c := newSearchCache(nil)

	base := c.key("daft punk")
	for _, variant := range []string{"Daft Punk", "  daft punk  ", "DAFT PUNK", "\tdaft punk\n"} {
		if got := c.key(variant); got != base {
			t.Errorf("key(%q) = %q, want %q", variant, got, base)
		}
	}
	if c.key("daft punk live") == base {
		t.Error("distinct queries must not share a key")
	}
}

// The key must be a fixed-size hash with a scannable prefix, never the raw query:
// user input is arbitrary length and would otherwise become a huge Redis key.
func TestSearchCacheKeyShape(t *testing.T) {
	c := newSearchCache(nil)

	query := strings.Repeat("a", 4000)
	key := c.key(query)

	if !strings.HasPrefix(key, "music:search:") {
		t.Fatalf("key %q lacks the debug prefix", key)
	}
	// 64 hex chars of sha256.
	if got := len(strings.TrimPrefix(key, "music:search:")); got != 64 {
		t.Errorf("hash length = %d, want 64", got)
	}
	if strings.Contains(key, "aaaa") {
		t.Error("key must not contain the raw query")
	}
}
