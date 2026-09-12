package resourcecontrols

import (
	"errors"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type resourceControlUpdatePayload struct {
	// Permit access to the associated resource to any user
	Public bool `example:"true"`
	// List of user identifiers with access to the associated resource
	Users []int `example:"4"`
	// List of team identifiers with access to the associated resource
	Teams []int `example:"7"`
	// List of user identifiers with read-only access to a stack
	ReadOnlyUsers []int `example:"8,9"`
	// List of team identifiers with read-only access to a stack
	ReadOnlyTeams []int `example:"10,11"`
	// Permit access to resource only to admins
	AdministratorsOnly bool `example:"true"`
}

func (payload *resourceControlUpdatePayload) Validate(r *http.Request) error {
	if len(payload.Users) == 0 && len(payload.Teams) == 0 && !payload.Public && !payload.AdministratorsOnly {
		return errors.New("invalid payload: must specify Users, Teams, Public or AdministratorsOnly")
	}

	if payload.Public && payload.AdministratorsOnly {
		return errors.New("invalid payload: cannot set public and administrators only")
	}

	if err := validateDistinctResourceAccesses(payload.Users, payload.ReadOnlyUsers, payload.Teams, payload.ReadOnlyTeams); err != nil {
		return err
	}

	return nil
}

// @id ResourceControlUpdate
// @summary Update a resource control
// @description Update a resource control
// @description **Access policy**: authenticated
// @tags resource_controls
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Resource control identifier"
// @param body body resourceControlUpdatePayload true "Resource control details"
// @success 200 {object} portainer.ResourceControl "Success"
// @failure 400 "Invalid request"
// @failure 403 "Unauthorized"
// @failure 404 "Resource control not found"
// @failure 500 "Server error"
// @router /resource_controls/{id} [put]
func (handler *Handler) resourceControlUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	resourceControlID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid resource control identifier route variable", err)
	}

	var payload resourceControlUpdatePayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	resourceControl, err := handler.DataStore.ResourceControl().Read(portainer.ResourceControlID(resourceControlID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a resource control with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a resource control with the specified identifier inside the database", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	if !security.AuthorizedResourceControlAccess(resourceControl, securityContext) {
		return httperror.Forbidden("Permission denied to access the resource control", httperrors.ErrResourceAccessDenied)
	}

	if (len(payload.ReadOnlyUsers) > 0 || len(payload.ReadOnlyTeams) > 0) && resourceControl.Type != portainer.StackResourceControl {
		return httperror.BadRequest("Read-only access is only supported for stacks", errors.New("read-only access is only supported for stacks"))
	}

	if !securityContext.IsAdmin {
		existingReadOnlyUsers, existingReadOnlyTeams := readOnlyAccessIDs(resourceControl)
		if !equalIntSets(existingReadOnlyUsers, payload.ReadOnlyUsers) || !equalIntSets(existingReadOnlyTeams, payload.ReadOnlyTeams) {
			return httperror.Forbidden("Only administrators can change read-only access", httperrors.ErrResourceAccessDenied)
		}
	}

	resourceControl.Public = payload.Public
	resourceControl.AdministratorsOnly = payload.AdministratorsOnly

	var userAccesses = make([]portainer.UserResourceAccess, 0)
	for _, v := range payload.Users {
		userAccess := portainer.UserResourceAccess{
			UserID:      portainer.UserID(v),
			AccessLevel: portainer.ReadWriteAccessLevel,
		}
		userAccesses = append(userAccesses, userAccess)
	}
	for _, v := range payload.ReadOnlyUsers {
		userAccesses = append(userAccesses, portainer.UserResourceAccess{
			UserID:      portainer.UserID(v),
			AccessLevel: portainer.ReadOnlyAccessLevel,
		})
	}
	resourceControl.UserAccesses = userAccesses

	var teamAccesses = make([]portainer.TeamResourceAccess, 0)
	for _, v := range payload.Teams {
		teamAccess := portainer.TeamResourceAccess{
			TeamID:      portainer.TeamID(v),
			AccessLevel: portainer.ReadWriteAccessLevel,
		}
		teamAccesses = append(teamAccesses, teamAccess)
	}
	for _, v := range payload.ReadOnlyTeams {
		teamAccesses = append(teamAccesses, portainer.TeamResourceAccess{
			TeamID:      portainer.TeamID(v),
			AccessLevel: portainer.ReadOnlyAccessLevel,
		})
	}
	resourceControl.TeamAccesses = teamAccesses

	if !security.AuthorizedResourceControlUpdate(resourceControl, securityContext) {
		return httperror.Forbidden("Permission denied to update the resource control", httperrors.ErrResourceAccessDenied)
	}

	if err := handler.DataStore.ResourceControl().Update(resourceControl.ID, resourceControl); err != nil {
		return httperror.InternalServerError("Unable to persist resource control changes inside the database", err)
	}

	return response.JSON(w, resourceControl)
}
