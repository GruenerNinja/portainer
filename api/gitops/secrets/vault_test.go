package secrets

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultConnectionValidatesToken(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		require.Equal(t, "token-value", r.Header.Get("X-Vault-Token"))
		require.Equal(t, "team-a", r.Header.Get("X-Vault-Namespace"))

		switch r.URL.Path {
		case "/v1/sys/health":
			w.WriteHeader(http.StatusOK)
		case "/v1/auth/token/lookup-self":
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected Vault request %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(server.Close)

	err := TestVaultConnection(t.Context(), &portainer.VaultConfig{
		Address:   server.URL,
		Namespace: "team-a",
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	})

	require.NoError(t, err)
	require.EqualValues(t, 2, requests.Load())
}

func TestVaultConnectionRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/health":
			w.WriteHeader(http.StatusOK)
		case "/v1/auth/token/lookup-self":
			w.WriteHeader(http.StatusForbidden)
		default:
			t.Fatalf("unexpected Vault request %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(server.Close)

	err := TestVaultConnection(t.Context(), &portainer.VaultConfig{
		Address: server.URL,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "expired-token",
		},
	})

	require.EqualError(t, err, "vault token validation failed with status 403")
}

func TestVaultConnectionAcceptsHealthyStandby(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/health":
			w.WriteHeader(http.StatusTooManyRequests)
		case "/v1/auth/token/lookup-self":
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected Vault request %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(server.Close)

	err := TestVaultConnection(t.Context(), &portainer.VaultConfig{
		Address: server.URL,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	})

	require.NoError(t, err)
}

func TestVaultConnectionPrefersInternalAddress(t *testing.T) {
	t.Parallel()

	publicRequests := atomic.Int32{}
	publicServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		publicRequests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(publicServer.Close)

	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/health", "/v1/auth/token/lookup-self":
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected Vault request %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(internalServer.Close)

	err := TestVaultConnection(t.Context(), &portainer.VaultConfig{
		Address:         publicServer.URL,
		InternalAddress: internalServer.URL,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	})

	require.NoError(t, err)
	require.Zero(t, publicRequests.Load())
}

func TestVaultConnectionFallsBackToPublicAddress(t *testing.T) {
	t.Parallel()

	internalRequests := atomic.Int32{}
	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		internalRequests.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(internalServer.Close)

	publicServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/health", "/v1/auth/token/lookup-self":
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected Vault request %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(publicServer.Close)

	err := TestVaultConnection(t.Context(), &portainer.VaultConfig{
		Address:         publicServer.URL,
		InternalAddress: internalServer.URL,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	})

	require.NoError(t, err)
	require.EqualValues(t, 1, internalRequests.Load())
}

func TestVaultConnectionRejectsUnhealthyServer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	err := TestVaultConnection(t.Context(), &portainer.VaultConfig{
		Address: server.URL,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	})

	require.EqualError(t, err, "vault health check failed with status 503")
}

func TestLookupVaultTokenReturnsRenewalMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/v1/auth/token/lookup-self", r.URL.Path)
		require.Equal(t, "token-value", r.Header.Get("X-Vault-Token"))
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"ttl":              1_296_000,
				"creation_ttl":     2_592_000,
				"period":           2_592_000,
				"explicit_max_ttl": 0,
				"renewable":        true,
			},
		}))
	}))
	t.Cleanup(server.Close)

	info, err := LookupVaultToken(t.Context(), &portainer.VaultConfig{
		Address: server.URL,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1_296_000), info.TTL)
	assert.Equal(t, int64(2_592_000), info.Period)
	assert.True(t, info.Renewable)
}

func TestRenewVaultTokenIfNeededRenewsPeriodicTokenAtHalfPeriod(t *testing.T) {
	t.Parallel()

	var renewals atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "token-value", r.Header.Get("X-Vault-Token"))

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/token/lookup-self":
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"ttl":       60,
					"period":    120,
					"renewable": true,
				},
			}))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/auth/token/renew-self":
			renewals.Add(1)
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"auth": map[string]any{
					"lease_duration": 120,
					"renewable":      true,
				},
			}))
		default:
			t.Fatalf("unexpected Vault request %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(server.Close)

	result, err := RenewVaultTokenIfNeeded(t.Context(), &portainer.VaultConfig{
		Address: server.URL,
		Authentication: portainer.VaultAuthentication{
			Method: "token",
			Token:  "token-value",
		},
	})

	require.NoError(t, err)
	assert.True(t, result.Renewed)
	assert.Equal(t, int64(120), result.TTL)
	assert.EqualValues(t, 1, renewals.Load())
}

func TestRenewVaultTokenIfNeededDoesNotRenewEarlyOrNonPeriodicTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		ttl    int64
		period int64
	}{
		{name: "periodic token above half period", ttl: 61, period: 120},
		{name: "renewable non-periodic token", ttl: 60, period: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v1/auth/token/lookup-self", r.URL.Path)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"ttl":       test.ttl,
						"period":    test.period,
						"renewable": true,
					},
				}))
			}))
			t.Cleanup(server.Close)

			result, err := RenewVaultTokenIfNeeded(t.Context(), &portainer.VaultConfig{
				Address: server.URL,
				Authentication: portainer.VaultAuthentication{
					Method: "token",
					Token:  "token-value",
				},
			})

			require.NoError(t, err)
			assert.False(t, result.Renewed)
		})
	}
}

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

func TestResolveVaultSecretFallsBackToPublicAddress(t *testing.T) {
	t.Parallel()

	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(internalServer.Close)

	publicServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/secret/data/app", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"data": map[string]any{"password": "p@ss"},
			},
		}))
	}))
	t.Cleanup(publicServer.Close)

	value, err := ResolveVaultSecret(t.Context(), &portainer.VaultConfig{
		Address:         publicServer.URL,
		InternalAddress: internalServer.URL,
		KVVersion:       2,
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
