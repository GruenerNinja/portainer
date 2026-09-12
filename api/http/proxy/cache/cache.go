package cache

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/portainer/portainer/pkg/schedule"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

const (
	DefaultRefreshInterval = time.Hour
	defaultIdleTTL         = 2 * time.Hour
	maxConcurrentRefreshes = 4
	maxEntries             = 512
	maxCacheableBodySize   = 128 << 20
)

type Fetcher func(context.Context) (*http.Response, error)

type responseSnapshot struct {
	status        string
	statusCode    int
	proto         string
	protoMajor    int
	protoMinor    int
	header        http.Header
	body          []byte
	contentLength int64
	trailer       http.Header
	uncompressed  bool
}

func snapshotResponse(response *http.Response) (*responseSnapshot, error) {
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil, err
	}

	header := response.Header.Clone()
	if header == nil {
		header = http.Header{}
	}
	header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	header.Del("Transfer-Encoding")

	return &responseSnapshot{
		status:        response.Status,
		statusCode:    response.StatusCode,
		proto:         response.Proto,
		protoMajor:    response.ProtoMajor,
		protoMinor:    response.ProtoMinor,
		header:        header,
		body:          body,
		contentLength: int64(len(body)),
		trailer:       response.Trailer.Clone(),
		uncompressed:  response.Uncompressed,
	}, nil
}

func (snapshot *responseSnapshot) response(request *http.Request) *http.Response {
	return &http.Response{
		Status:        snapshot.status,
		StatusCode:    snapshot.statusCode,
		Proto:         snapshot.proto,
		ProtoMajor:    snapshot.protoMajor,
		ProtoMinor:    snapshot.protoMinor,
		Header:        snapshot.header.Clone(),
		Body:          io.NopCloser(bytes.NewReader(snapshot.body)),
		ContentLength: snapshot.contentLength,
		Trailer:       snapshot.trailer.Clone(),
		Request:       request,
		Uncompressed:  snapshot.uncompressed,
	}
}

type entry struct {
	response   *responseSnapshot
	fetch      Fetcher
	lastAccess time.Time
}

// Cache stores unfiltered overview responses. Callers remain responsible for
// applying authorization and response filtering after Get returns.
type Cache struct {
	mu              sync.RWMutex
	entries         map[string]*entry
	backgroundCtx   context.Context
	refreshInterval time.Duration
	idleTTL         time.Duration
	refreshGroup    singleflight.Group
	startOnce       sync.Once
	refreshSlots    chan struct{}
}

func New() *Cache {
	return newCache(DefaultRefreshInterval, defaultIdleTTL)
}

func newCache(refreshInterval, idleTTL time.Duration) *Cache {
	return &Cache{
		entries:         map[string]*entry{},
		backgroundCtx:   context.Background(),
		refreshInterval: refreshInterval,
		idleTTL:         idleTTL,
		refreshSlots:    make(chan struct{}, maxConcurrentRefreshes),
	}
}

func (cache *Cache) Start(ctx context.Context) {
	cache.startOnce.Do(func() {
		cache.mu.Lock()
		cache.backgroundCtx = ctx
		cache.mu.Unlock()

		go schedule.RunOnInterval(ctx, cache.refreshInterval, cache.refreshAll, nil)
	})
}

// Get serves a warm response immediately and refreshes it asynchronously. A
// cold request is fetched once synchronously so callers never receive a made-up
// empty response.
func (cache *Cache) Get(ctx context.Context, key string, request *http.Request, fetch Fetcher) (*http.Response, error) {
	now := time.Now()
	cache.mu.Lock()
	cachedEntry := cache.entries[key]
	if cachedEntry == nil {
		cache.evictOldestEntryIfFull()
		cachedEntry = &entry{}
		cache.entries[key] = cachedEntry
	}
	cachedEntry.fetch = fetch
	cachedEntry.lastAccess = now
	cachedResponse := cachedEntry.response
	backgroundCtx := cache.backgroundCtx
	cache.mu.Unlock()

	if cachedResponse != nil {
		cache.refreshAsync(backgroundCtx, key, fetch)
		return cachedResponse.response(request), nil
	}

	freshResponse, err := cache.refresh(ctx, key, fetch)
	if err != nil {
		return nil, err
	}

	return freshResponse.response(request), nil
}

func (cache *Cache) refresh(ctx context.Context, key string, fetch Fetcher) (*responseSnapshot, error) {
	value, err, _ := cache.refreshGroup.Do(key, func() (any, error) {
		return cache.fetchAndStore(ctx, key, fetch)
	})
	if err != nil {
		return nil, err
	}

	return value.(*responseSnapshot), nil
}

func (cache *Cache) refreshAsync(ctx context.Context, key string, fetch Fetcher) {
	_ = cache.refreshGroup.DoChan(key, func() (any, error) {
		select {
		case cache.refreshSlots <- struct{}{}:
			defer func() { <-cache.refreshSlots }()
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		response, err := cache.fetchAndStore(ctx, key, fetch)
		if err != nil && ctx.Err() == nil {
			log.Warn().Err(err).Msg("unable to refresh environment overview cache")
		}

		return response, err
	})
}

func (cache *Cache) fetchAndStore(ctx context.Context, key string, fetch Fetcher) (*responseSnapshot, error) {
	response, err := fetch(ctx)
	if err != nil {
		return nil, err
	}

	snapshot, err := snapshotResponse(response)
	if err != nil {
		return nil, err
	}

	// Preserve the last successful list when the environment temporarily
	// returns an error. Cold callers still receive the original error response.
	if snapshot.statusCode >= http.StatusOK && snapshot.statusCode < http.StatusMultipleChoices && len(snapshot.body) <= maxCacheableBodySize {
		cache.mu.Lock()
		cachedEntry := cache.entries[key]
		if cachedEntry != nil {
			cachedEntry.response = snapshot
		}
		cache.mu.Unlock()
	}

	return snapshot, nil
}

// evictOldestEntryIfFull must be called with cache.mu held.
func (cache *Cache) evictOldestEntryIfFull() {
	if len(cache.entries) < maxEntries {
		return
	}

	var oldestKey string
	var oldestAccess time.Time
	for key, cachedEntry := range cache.entries {
		if oldestKey == "" || cachedEntry.lastAccess.Before(oldestAccess) {
			oldestKey = key
			oldestAccess = cachedEntry.lastAccess
		}
	}

	delete(cache.entries, oldestKey)
}

func (cache *Cache) refreshAll() {
	now := time.Now()

	type refresh struct {
		key   string
		fetch Fetcher
	}
	refreshes := []refresh{}

	cache.mu.Lock()
	ctx := cache.backgroundCtx
	for key, cachedEntry := range cache.entries {
		if now.Sub(cachedEntry.lastAccess) >= cache.idleTTL {
			delete(cache.entries, key)
			continue
		}
		if cachedEntry.response != nil && cachedEntry.fetch != nil {
			refreshes = append(refreshes, refresh{key: key, fetch: cachedEntry.fetch})
		}
	}
	cache.mu.Unlock()

	for _, refresh := range refreshes {
		cache.refreshAsync(ctx, refresh.key, refresh.fetch)
	}
}

// DeletePrefix removes cached responses and refresh metadata for an environment
// whose proxy is being deleted or replaced.
func (cache *Cache) DeletePrefix(prefix string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for key := range cache.entries {
		if strings.HasPrefix(key, prefix) {
			delete(cache.entries, key)
		}
	}
}

// Clear invalidates every cached overview response. It is used after API
// mutations because those mutations can replace runtime resources and make a
// cached collection point at identifiers that no longer exist.
func (cache *Cache) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.entries = map[string]*entry{}
}

// NewRequestFetcher snapshots a GET request so it can safely be replayed by a
// background refresh after the original request has completed.
func NewRequestFetcher(request *http.Request, roundTrip func(*http.Request) (*http.Response, error)) Fetcher {
	template := request.Clone(context.Background())
	template.Body = nil
	template.GetBody = nil

	return func(ctx context.Context) (*http.Response, error) {
		return roundTrip(template.Clone(ctx))
	}
}

// Key creates an unambiguous cache key without exposing delimiters contained
// in a URL, query, target node, or authorization scope.
func Key(parts ...string) string {
	var builder strings.Builder
	for _, part := range parts {
		fmt.Fprintf(&builder, "%d:%s", len(part), part)
	}

	return builder.String()
}
