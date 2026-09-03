package secrets

import (
	"context"
	"fmt"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	sourceDS "github.com/portainer/portainer/api/dataservices/source"

	"github.com/rs/zerolog/log"
)

const vaultTokenRenewalCheckInterval = time.Hour

type vaultTokenSourceStore interface {
	Source() dataservices.SourceService
}

type VaultTokenRenewer struct {
	dataStore     vaultTokenSourceStore
	checkInterval time.Duration
}

func NewVaultTokenRenewer(dataStore vaultTokenSourceStore) *VaultTokenRenewer {
	return &VaultTokenRenewer{
		dataStore:     dataStore,
		checkInterval: vaultTokenRenewalCheckInterval,
	}
}

func (renewer *VaultTokenRenewer) Start(ctx context.Context) {
	if renewer == nil || renewer.dataStore == nil {
		return
	}

	go func() {
		renewer.runOnce(ctx)

		ticker := time.NewTicker(renewer.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				renewer.runOnce(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (renewer *VaultTokenRenewer) runOnce(ctx context.Context) {
	sources, err := renewer.dataStore.Source().ReadAll(sourceDS.InsecureNewAdminContext(), func(src portainer.Source) bool {
		return src.Type == portainer.SourceTypeVault && src.Vault != nil
	})
	if err != nil {
		log.Warn().Err(fmt.Errorf("failed to read Vault sources: %w", err)).Msg("Vault token renewal check failed")
		return
	}

	for i := range sources {
		if ctx.Err() != nil {
			return
		}

		result, err := RenewVaultTokenIfNeeded(ctx, sources[i].Vault)
		if err != nil {
			log.Warn().Err(err).Int("source_id", int(sources[i].ID)).Msg("Vault token renewal check failed")
			continue
		}

		if result.Renewed {
			log.Info().Int("source_id", int(sources[i].ID)).Int64("ttl_seconds", result.TTL).Msg("renewed periodic Vault token")
		}
	}
}
