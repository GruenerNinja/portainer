package deployments

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackWithResolvedSecrets_ExpandsVaultPathMapping(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/secret/data/app", r.URL.Path)

		err := json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"data": map[string]any{
					"PASSWORD": "p@ss",
					"TOKEN":    "new-token",
				},
			},
		})
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	_, store := datastore.MustNewTestStore(t, false, true)

	src := &portainer.Source{
		Type: portainer.SourceTypeVault,
		Vault: &portainer.VaultConfig{
			Address:   server.URL,
			KVVersion: 2,
			Authentication: portainer.VaultAuthentication{
				Method: "token",
				Token:  "token-value",
			},
		},
	}
	require.NoError(t, store.Source().Create(adminUserContext, src))

	stack := &portainer.Stack{
		Env: []portainer.Pair{
			{Name: "TOKEN", Value: "old-token"},
		},
		SecretMappings: []portainer.StackSecretMapping{
			{
				SourceID: src.ID,
				Path:     "secret/app",
			},
		},
	}

	resolved, err := stackWithResolvedSecrets(t.Context(), store, stack)
	require.NoError(t, err)

	assert.ElementsMatch(t, []portainer.Pair{
		{Name: "PASSWORD", Value: "p@ss"},
		{Name: "TOKEN", Value: "new-token"},
	}, resolved.Env)
}
