package migrator

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/database/boltdb"
	"github.com/portainer/portainer/api/dataservices/role"
	"github.com/portainer/portainer/api/dataservices/settings"
	"github.com/portainer/portainer/api/logs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestorePredefinedRoles(t *testing.T) {
	t.Parallel()

	conn := &boltdb.DbConnection{Path: t.TempDir()}
	require.NoError(t, conn.Open())
	defer logs.CloseAndLogErr(conn)

	roleService, err := role.NewService(conn)
	require.NoError(t, err)
	settingsService, err := settings.NewService(conn)
	require.NoError(t, err)
	require.NoError(t, settingsService.UpdateSettings(&portainer.Settings{
		AllowVolumeBrowserForRegularUsers: true,
	}))

	migrator := NewMigrator(&MigratorParameters{
		RoleService:     roleService,
		SettingsService: settingsService,
	})

	require.NoError(t, migrator.restorePredefinedRoles())
	require.NoError(t, migrator.restorePredefinedRoles(), "migration must be idempotent")

	roles, err := roleService.ReadAll()
	require.NoError(t, err)
	require.Len(t, roles, 5)

	operator, err := roleService.Read(portainer.RoleIDOperator)
	require.NoError(t, err)
	assert.Equal(t, 2, operator.Priority)
	assert.True(t, operator.Authorizations[portainer.OperationDockerContainerRestart])
	assert.True(t, operator.Authorizations[portainer.OperationDockerAgentBrowseList])
	assert.False(t, operator.Authorizations[portainer.OperationDockerContainerCreate])

	helpdesk, err := roleService.Read(portainer.RoleIDHelpdesk)
	require.NoError(t, err)
	assert.Equal(t, 3, helpdesk.Priority)

	readonly, err := roleService.Read(portainer.RoleIDReadonly)
	require.NoError(t, err)
	assert.Equal(t, 5, readonly.Priority)
}
