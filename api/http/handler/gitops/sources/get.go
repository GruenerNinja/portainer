package sources

import (
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	sourceDS "github.com/portainer/portainer/api/dataservices/source"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type gitAuthInfo struct {
	Username string `json:"username"`
}

type connectionInfo struct {
	TLSSkipVerify  bool         `json:"tlsSkipVerify"`
	Authentication *gitAuthInfo `json:"authentication,omitempty"`
	Vault          *vaultInfo   `json:"vault,omitempty"`
}

type vaultInfo struct {
	Address       string `json:"address"`
	TLSSkipVerify bool   `json:"tlsSkipVerify"`
	Namespace     string `json:"namespace,omitempty"`
	KVVersion     int    `json:"kvVersion"`
	AuthMethod    string `json:"authMethod"`
}

type SourceAccess struct {
	Public bool               `json:"public,omitempty"`
	Users  []portainer.UserID `json:"users,omitempty"`
	Teams  []portainer.TeamID `json:"teams,omitempty"`
}

// SourceDetail is the get-by-id response for a GitOps source
type SourceDetail struct {
	SourceBase
	Connection connectionInfo `json:"connection" validate:"required"`
	Access     SourceAccess   `json:"access"`
}

// @id GitOpsSourceGet
// @summary Get a GitOps source by ID
// @description Returns a single GitOps source with its connection settings and access rules.
// @description **Access policy**: authenticated
// @tags gitops
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Source identifier"
// @success 200 {object} SourceDetail
// @failure 400 "Invalid request"
// @failure 403 "Access denied"
// @failure 404 "Source not found"
// @failure 500 "Server error"
// @router /gitops/sources/{id} [get]
func (h *Handler) getSource(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	srcID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid source identifier route variable", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	sourceID := portainer.SourceID(srcID)

	var source *portainer.Source

	err = h.dataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		userContext := sourceDS.NewUserContext(securityContext.User, securityContext.UserMemberships)
		var handlerErr *httperror.HandlerError
		source, handlerErr = ReadSource(tx, userContext, sourceID)
		if handlerErr != nil {
			return handlerErr
		}
		return nil
	})

	return response.TxFuncResponse(err, func() *httperror.HandlerError {
		access := BuildSourceAccess(source)
		detail := BuildSourceDetail(h.buildSourceBase(source), source, access)
		return response.JSON(w, detail)
	})
}

func BuildSourceDetail(baseSource SourceBase, src *portainer.Source, access SourceAccess) SourceDetail {
	return SourceDetail{
		SourceBase: baseSource,
		Connection: buildConnectionInfo(src),
		Access:     access,
	}
}

func BuildSourceAccess(source *portainer.Source) SourceAccess {
	if source == nil {
		return SourceAccess{}
	}

	if source.AdministratorsOnly {
		return SourceAccess{}
	}

	if source.Public {
		return SourceAccess{
			Public: true,
		}
	}

	return SourceAccess{
		Public: source.Public,
		Users:  source.UserAccesses,
		Teams:  source.TeamAccesses,
	}
}

func buildConnectionInfo(src *portainer.Source) connectionInfo {
	if src == nil {
		return connectionInfo{}
	}
	cfg := src.Git
	if src.Vault != nil {
		return connectionInfo{
			TLSSkipVerify: src.Vault.TLSSkipVerify,
			Vault: &vaultInfo{
				Address:       src.Vault.Address,
				TLSSkipVerify: src.Vault.TLSSkipVerify,
				Namespace:     src.Vault.Namespace,
				KVVersion:     src.Vault.KVVersion,
				AuthMethod:    src.Vault.Authentication.Method,
			},
		}
	}
	if cfg == nil {
		return connectionInfo{}
	}
	return buildGitConnectionInfo(cfg)
}

func buildGitConnectionInfo(cfg *gittypes.GitSource) connectionInfo {
	if cfg == nil {
		return connectionInfo{}
	}
	return connectionInfo{
		TLSSkipVerify:  cfg.TLSSkipVerify,
		Authentication: buildGitAuthInfo(cfg.Authentication),
	}
}

func buildGitAuthInfo(auth *gittypes.GitAuthentication) *gitAuthInfo {
	if auth == nil {
		return nil
	}
	return &gitAuthInfo{
		Username: auth.Username,
	}
}
