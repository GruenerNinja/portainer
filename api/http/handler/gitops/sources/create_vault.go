package sources

import (
	"errors"
	"net/http"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type VaultAuthenticationPayload struct {
	Method string `json:"method"`
	Token  string `json:"token"`
}

type VaultSourceCreatePayload struct {
	Name           string                     `json:"name"`
	Address        string                     `json:"address" validate:"required"`
	TLSSkipVerify  bool                       `json:"tlsSkipVerify"`
	Namespace      string                     `json:"namespace"`
	KVVersion      int                        `json:"kvVersion"`
	Authentication VaultAuthenticationPayload `json:"authentication"`
}

func (payload *VaultSourceCreatePayload) Validate(_ *http.Request) error {
	if strings.TrimSpace(payload.Address) == "" {
		return errors.New("address is required")
	}
	if payload.Authentication.Method == "" {
		return errors.New("authentication method is required")
	}
	if payload.Authentication.Method != "token" {
		return errors.New("only token authentication is supported")
	}
	if strings.TrimSpace(payload.Authentication.Token) == "" {
		return errors.New("token is required")
	}
	return nil
}

// @id GitOpsSourcesCreateVault
// @summary Create a Vault source
// @description Creates a reusable HashiCorp Vault source for deploy-time secret resolution.
// @description **Access policy**: administrator
// @tags gitops
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body VaultSourceCreatePayload true "Vault source details"
// @success 201 {object} portainer.Source
// @failure 400 "Invalid request payload"
// @failure 403 "Access denied"
// @failure 500 "Server error"
// @router /gitops/sources/vault [post]
func (h *Handler) vaultSourceCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload VaultSourceCreatePayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	src := BuildVaultSource(payload)
	if err := h.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		return tx.Source().Create(src)
	}); err != nil {
		return httperror.InternalServerError("Unable to create source", err)
	}

	redactVaultSource(src)
	return response.JSONWithStatus(w, src, http.StatusCreated)
}

func BuildVaultSource(payload VaultSourceCreatePayload) *portainer.Source {
	kvVersion := payload.KVVersion
	if kvVersion == 0 {
		kvVersion = 2
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		name = strings.TrimSpace(payload.Address)
	}

	return &portainer.Source{
		Name: name,
		Type: portainer.SourceTypeVault,
		Vault: &portainer.VaultConfig{
			Address:       strings.TrimSpace(payload.Address),
			TLSSkipVerify: payload.TLSSkipVerify,
			Namespace:     strings.TrimSpace(payload.Namespace),
			KVVersion:     kvVersion,
			Authentication: portainer.VaultAuthentication{
				Method: payload.Authentication.Method,
				Token:  payload.Authentication.Token,
			},
		},
	}
}

func redactVaultSource(src *portainer.Source) {
	if src == nil || src.Vault == nil {
		return
	}
	src.Vault.Authentication.Token = ""
}
