package authorization

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/require"
)

func TestFilterAuthorizedStacksMarksViewerStacksReadOnly(t *testing.T) {
	t.Parallel()

	stacks := []portainer.Stack{
		{ID: 1, ResourceControl: &portainer.ResourceControl{UserAccesses: []portainer.UserResourceAccess{{UserID: 1, AccessLevel: portainer.ReadWriteAccessLevel}}}},
		{ID: 2, ResourceControl: &portainer.ResourceControl{TeamAccesses: []portainer.TeamResourceAccess{{TeamID: 2, AccessLevel: portainer.ReadOnlyAccessLevel}}}},
		{ID: 3, ResourceControl: &portainer.ResourceControl{UserAccesses: []portainer.UserResourceAccess{{UserID: 9, AccessLevel: portainer.ReadWriteAccessLevel}}}},
	}

	filtered := FilterAuthorizedStacks(stacks, 1, []portainer.TeamID{2})
	require.Len(t, filtered, 2)
	require.Equal(t, portainer.StackID(1), filtered[0].ID)
	require.False(t, filtered[0].ReadOnly)
	require.Equal(t, portainer.StackID(2), filtered[1].ID)
	require.True(t, filtered[1].ReadOnly)
}
