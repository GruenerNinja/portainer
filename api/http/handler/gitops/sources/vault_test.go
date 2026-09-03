package sources

import (
	"testing"

	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVaultSourceStoresInternalAddress(t *testing.T) {
	t.Parallel()

	src := BuildVaultSource(VaultSourceCreatePayload{
		Name:            "production-vault",
		Address:         " https://vault.example.com ",
		InternalAddress: " http://vault:8200 ",
		Authentication: VaultAuthenticationPayload{
			Method: "token",
			Token:  "token-value",
		},
	})

	require.NotNil(t, src.Vault)
	assert.Equal(t, "https://vault.example.com", src.Vault.Address)
	assert.Equal(t, "http://vault:8200", src.Vault.InternalAddress)
	assert.Equal(t, 2, src.Vault.KVVersion)
}

func TestApplyVaultSourceChangesCanUpdateAndClearInternalAddress(t *testing.T) {
	t.Parallel()

	src := &portainer.Source{
		Type: portainer.SourceTypeVault,
		Vault: &portainer.VaultConfig{
			Address:         "https://vault.example.com",
			InternalAddress: "http://old-vault:8200",
		},
	}
	empty := ""

	require.NoError(t, ApplyVaultSourceChanges(src, VaultSourceUpdatePayload{
		InternalAddress: &empty,
	}))
	require.Empty(t, src.Vault.InternalAddress)
}

func TestBuildSourceDetailIncludesInternalVaultAddress(t *testing.T) {
	t.Parallel()

	detail := BuildSourceDetail(SourceBase{}, &portainer.Source{
		Type: portainer.SourceTypeVault,
		Vault: &portainer.VaultConfig{
			Address:         "https://vault.example.com",
			InternalAddress: "http://vault:8200",
		},
	}, SourceAccess{})

	require.NotNil(t, detail.Connection.Vault)
	assert.Equal(t, "http://vault:8200", detail.Connection.Vault.InternalAddress)
}
