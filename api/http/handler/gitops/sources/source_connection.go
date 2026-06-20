package sources

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/api/gitops/secrets"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id GitOpsSourcesTestById
// @summary Test the connection of a stored source
// @description Tests connectivity for a GitOps source, applying optional overrides to the stored configuration.
// @description **Access policy**: administrator
// @tags gitops
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Source identifier"
// @param body body GitSourceUpdatePayload false "Optional connection overrides; omitted fields fall back to stored values"
// @success 200 {object} ConnectionTestResult "Connection test result"
// @failure 400 "Invalid request payload"
// @failure 403 "Access denied"
// @failure 404 "Source not found"
// @failure 500 "Server error"
// @router /gitops/sources/{id}/test [post]
func (h *Handler) sourceTestConnection(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	sourceID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid source identifier route variable", err)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var src *portainer.Source
	if err := h.dataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		src, err = tx.Source().Read(portainer.SourceID(sourceID))
		return err
	}); h.dataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a source with the specified identifier", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find source", err)
	}

	switch src.Type {
	case portainer.SourceTypeGit:
		var payload GitSourceUpdatePayload
		if len(body) > 0 {
			if err := json.Unmarshal(body, &payload); err != nil && !errors.Is(err, io.EOF) {
				return httperror.BadRequest("Invalid request payload", err)
			}
		}
		if err := ApplyGitSourceChanges(src, payload); errors.Is(err, ErrNotGitSource) {
			return httperror.BadRequest("Source is not a Git source", err)
		} else if err != nil {
			return httperror.InternalServerError("Unable to apply source changes", err)
		}
		if src.Git == nil {
			return httperror.InternalServerError("Source has no git configuration", nil)
		}
		return response.JSON(w, testSourceConnection(r.Context(), h.gitService, src.Git))
	case portainer.SourceTypeVault:
		var payload VaultSourceUpdatePayload
		if len(body) > 0 {
			if err := json.Unmarshal(body, &payload); err != nil && !errors.Is(err, io.EOF) {
				return httperror.BadRequest("Invalid request payload", err)
			}
		}
		if err := ApplyVaultSourceChanges(src, payload); errors.Is(err, ErrNotVaultSource) {
			return httperror.BadRequest("Source is not a Vault source", err)
		} else if err != nil {
			return httperror.InternalServerError("Unable to apply source changes", err)
		}
		return response.JSON(w, testVaultSourceConnection(r.Context(), src.Vault))
	default:
		return httperror.BadRequest("Unsupported source type", nil)
	}
}

type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// @id GitOpsSourcesTest
// @summary Test a Git source connection
// @description Tests connectivity for Git connection details that have not been persisted yet.
// @description **Access policy**: administrator
// @tags gitops
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body GitSourceCreatePayload true "Git connection details"
// @success 200 {object} ConnectionTestResult "Connection test result"
// @failure 400 "Invalid request payload"
// @failure 403 "Access denied"
// @failure 500 "Server error"
// @router /gitops/sources/test [post]
func (h *Handler) gitSourceTest(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload GitSourceCreatePayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	src, err := BuildGitSource(payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}
	if src.Git == nil {
		return httperror.InternalServerError("Source has no git configuration", nil)
	}

	result := testSourceConnection(r.Context(), h.gitService, src.Git)

	return response.JSON(w, result)
}

// @id GitOpsSourcesVaultTest
// @summary Test a Vault source connection
// @description Tests connectivity for Vault connection details that have not been persisted yet.
// @description **Access policy**: administrator
// @tags gitops
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body VaultSourceCreatePayload true "Vault connection details"
// @success 200 {object} ConnectionTestResult "Connection test result"
// @failure 400 "Invalid request payload"
// @failure 403 "Access denied"
// @failure 500 "Server error"
// @router /gitops/sources/vault/test [post]
func (h *Handler) vaultSourceTest(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload VaultSourceCreatePayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	src := BuildVaultSource(payload)
	return response.JSON(w, testVaultSourceConnection(r.Context(), src.Vault))
}

// testSourceConnection verifies that a git repository is reachable with the given config.
func testSourceConnection(ctx context.Context, gitService portainer.GitService, config *gittypes.RepoConfig) ConnectionTestResult {
	var username, password string
	if config.Authentication != nil {
		username = config.Authentication.Username
		password = config.Authentication.Password
	}

	_, err := gitService.ListRefs(ctx, config.URL, username, password, false, config.TLSSkipVerify)
	if err != nil {
		return ConnectionTestResult{Success: false, Error: err.Error()}
	}

	return ConnectionTestResult{Success: true}
}

func testVaultSourceConnection(ctx context.Context, config *portainer.VaultConfig) ConnectionTestResult {
	if err := secrets.TestVaultConnection(ctx, config); err != nil {
		return ConnectionTestResult{Success: false, Error: err.Error()}
	}

	return ConnectionTestResult{Success: true}
}
