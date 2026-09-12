package roles

import (
	"errors"
	"net/http"
	"strings"

	portainer "github.com/portainer/portainer/api"
)

const firstCustomRoleID portainer.RoleID = portainer.RoleIDOperator + 1

type rolePayload struct {
	Name           string                   `json:"Name" validate:"required" example:"Automation reader"`
	Description    string                   `json:"Description" example:"Read-only access for automation accounts"`
	Authorizations portainer.Authorizations `json:"Authorizations"`
	Priority       int                      `json:"Priority" validate:"required,min=1,max=1000" example:"5"`
}

func (payload *rolePayload) Validate(_ *http.Request) error {
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Description = strings.TrimSpace(payload.Description)
	if payload.Name == "" {
		return errors.New("role name is required")
	}
	if len(payload.Name) > 128 {
		return errors.New("role name cannot exceed 128 characters")
	}
	if payload.Priority < 1 || payload.Priority > 1000 {
		return errors.New("role priority must be between 1 and 1000")
	}
	if payload.Authorizations == nil {
		payload.Authorizations = portainer.Authorizations{}
	}

	// False entries have no effect in the resolver. Dropping them keeps the
	// persisted role compact and makes API responses unambiguous.
	for authorization, enabled := range payload.Authorizations {
		if !enabled {
			delete(payload.Authorizations, authorization)
		}
	}

	return nil
}

func isPredefinedRole(id portainer.RoleID) bool {
	return id > 0 && id < firstCustomRoleID
}

func roleNameExists(roles []portainer.Role, name string, excluding portainer.RoleID) bool {
	for _, role := range roles {
		if role.ID != excluding && strings.EqualFold(strings.TrimSpace(role.Name), name) {
			return true
		}
	}
	return false
}
