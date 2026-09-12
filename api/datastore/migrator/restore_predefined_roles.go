package migrator

import (
	"reflect"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/internal/authorization"
)

// restorePredefinedRoles keeps the fork's complete environment-role set
// available for new databases and databases upgraded from Community Edition.
func (m *Migrator) restorePredefinedRoles() error {
	settings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}

	volumeBrowsing := settings.AllowVolumeBrowserForRegularUsers
	roles := []*portainer.Role{
		{
			ID:             portainer.RoleIDEndpointAdmin,
			Name:           "Environment administrator",
			Description:    "Full control of all resources in an environment",
			Priority:       1,
			Authorizations: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
		},
		{
			ID:             portainer.RoleIDOperator,
			Name:           "Operator",
			Description:    "Operational control of all existing resources in an environment",
			Priority:       2,
			Authorizations: authorization.DefaultEndpointAuthorizationsForOperatorRole(volumeBrowsing),
		},
		{
			ID:             portainer.RoleIDHelpdesk,
			Name:           "Helpdesk",
			Description:    "Read-only access of all resources in an environment",
			Priority:       3,
			Authorizations: authorization.DefaultEndpointAuthorizationsForHelpDeskRole(volumeBrowsing),
		},
		{
			ID:             portainer.RoleIDStandardUser,
			Name:           "Standard user",
			Description:    "Full control of assigned resources in an environment",
			Priority:       4,
			Authorizations: authorization.DefaultEndpointAuthorizationsForStandardUserRole(volumeBrowsing),
		},
		{
			ID:             portainer.RoleIDReadonly,
			Name:           "Read-only user",
			Description:    "Read-only access of assigned resources in an environment",
			Priority:       5,
			Authorizations: authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole(volumeBrowsing),
		},
	}

	for _, expected := range roles {
		current, err := m.roleService.Read(expected.ID)
		if err == nil && reflect.DeepEqual(current, expected) {
			continue
		}
		if err != nil && !dataservices.IsErrObjectNotFound(err) {
			return err
		}

		if err := m.roleService.Update(expected.ID, expected); err != nil {
			return err
		}
	}

	return nil
}
