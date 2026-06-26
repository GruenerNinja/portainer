package deployments

import (
	"context"
	"fmt"
	"sort"
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
		src, err := dataStore.Source().Read(sourceDS.InsecureNewAdminContext(), mapping.SourceID)
		if dataStore.IsErrObjectNotFound(err) {
			return nil, fmt.Errorf("secret source %d was not found", mapping.SourceID)
		} else if err != nil {
			return nil, fmt.Errorf("failed to read secret source %d: %w", mapping.SourceID, err)
		}

		if src.Type != portainer.SourceTypeVault || src.Vault == nil {
			return nil, fmt.Errorf("secret source %d is not a Vault source", mapping.SourceID)
		}

		key := strings.TrimSpace(mapping.Key)
		if key == "" {
			values, err := secrets.ResolveVaultSecretValues(ctx, src.Vault, mapping.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secrets from source %d: %w", mapping.SourceID, err)
			}

			keys := make([]string, 0, len(values))
			for name := range values {
				keys = append(keys, name)
			}
			sort.Strings(keys)

			for _, secretKey := range keys {
				name := strings.TrimSpace(secretKey)
				if name == "" {
					return nil, fmt.Errorf("secret mapping environment variable name is required")
				}
				upsertEnv(&resolvedEnv, indexByName, name, values[secretKey])
			}

			continue
		}

		name := strings.TrimSpace(mapping.Name)
		if name == "" {
			name = key
		}

		value, err := secrets.ResolveVaultSecret(ctx, src.Vault, mapping.Path, key)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve secret %s from source %d: %w", name, mapping.SourceID, err)
		}

		upsertEnv(&resolvedEnv, indexByName, name, value)
	}

	resolvedStack := *stack
	resolvedStack.Env = resolvedEnv

	return &resolvedStack, nil
}

func upsertEnv(env *[]portainer.Pair, indexByName map[string]int, name, value string) {
	if idx, ok := indexByName[name]; ok {
		(*env)[idx].Value = value
		return
	}

	indexByName[name] = len(*env)
	*env = append(*env, portainer.Pair{Name: name, Value: value})
}
