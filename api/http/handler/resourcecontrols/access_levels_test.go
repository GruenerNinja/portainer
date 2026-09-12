package resourcecontrols

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/require"
)

func TestValidateDistinctResourceAccesses(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateDistinctResourceAccesses([]int{1}, []int{2}, []int{3}, []int{4}))
	require.Error(t, validateDistinctResourceAccesses([]int{1}, []int{1}, nil, nil))
	require.Error(t, validateDistinctResourceAccesses(nil, nil, []int{3}, []int{3}))
}

func TestReadOnlyAccessIDs(t *testing.T) {
	t.Parallel()

	rc := &portainer.ResourceControl{
		UserAccesses: []portainer.UserResourceAccess{
			{UserID: 5, AccessLevel: portainer.ReadOnlyAccessLevel},
			{UserID: 1, AccessLevel: portainer.ReadWriteAccessLevel},
			{UserID: 2, AccessLevel: portainer.ReadOnlyAccessLevel},
		},
		TeamAccesses: []portainer.TeamResourceAccess{
			{TeamID: 8, AccessLevel: portainer.ReadOnlyAccessLevel},
			{TeamID: 3, AccessLevel: portainer.ReadWriteAccessLevel},
		},
	}

	users, teams := readOnlyAccessIDs(rc)
	require.Equal(t, []int{2, 5}, users)
	require.Equal(t, []int{8}, teams)
}
