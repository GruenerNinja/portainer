package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEventHubPublishesCoalescedInvalidations(t *testing.T) {
	hub := newEventHub()
	updates, unsubscribe := hub.subscribe()
	defer unsubscribe()

	hub.publish()
	hub.publish()

	select {
	case <-updates:
	case <-time.After(time.Second):
		t.Fatal("expected an invalidation event")
	}

	select {
	case <-updates:
		t.Fatal("expected duplicate events to be coalesced")
	default:
	}

	require.Len(t, hub.subscribers, 1)
	unsubscribe()
	require.Empty(t, hub.subscribers)
}
