package auth

import (
	"context"
	"errors"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	httperrors "github.com/portainer/portainer/api/http/errors"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"

	"github.com/rs/zerolog/log"
)

type oauthPayload struct {
	// OAuth code returned from OAuth Provided
	Code string
}

func (payload *oauthPayload) Validate(r *http.Request) error {
	if len(payload.Code) == 0 {
		return errors.New("Invalid OAuth authorization code")
	}

	return nil
}

func (handler *Handler) authenticateOAuth(ctx context.Context, code string, settings *portainer.OAuthSettings) (*portainer.OAuthIdentity, error) {
	if code == "" {
		return nil, errors.New("Invalid OAuth authorization code")
	}

	if settings == nil {
		return nil, errors.New("Invalid OAuth configuration")
	}

	identity, err := handler.OAuthService.Authenticate(ctx, code, settings)
	if err != nil {
		return nil, err
	}

	return identity, nil
}

// @id ValidateOAuth
// @summary Authenticate with OAuth
// @description **Access policy**: public
// @tags auth
// @accept json
// @produce json
// @param body body oauthPayload true "OAuth Credentials used for authentication"
// @success 200 {object} authenticateResponse "Success"
// @failure 400 "Invalid request"
// @failure 422 "Invalid Credentials"
// @failure 500 "Server error"
// @router /auth/oauth/validate [post]
func (handler *Handler) validateOAuth(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload oauthPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var settings *portainer.Settings
	if err := handler.DataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		var err error
		settings, err = tx.Settings().Settings()
		return err
	}); err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	if settings.AuthenticationMethod != portainer.AuthenticationOAuth {
		return httperror.Forbidden("OAuth authentication is not enabled", errors.New("OAuth authentication is not enabled"))
	}

	identity, err := handler.authenticateOAuth(r.Context(), payload.Code, &settings.OAuthSettings)
	if err != nil {
		log.Debug().Err(err).Msg("OAuth authentication error")

		return httperror.InternalServerError("Unable to authenticate through OAuth", httperrors.ErrUnauthorized)
	}
	username := identity.Username

	user, err := handler.DataStore.User().UserByUsername(username)
	if err != nil && !handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.InternalServerError("Unable to retrieve a user with the specified username from the database", err)
	}

	if user == nil && !settings.OAuthSettings.OAuthAutoCreateUsers {
		return httperror.Forbidden("Account not created beforehand in Portainer and automatic user provisioning not enabled", httperrors.ErrUnauthorized)
	}

	if user == nil {
		user = &portainer.User{
			Username: username,
			Role:     portainer.StandardUserRole,
		}

		err = handler.DataStore.User().Create(user)
		if err != nil {
			return httperror.InternalServerError("Unable to persist user inside the database", err)
		}

		if settings.OAuthSettings.DefaultTeamID != 0 && !settings.OAuthSettings.OAuthAutoMapTeamMemberships {
			membership := &portainer.TeamMembership{
				UserID: user.ID,
				TeamID: settings.OAuthSettings.DefaultTeamID,
				Role:   portainer.TeamMember,
			}

			err = handler.DataStore.TeamMembership().Create(membership)
			if err != nil {
				return httperror.InternalServerError("Unable to persist team membership inside the database", err)
			}
		}

	}

	if settings.OAuthSettings.OAuthAutoMapTeamMemberships {
		if err := handler.syncOAuthTeamMemberships(user.ID, &settings.OAuthSettings, identity.Claims); err != nil {
			return httperror.InternalServerError("Unable to synchronize OAuth team memberships", err)
		}
	}

	return handler.writeToken(w, r, user, false, settings.ForceSecureCookies)
}
