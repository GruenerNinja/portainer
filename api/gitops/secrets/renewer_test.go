package secrets

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	portainer "github.com/portainer/portainer/api"
	sourceDS "github.com/portainer/portainer/api/dataservices/source"
	dataStorePkg "github.com/portainer/portainer/api/datastore"

	"github.com/stretchr/testify/require"
)

func TestVaultTokenRenewerRunOnceRenewsStoredPeriodicTokens(t *testing.T) {
	t.Parallel()

	var renewals atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/token/lookup-self":
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"ttl":       60,
					"period":    120,
					"renewable": true,
				},
			}))
		case "/v1/auth/token/renew-self":
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

	_, store := dataStorePkg.MustNewTestStore(t, false, true)
	require.NoError(t, store.Source().Create(sourceDS.InsecureNewAdminContext(), &portainer.Source{
		Type: portainer.SourceTypeVault,
		Vault: &portainer.VaultConfig{
			Address: server.URL,
			Authentication: portainer.VaultAuthentication{
				Method: "token",
				Token:  "token-value",
			},
		},
	}))

	renewer := NewVaultTokenRenewer(store)
	renewer.runOnce(t.Context())

	require.EqualValues(t, 1, renewals.Load())
}
