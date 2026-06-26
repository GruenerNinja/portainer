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

func TestResolveVaultSecretValues_KV2(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/secret/data/app", r.URL.Path)

		err := json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"data": map[string]any{
					"password": "p@ss",
					"port":     5432,
				},
			},
		})
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	values, err := ResolveVaultSecretValues(t.Context(), &portainer.VaultConfig{
		Address:   server.URL,
		KVVersion: 2,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	}, "secret/app")

	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"password": "p@ss",
		"port":     "5432",
	}, values)
}

func TestResolveVaultSecretValues_KV2FolderFallback(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/kv/data/tmc-proxy":
			http.NotFound(w, r)
		case r.Method == "LIST" && r.URL.Path == "/v1/kv/metadata/tmc-proxy":
			err := json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"keys": []string{"DATABASE_PASSWORD", "API_TOKEN", "nested/"},
				},
			})
			require.NoError(t, err)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/kv/data/tmc-proxy/DATABASE_PASSWORD":
			err := json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"value": "p@ss",
					},
				},
			})
			require.NoError(t, err)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/kv/data/tmc-proxy/API_TOKEN":
			err := json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"token":   "abc",
						"expires": 30,
					},
				},
			})
			require.NoError(t, err)
		default:
			t.Fatalf("unexpected Vault request %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(server.Close)

	values, err := ResolveVaultSecretValues(t.Context(), &portainer.VaultConfig{
		Address:   server.URL,
		KVVersion: 2,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	}, "kv/tmc-proxy")

	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"DATABASE_PASSWORD": "p@ss",
		"API_TOKEN_expires": "30",
		"API_TOKEN_token":   "abc",
	}, values)
}
