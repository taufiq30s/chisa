package music

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/model"
	"github.com/taufiq30s/chisa/utils"
)

const (
	// SearchCacheExpiryDuration is how long a search result stays valid. 24h
	// because a video's ID, title, and duration are effectively immutable - a
	// stale entry only goes wrong when the video is deleted or region-blocked,
	// and the player's existing yt-dlp error path already reports that.
	SearchCacheExpiryDuration = 24 * time.Hour

	// cacheTimeout bounds a single Redis round trip so a hung connection cannot
	// stall a slash command.
	cacheTimeout = 2 * time.Second
)

// searchCache stores resolved searches in Redis so a repeated query costs no
// subprocess and no API quota.
//
// Fail-open everywhere: a nil client (Redis not configured), a connection error,
// or malformed JSON degrades to "cache miss". Music must keep working when Redis
// is down - the same rule checkRateLimit in internal/handlers follows.
//
// A zero searchCache is usable: methods no-op when rdb is nil, so callers never
// need to nil-check.
type searchCache struct {
	rdb *redis.Client
}

func newSearchCache(rdb *redis.Client) *searchCache {
	return &searchCache{rdb: rdb}
}

// key hashes the normalized query. Hashing keeps an arbitrarily long user query
// from becoming a huge Redis key; the readable prefix keeps it findable with
// `SCAN music:search:*` while debugging.
func (c *searchCache) key(query string) string {
	sum := sha256.Sum256([]byte(normalizeQuery(query)))
	return "music:search:" + hex.EncodeToString(sum[:])
}

// normalizeQuery makes " Daft Punk " and "daft punk" the same search.
func normalizeQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

// get returns the cached tracks. The bool reports a hit, so an empty-but-cached
// result ("this query has no videos") is distinguishable from a miss.
func (c *searchCache) get(ctx context.Context, query string) ([]model.Track, bool) {
	if c == nil || c.rdb == nil {
		return nil, false
	}

	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	raw, err := c.rdb.Get(ctx, c.key(query)).Bytes()
	if err != nil {
		if err != redis.Nil {
			utils.ErrorLog.Println("Search cache read failed:", err)
		}
		return nil, false
	}

	var tracks []model.Track
	if err := json.Unmarshal(raw, &tracks); err != nil {
		utils.ErrorLog.Println("Search cache decode failed:", err)
		return nil, false
	}
	return tracks, true
}

// set stores tracks under the query. Write errors are logged, never returned:
// a cache that cannot write is slower, not broken.
func (c *searchCache) set(ctx context.Context, query string, tracks []model.Track) {
	if c == nil || c.rdb == nil {
		return
	}

	payload, err := json.Marshal(tracks)
	if err != nil {
		utils.ErrorLog.Println("Search cache encode failed:", err)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	if err := c.rdb.Set(ctx, c.key(query), payload, SearchCacheExpiryDuration).Err(); err != nil {
		utils.ErrorLog.Println("Search cache write failed:", err)
	}
}
