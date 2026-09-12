package websocket

import (
	"net/http"
	"sync"
	"time"

	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/websocket"
)

type eventHub struct {
	mu          sync.RWMutex
	subscribers map[chan struct{}]struct{}
}

func newEventHub() *eventHub {
	return &eventHub{subscribers: map[chan struct{}]struct{}{}}
}

func (hub *eventHub) subscribe() (<-chan struct{}, func()) {
	updates := make(chan struct{}, 1)
	hub.mu.Lock()
	hub.subscribers[updates] = struct{}{}
	hub.mu.Unlock()
	return updates, func() {
		hub.mu.Lock()
		delete(hub.subscribers, updates)
		hub.mu.Unlock()
	}
}

func (hub *eventHub) publish() {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for subscriber := range hub.subscribers {
		select {
		case subscriber <- struct{}{}:
		default:
		}
	}
}

func (h *Handler) websocketEvents(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	connection, err := h.connectionUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return httperror.InternalServerError("Unable to upgrade event stream connection", err)
	}
	defer connection.Close()

	updates, unsubscribe := h.eventHub.subscribe()
	defer unsubscribe()

	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-updates:
			if err := connection.WriteJSON(map[string]string{"type": "invalidate"}); err != nil {
				return nil
			}
		case <-keepalive.C:
			if err := connection.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return nil
			}
		case <-r.Context().Done():
			return nil
		}
	}
}
