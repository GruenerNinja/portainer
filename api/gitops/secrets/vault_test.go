package secrets

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/require"
)

func TestResolveVaultSecret_KV2(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/secret/data/app", r.URL.Path)
		require.Equal(t, "token-value", r.Header.Get("X-Vault-Token"))
		require.Equal(t, "team-a", r.Header.Get("X-Vault-Namespace"))

		err := json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"data": map[string]any{
					"password": "p@ss",
				},
			},
		})
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	value, err := ResolveVaultSecret(t.Context(), &portainer.VaultConfig{
		Address:   server.URL,
		Namespace: "team-a",
		KVVersion: 2,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	}, "secret/app", "password")

	require.NoError(t, err)
	require.Equal(t, "p@ss", value)
}

func TestResolveVaultSecret_KV1(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/secret/app", r.URL.Path)

		err := json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"password": "p@ss",
			},
		})
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	value, err := ResolveVaultSecret(t.Context(), &portainer.VaultConfig{
		Address:   server.URL,
		KVVersion: 1,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	}, "secret/app", "password")

	require.NoError(t, err)
	require.Equal(t, "p@ss", value)
}
