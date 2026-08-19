package activitylog

import (
	"testing"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/database"
	"github.com/stretchr/testify/require"
)

func TestCreateRetainsRecentEntriesAndPrunesExpiredEntries(t *testing.T) {
	connection, err := database.NewDatabase("boltdb", t.TempDir(), nil, false)
	require.NoError(t, err)
	require.NoError(t, connection.Open())
	t.Cleanup(func() { require.NoError(t, connection.Close()) })

	service, err := NewService(connection)
	require.NoError(t, err)
	require.NoError(t, service.Create(&portainer.ActivityLog{Timestamp: time.Now().Add(-Retention - time.Hour).Unix(), Action: "OLD"}))
	require.NoError(t, service.Create(&portainer.ActivityLog{Timestamp: time.Now().Unix(), Action: "POST"}))

	entries, err := service.ReadAll()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "POST", entries[0].Action)
}
