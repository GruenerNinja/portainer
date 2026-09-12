package registries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryRepositoriesList(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, false, true)
	registry := &portainer.Registry{ID: 6, Name: "Github", Type: portainer.GithubRegistry, URL: "ghcr.io"}
	require.NoError(t, store.Registry().Create(registry))

	handler := &Handler{
		DataStore: store,
		listRepositories: func(_ context.Context, got *portainer.Registry) ([]string, error) {
			assert.Equal(t, registry.ID, got.ID)
			return []string{"themodcrafttmc/portainer", "themodcrafttmc/wall"}, nil
		},
	}

	id := strconv.Itoa(int(registry.ID))
	r := httptest.NewRequest(http.MethodGet, "/registries/"+id+"/v2/_catalog", nil)
	r = mux.SetURLVars(r, map[string]string{"id": id})
	w := httptest.NewRecorder()

	handlerErr := handler.registryRepositoriesList(w, r)
	require.Nil(t, handlerErr)
	require.Equal(t, http.StatusOK, w.Code)

	var got registryCatalogResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Equal(t, []string{"themodcrafttmc/portainer", "themodcrafttmc/wall"}, got.Repositories)
}

func TestRegistryRepositoryTagsList(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, false, true)
	registry := &portainer.Registry{ID: 6, Name: "Github", Type: portainer.GithubRegistry, URL: "ghcr.io"}
	require.NoError(t, store.Registry().Create(registry))

	handler := &Handler{
		DataStore: store,
		listTags: func(_ context.Context, got *portainer.Registry, repository string) ([]string, error) {
			assert.Equal(t, registry.ID, got.ID)
			assert.Equal(t, "themodcrafttmc/portainer", repository)
			return []string{"latest", "2.39.3.2.20"}, nil
		},
	}

	id := strconv.Itoa(int(registry.ID))
	r := httptest.NewRequest(http.MethodGet, "/registries/"+id+"/v2/themodcrafttmc/portainer/tags/list", nil)
	r = mux.SetURLVars(r, map[string]string{"id": id, "repository": "themodcrafttmc/portainer"})
	w := httptest.NewRecorder()

	handlerErr := handler.registryRepositoryTagsList(w, r)
	require.Nil(t, handlerErr)
	require.Equal(t, http.StatusOK, w.Code)

	var got registryTagsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Equal(t, "themodcrafttmc/portainer", got.Name)
	assert.Equal(t, []string{"latest", "2.39.3.2.20"}, got.Tags)
}
