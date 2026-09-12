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

// @id RoleDelete
// @summary Delete a custom role
// @description Built-in or assigned roles cannot be deleted.
// @description **Access policy**: administrator
// @tags roles
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Role identifier"
// @success 204 "Success"
// @failure 403 "Built-in role"
// @failure 404 "Role not found"
// @failure 409 "Role is in use"
// @router /roles/{id} [delete]
func (handler *Handler) roleDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := librequest.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid role identifier", err)
	}
	roleID := portainer.RoleID(id)
	if isPredefinedRole(roleID) {
		return httperror.Forbidden("Built-in roles cannot be deleted", errors.New("predefined role"))
	}

	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		if _, err := tx.Role().Read(roleID); tx.IsErrObjectNotFound(err) {
			return httperror.NotFound("Unable to find the role", err)
		} else if err != nil {
			return httperror.InternalServerError("Unable to retrieve the role", err)
		}
		inUse, err := roleIsInUse(tx, roleID)
		if err != nil {
			return err
		}
		if inUse {
			return httperror.Conflict("The role is assigned to one or more access policies", errors.New("role is in use"))
		}
		if err := tx.Role().Delete(roleID); err != nil {
			return httperror.InternalServerError("Unable to delete the role", err)
		}
		return nil
	})

	return libresponse.TxEmptyResponse(w, err)
}

func roleIsInUse(tx dataservices.DataStoreTx, roleID portainer.RoleID) (bool, error) {
	endpoints, err := tx.Endpoint().ReadAll()
	if err != nil {
		return false, httperror.InternalServerError("Unable to inspect environment access policies", err)
	}
	for _, endpoint := range endpoints {
		if policiesContainRole(endpoint.UserAccessPolicies, endpoint.TeamAccessPolicies, roleID) {
			return true, nil
		}
	}

	groups, err := tx.EndpointGroup().ReadAll()
	if err != nil {
		return false, httperror.InternalServerError("Unable to inspect environment group access policies", err)
	}
	for _, group := range groups {
		if policiesContainRole(group.UserAccessPolicies, group.TeamAccessPolicies, roleID) {
			return true, nil
		}
	}

	registries, err := tx.Registry().ReadAll()
	if err != nil {
		return false, httperror.InternalServerError("Unable to inspect registry access policies", err)
	}
	for _, registry := range registries {
		if policiesContainRole(registry.UserAccessPolicies, registry.TeamAccessPolicies, roleID) {
			return true, nil
		}
		for _, access := range registry.RegistryAccesses {
			if policiesContainRole(access.UserAccessPolicies, access.TeamAccessPolicies, roleID) {
				return true, nil
			}
		}
	}

	return false, nil
}

func policiesContainRole(users portainer.UserAccessPolicies, teams portainer.TeamAccessPolicies, roleID portainer.RoleID) bool {
	for _, policy := range users {
		if policy.RoleID == roleID {
			return true
		}
	}
	for _, policy := range teams {
		if policy.RoleID == roleID {
			return true
		}
	}
	return false
}
