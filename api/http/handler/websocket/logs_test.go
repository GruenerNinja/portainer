package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDockerLogsProxyRequest(t *testing.T) {
	r := httptest.NewRequest(
		http.MethodGet,
		"/websocket/logs?resource=containers&id=abc123&timestamps=1&since=42&tail=20000&nodeName=worker-1",
		nil,
	)
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Accept-Encoding", "gzip")
	r.Header.Set("Authorization", "Bearer secret")
	r.AddCookie(&http.Cookie{Name: "jwt", Value: "secret"})

	proxyRequest, err := createDockerLogsProxyRequest(r)
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, proxyRequest.Method)
	assert.Equal(t, "/containers/abc123/logs", proxyRequest.URL.Path)
	assert.Equal(t, "1", proxyRequest.URL.Query().Get("follow"))
	assert.Equal(t, "true", proxyRequest.URL.Query().Get("timestamps"))
	assert.Equal(t, "42", proxyRequest.URL.Query().Get("since"))
	assert.Equal(t, "10000", proxyRequest.URL.Query().Get("tail"))
	assert.Equal(t, "worker-1", proxyRequest.Header.Get(portainer.PortainerAgentTargetHeader))
	assert.Equal(t, "application/octet-stream", proxyRequest.Header.Get("Accept"))
	assert.Empty(t, proxyRequest.Header.Get("Connection"))
	assert.Empty(t, proxyRequest.Header.Get("Upgrade"))
	assert.Empty(t, proxyRequest.Header.Get("Accept-Encoding"))
	assert.Empty(t, proxyRequest.Header.Get("Authorization"))
	assert.Empty(t, proxyRequest.Header.Get("Cookie"))
}

func TestCreateDockerLogsProxyRequestRejectsInvalidParameters(t *testing.T) {
	tests := []string{
		"resource=images&id=abc123",
		"resource=containers&id=../secret",
		"resource=services&id=abc123&tail=-1",
		"resource=tasks&id=abc123&since=nope",
	}

	for _, query := range tests {
		r := httptest.NewRequest(http.MethodGet, "/websocket/logs?"+query, nil)
		_, err := createDockerLogsProxyRequest(r)
		assert.Error(t, err, query)
	}
}
