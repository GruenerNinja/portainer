package cache

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func response(body string) *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{"X-Upstream": []string{"original"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func responseBody(t *testing.T, response *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	return string(body)
}

func TestCacheReturnsStaleResponseWhileRefreshing(t *testing.T) {
	t.Parallel()

	refreshStarted := make(chan struct{})
	finishRefresh := make(chan struct{})
	var fetchCount atomic.Int32
	cache := newCache(time.Hour, 2*time.Hour)
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://docker/containers/json", nil)
	require.NoError(t, err)

	fetch := func(ctx context.Context) (*http.Response, error) {
		call := fetchCount.Add(1)
		if call == 2 {
			close(refreshStarted)
			select {
			case <-finishRefresh:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return response(string(rune('0' + call))), nil
	}

	initial, err := cache.Get(t.Context(), "containers", request, fetch)
	require.NoError(t, err)
	assert.Equal(t, "1", responseBody(t, initial))

	stale, err := cache.Get(t.Context(), "containers", request, fetch)
	require.NoError(t, err)
	assert.Equal(t, "1", responseBody(t, stale))

	select {
	case <-refreshStarted:
	case <-time.After(time.Second):
		t.Fatal("background refresh did not start")
	}

	stale, err = cache.Get(t.Context(), "containers", request, fetch)
	require.NoError(t, err)
	assert.Equal(t, "1", responseBody(t, stale))
	assert.Equal(t, int32(2), fetchCount.Load())

	close(finishRefresh)
	require.Eventually(t, func() bool {
		cache.mu.RLock()
		defer cache.mu.RUnlock()
		return cache.entries["containers"].response != nil && string(cache.entries["containers"].response.body) == "2"
	}, time.Second, 10*time.Millisecond)
}

func TestCacheReplaysIndependentResponseCopies(t *testing.T) {
	t.Parallel()

	refreshStarted := make(chan struct{})
	finishRefresh := make(chan struct{})
	var fetchCount atomic.Int32
	cache := newCache(time.Hour, 2*time.Hour)
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://docker/networks", nil)
	require.NoError(t, err)

	fetch := func(ctx context.Context) (*http.Response, error) {
		if fetchCount.Add(1) > 1 {
			close(refreshStarted)
			select {
			case <-finishRefresh:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return response("raw"), nil
	}

	first, err := cache.Get(t.Context(), "networks", request, fetch)
	require.NoError(t, err)
	first.Header.Set("X-Upstream", "filtered")
	assert.Equal(t, "raw", responseBody(t, first))

	second, err := cache.Get(t.Context(), "networks", request, fetch)
	require.NoError(t, err)
	assert.Equal(t, "original", second.Header.Get("X-Upstream"))
	assert.Equal(t, "raw", responseBody(t, second))

	select {
	case <-refreshStarted:
	case <-time.After(time.Second):
		t.Fatal("background refresh did not start")
	}
	close(finishRefresh)
}

func TestCacheKeepsSuccessfulResponseAfterRefreshError(t *testing.T) {
	t.Parallel()

	refreshAttempted := make(chan struct{})
	var fetchCount atomic.Int32
	cache := newCache(time.Hour, 2*time.Hour)
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://docker/nodes", nil)
	require.NoError(t, err)

	fetch := func(context.Context) (*http.Response, error) {
		if fetchCount.Add(1) == 2 {
			close(refreshAttempted)
			return nil, errors.New("runner unavailable")
		}
		return response("last-good"), nil
	}

	_, err = cache.Get(t.Context(), "nodes", request, fetch)
	require.NoError(t, err)

	stale, err := cache.Get(t.Context(), "nodes", request, fetch)
	require.NoError(t, err)
	assert.Equal(t, "last-good", responseBody(t, stale))

	select {
	case <-refreshAttempted:
	case <-time.After(time.Second):
		t.Fatal("background refresh did not run")
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	assert.Equal(t, "last-good", string(cache.entries["nodes"].response.body))
}

func TestCachePeriodicallyRefreshesIdleOverview(t *testing.T) {
	t.Parallel()

	var fetchCount atomic.Int32
	cache := newCache(10*time.Millisecond, time.Hour)
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://docker/services", nil)
	require.NoError(t, err)
	fetch := func(context.Context) (*http.Response, error) {
		fetchCount.Add(1)
		return response("services"), nil
	}

	_, err = cache.Get(t.Context(), "services", request, fetch)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	cache.Start(ctx)

	require.Eventually(t, func() bool {
		return fetchCount.Load() >= 2
	}, time.Second, 10*time.Millisecond)
}

func TestNewRequestFetcherSnapshotsRequest(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest(http.MethodGet, "http://docker/containers/json?all=1", nil)
	require.NoError(t, err)
	request.Header.Set("X-Test", "original")

	fetch := NewRequestFetcher(request, func(clone *http.Request) (*http.Response, error) {
		assert.Equal(t, "original", clone.Header.Get("X-Test"))
		assert.Equal(t, "all=1", clone.URL.RawQuery)
		return response("ok"), nil
	})
	request.Header.Set("X-Test", "mutated")

	result, err := fetch(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "ok", responseBody(t, result))
}

func TestKeyDoesNotCollideWhenPartsContainDelimiters(t *testing.T) {
	t.Parallel()
	assert.NotEqual(t, Key("a", "bc"), Key("ab", "c"))
}

func TestDeletePrefixInvalidatesOnlyMatchingEnvironment(t *testing.T) {
	t.Parallel()

	cache := newCache(time.Hour, 2*time.Hour)
	cache.entries[Key("docker", "1", "/containers/json")] = &entry{}
	cache.entries[Key("docker", "2", "/containers/json")] = &entry{}
	cache.entries[Key("kubernetes", "1", "/api/v1/pods")] = &entry{}

	cache.DeletePrefix(Key("docker", "1"))

	assert.NotContains(t, cache.entries, Key("docker", "1", "/containers/json"))
	assert.Contains(t, cache.entries, Key("docker", "2", "/containers/json"))
	assert.Contains(t, cache.entries, Key("kubernetes", "1", "/api/v1/pods"))
}

func TestClearInvalidatesAllEntries(t *testing.T) {
	t.Parallel()

	cache := newCache(time.Hour, 2*time.Hour)
	cache.entries[Key("docker", "1", "/containers/json")] = &entry{}
	cache.entries[Key("kubernetes", "2", "/api/v1/pods")] = &entry{}

	cache.Clear()

	assert.Empty(t, cache.entries)
}
