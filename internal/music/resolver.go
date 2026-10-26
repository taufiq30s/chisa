package music

import (
	"context"
	"fmt"

	"github.com/taufiq30s/chisa/internal/model"
	"github.com/taufiq30s/chisa/internal/music/providers"
)

type AudioProvider interface {
	Resolve(ctx context.Context, query string) ([]model.Track, bool, error)
	Accepts(query string) bool
}

type Resolver struct {
	providers []AudioProvider
	cache     *searchCache
}

// Resolve is the single place search results are cached, so both the URL and the
// text-search path share one cache without a second wrapper type.
//
// Only text searches are cached. A URL resolve is one yt-dlp run with no quota
// cost and users rarely repeat the same link; caching it would just fill Redis
// with one-entry keys.
//
// Hidden cost note: a cache hit skips the provider entirely - no yt-dlp
// subprocess, no API unit. A miss costs one `ytsearch` (~2.5s, 0 units).
func (r *Resolver) Resolve(ctx context.Context, query string) ([]model.Track, bool, error) {
	for _, provider := range r.providers {
		if !provider.Accepts(query) {
			continue
		}
		if !providers.IsURL(query) {
			if tracks, ok := r.cache.get(ctx, query); ok {
				return tracks, true, nil
			}
			tracks, isSearch, err := provider.Resolve(ctx, query)
			if err != nil {
				return nil, isSearch, err
			}
			// No negative caching: an empty result may be a transient YouTube
			// hiccup, and caching it would hide the query for a full day.
			if len(tracks) > 0 {
				r.cache.set(ctx, query, tracks)
			}
			return tracks, isSearch, nil
		}
		return provider.Resolve(ctx, query)
	}
	return nil, false, fmt.Errorf("no provider found for query: %s", query)
}
