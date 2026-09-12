package roles

import (
	"errors"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	librequest "github.com/portainer/portainer/pkg/libhttp/request"
	libresponse "github.com/portainer/portainer/pkg/libhttp/response"
)

// @id RoleCreate
// @summary Create a custom role
// @description Create a custom environment role. Built-in roles remain managed by Portainer.
// @description **Access policy**: administrator
// @tags roles
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body rolePayload true "Role details"
// @success 200 {object} portainer.Role "Success"
// @failure 400 "Invalid request"
// @failure 409 "Role name already exists"
// @failure 500 "Server error"
// @router /roles [post]
func (handler *Handler) roleCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload rolePayload
	if err := librequest.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var role *portainer.Role
	err := handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		roles, err := tx.Role().ReadAll()
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve roles from the database", err)
		}
		if roleNameExists(roles, payload.Name, 0) {
			return httperror.Conflict("A role with this name already exists", errors.New("duplicate role name"))
		}

		nextID := firstCustomRoleID
		for _, existing := range roles {
			if existing.ID >= nextID {
				nextID = existing.ID + 1
			}
		}

		role = &portainer.Role{
			ID: nextID, Name: payload.Name, Description: payload.Description,
			Authorizations: payload.Authorizations, Priority: payload.Priority,
		}
		if err := tx.Role().Update(role.ID, role); err != nil {
			return httperror.InternalServerError("Unable to persist the role", err)
		}
		return nil
	})

	return libresponse.TxResponse(w, role, err)
}
