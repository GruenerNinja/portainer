package stacks

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/require"
)

func TestSanitizeReadOnlyStackRemovesMutationTokens(t *testing.T) {
	t.Parallel()

	autoUpdate := &portainer.AutoUpdateSettings{
		Webhook:  "secret-webhook-id",
		JobID:    "internal-job-id",
		Interval: "5m",
	}
	stack := &portainer.Stack{
		GitConfig:  &gittypes.RepoConfig{URL: "https://example.invalid/repo.git"},
		AutoUpdate: autoUpdate,
	}

	sanitizeReadOnlyStack(stack)

	require.Nil(t, stack.GitConfig)
	require.Empty(t, stack.AutoUpdate.Webhook)
	require.Empty(t, stack.AutoUpdate.JobID)
	require.Equal(t, "5m", stack.AutoUpdate.Interval)
	require.Equal(t, "secret-webhook-id", autoUpdate.Webhook, "sanitizing a response must not mutate shared datastore state")
}
