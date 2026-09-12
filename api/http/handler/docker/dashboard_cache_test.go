package docker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardDataCacheReturnsStaleDataWhileRefreshing(t *testing.T) {
	t.Parallel()

	refreshStarted := make(chan struct{})
	finishRefresh := make(chan struct{})
	var fetchCount atomic.Int32

	cache := newDashboardDataCache(func(ctx context.Context, _ dashboardCacheKey) (*dashboardData, error) {
		call := fetchCount.Add(1)
		if call == 2 {
			close(refreshStarted)
			select {
			case <-finishRefresh:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		return &dashboardData{images: imagesCounters{Total: int(call)}}, nil
	}, time.Hour)

	key := dashboardCacheKey{endpointID: 1}
	initialData, err := cache.get(t.Context(), key)
	require.NoError(t, err)
	require.Equal(t, 1, initialData.images.Total)

	cachedData, err := cache.get(t.Context(), key)
	require.NoError(t, err)
	assert.Equal(t, 1, cachedData.images.Total)

	select {
	case <-refreshStarted:
	case <-time.After(time.Second):
		t.Fatal("background refresh did not start")
	}

	// A request made while the refresh is in flight must return the stale value
	// without starting another Docker fetch.
	cachedData, err = cache.get(t.Context(), key)
	require.NoError(t, err)
	assert.Equal(t, 1, cachedData.images.Total)
	assert.Equal(t, int32(2), fetchCount.Load())

	close(finishRefresh)
	require.Eventually(t, func() bool {
		cache.mu.RLock()
		defer cache.mu.RUnlock()
		return cache.data[key] != nil && cache.data[key].images.Total == 2
	}, time.Second, 10*time.Millisecond)
}

func TestDashboardDataCacheRefreshAllUpdatesIdleEntries(t *testing.T) {
	t.Parallel()

	var fetchCount atomic.Int32
	cache := newDashboardDataCache(func(context.Context, dashboardCacheKey) (*dashboardData, error) {
		call := fetchCount.Add(1)
		return &dashboardData{images: imagesCounters{Total: int(call)}}, nil
	}, time.Hour)

	key := dashboardCacheKey{endpointID: 1, agentTarget: "node-1"}
	_, err := cache.get(t.Context(), key)
	require.NoError(t, err)

	cache.refreshAll()
	require.Eventually(t, func() bool {
		cache.mu.RLock()
		defer cache.mu.RUnlock()
		return cache.data[key].images.Total == 2
	}, time.Second, 10*time.Millisecond)
}

func TestDashboardDataCacheKeepsLastGoodDataWhenRefreshFails(t *testing.T) {
	t.Parallel()

	refreshAttempted := make(chan struct{})
	var fetchCount atomic.Int32
	cache := newDashboardDataCache(func(context.Context, dashboardCacheKey) (*dashboardData, error) {
		if fetchCount.Add(1) == 2 {
			close(refreshAttempted)
			return nil, errors.New("Docker is unavailable")
		}

		return &dashboardData{images: imagesCounters{Total: 42}}, nil
	}, time.Hour)

	key := dashboardCacheKey{endpointID: 1}
	_, err := cache.get(t.Context(), key)
	require.NoError(t, err)

	staleData, err := cache.get(t.Context(), key)
	require.NoError(t, err)
	assert.Equal(t, 42, staleData.images.Total)

	select {
	case <-refreshAttempted:
	case <-time.After(time.Second):
		t.Fatal("background refresh did not run")
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	assert.Equal(t, 42, cache.data[key].images.Total)
}
