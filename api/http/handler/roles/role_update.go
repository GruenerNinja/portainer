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

// @id RoleUpdate
// @summary Update a custom role
// @description Built-in roles cannot be modified.
// @description **Access policy**: administrator
// @tags roles
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Role identifier"
// @param body body rolePayload true "Role details"
// @success 200 {object} portainer.Role "Success"
// @failure 400 "Invalid request"
// @failure 403 "Built-in role"
// @failure 404 "Role not found"
// @failure 409 "Role name already exists"
// @router /roles/{id} [put]
func (handler *Handler) roleUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := librequest.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid role identifier", err)
	}
	roleID := portainer.RoleID(id)
	if isPredefinedRole(roleID) {
		return httperror.Forbidden("Built-in roles cannot be modified", errors.New("predefined role"))
	}
	var payload rolePayload
	if err := librequest.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var role *portainer.Role
	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		existing, err := tx.Role().Read(roleID)
		if tx.IsErrObjectNotFound(err) {
			return httperror.NotFound("Unable to find the role", err)
		}
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve the role", err)
		}
		roles, err := tx.Role().ReadAll()
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve roles from the database", err)
		}
		if roleNameExists(roles, payload.Name, roleID) {
			return httperror.Conflict("A role with this name already exists", errors.New("duplicate role name"))
		}
		existing.Name = payload.Name
		existing.Description = payload.Description
		existing.Authorizations = payload.Authorizations
		existing.Priority = payload.Priority
		if err := tx.Role().Update(roleID, existing); err != nil {
			return httperror.InternalServerError("Unable to persist the role", err)
		}
		role = existing
		return nil
	})

	return libresponse.TxResponse(w, role, err)
}
