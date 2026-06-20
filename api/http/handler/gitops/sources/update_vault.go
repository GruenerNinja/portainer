package sources

import (
	"errors"
	"strings"

	portainer "github.com/portainer/portainer/api"
)

var ErrNotVaultSource = errors.New("source is not a Vault source")

type VaultSourceUpdatePayload struct {
	Name           *string                           `json:"name"`
	Address        *string                           `json:"address"`
	TLSSkipVerify  *bool                             `json:"tlsSkipVerify"`
	Namespace      *string                           `json:"namespace"`
	KVVersion      *int                              `json:"kvVersion"`
	Authentication *VaultAuthenticationUpdatePayload `json:"authentication"`
}

type VaultAuthenticationUpdatePayload struct {
	Method *string `json:"method"`
	Token  *string `json:"token"`
}

func ApplyVaultSourceChanges(src *portainer.Source, payload VaultSourceUpdatePayload) error {
	if src.Type != portainer.SourceTypeVault {
		return ErrNotVaultSource
	}

	if payload.Name != nil && strings.TrimSpace(*payload.Name) != "" {
		src.Name = strings.TrimSpace(*payload.Name)
	}

	if src.Vault == nil {
		src.Vault = &portainer.VaultConfig{
			KVVersion: 2,
			Authentication: portainer.VaultAuthentication{
				Method: "token",
			},
		}
	}

	if payload.Address != nil {
		src.Vault.Address = strings.TrimSpace(*payload.Address)
	}
	if payload.TLSSkipVerify != nil {
		src.Vault.TLSSkipVerify = *payload.TLSSkipVerify
	}
	if payload.Namespace != nil {
		src.Vault.Namespace = strings.TrimSpace(*payload.Namespace)
	}
	if payload.KVVersion != nil {
		src.Vault.KVVersion = *payload.KVVersion
	}
	if src.Vault.KVVersion == 0 {
		src.Vault.KVVersion = 2
	}

	if payload.Authentication != nil {
		if payload.Authentication.Method != nil {
			src.Vault.Authentication.Method = *payload.Authentication.Method
		}
		if payload.Authentication.Token != nil {
			src.Vault.Authentication.Token = *payload.Authentication.Token
		}
	}
	if src.Vault.Authentication.Method == "" {
		src.Vault.Authentication.Method = "token"
	}

	return nil
}
