package deployments

import (
	"context"
	"fmt"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	sourceDS "github.com/portainer/portainer/api/dataservices/source"
	"github.com/portainer/portainer/api/gitops/secrets"
)

func stackWithResolvedSecrets(ctx context.Context, dataStore dataservices.DataStore, stack *portainer.Stack) (*portainer.Stack, error) {
	if stack == nil || len(stack.SecretMappings) == 0 {
		return stack, nil
	}

	resolvedEnv := append([]portainer.Pair{}, stack.Env...)
	indexByName := make(map[string]int, len(resolvedEnv))
	for i, pair := range resolvedEnv {
		indexByName[pair.Name] = i
	}

	for _, mapping := range stack.SecretMappings {
		name := strings.TrimSpace(mapping.Name)
		if name == "" {
			return nil, fmt.Errorf("secret mapping environment variable name is required")
		}

		src, err := dataStore.Source().Read(sourceDS.InsecureNewAdminContext(), mapping.SourceID)
		if dataStore.IsErrObjectNotFound(err) {
			return nil, fmt.Errorf("secret source %d was not found", mapping.SourceID)
		} else if err != nil {
			return nil, fmt.Errorf("failed to read secret source %d: %w", mapping.SourceID, err)
		}

		if src.Type != portainer.SourceTypeVault || src.Vault == nil {
			return nil, fmt.Errorf("secret source %d is not a Vault source", mapping.SourceID)
		}

		value, err := secrets.ResolveVaultSecret(ctx, src.Vault, mapping.Path, mapping.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve secret %s from source %d: %w", name, mapping.SourceID, err)
		}

		if idx, ok := indexByName[name]; ok {
			resolvedEnv[idx].Value = value
		} else {
			indexByName[name] = len(resolvedEnv)
			resolvedEnv = append(resolvedEnv, portainer.Pair{Name: name, Value: value})
		}
	}

	resolvedStack := *stack
	resolvedStack.Env = resolvedEnv

	return &resolvedStack, nil
}
