package useractivity

import (
	"net/http/httptest"
	"testing"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
	"github.com/stretchr/testify/require"
)

func TestFilteredAppliesSearchSortAndPagination(t *testing.T) {
	t.Parallel()
	_, store := datastore.MustNewTestStore(t, false, true)
	base := time.Now().Unix()
	entries := []portainer.ActivityLog{
		{Timestamp: base, Username: "alice", Action: "POST", Context: "/stacks"},
		{Timestamp: base + 10, Username: "bob", Action: "DELETE", Context: "/registries/1"},
		{Timestamp: base + 20, Username: "alice", Action: "PUT", Context: "/stacks/2"},
	}
	for i := range entries {
		require.NoError(t, store.ActivityLog().Create(&entries[i]))
	}

	handler := &Handler{DataStore: store}
	request := httptest.NewRequest("GET", "/useractivity/logs?keyword=alice&sortBy=Timestamp&sortDesc=true&offset=1&limit=1", nil)
	logs, total, err := handler.filtered(request)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, logs, 1)
	require.Equal(t, base, logs[0].Timestamp)
}
