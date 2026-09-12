package roles

import (
	"net/http"

	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	librequest "github.com/portainer/portainer/pkg/libhttp/request"
	libresponse "github.com/portainer/portainer/pkg/libhttp/response"
)

// @id RoleInspect
// @summary Inspect a role
// @description **Access policy**: administrator
// @tags roles
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Role identifier"
// @success 200 {object} portainer.Role "Success"
// @failure 404 "Role not found"
// @router /roles/{id} [get]
func (handler *Handler) roleInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := librequest.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid role identifier", err)
	}
	role, err := handler.DataStore.Role().Read(portainer.RoleID(id))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the role", err)
	}
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the role", err)
	}
	return libresponse.JSON(w, role)
}
