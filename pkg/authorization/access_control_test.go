package authorization

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/require"
)

func TestResourceAccessLevels(t *testing.T) {
	t.Parallel()

	rc := &portainer.ResourceControl{
		UserAccesses: []portainer.UserResourceAccess{
			{UserID: 1, AccessLevel: portainer.ReadWriteAccessLevel},
			{UserID: 2, AccessLevel: portainer.ReadOnlyAccessLevel},
		},
		TeamAccesses: []portainer.TeamResourceAccess{
			{TeamID: 3, AccessLevel: portainer.ReadWriteAccessLevel},
			{TeamID: 4, AccessLevel: portainer.ReadOnlyAccessLevel},
		},
	}

	require.True(t, UserCanAccessResource(1, nil, rc))
	require.True(t, UserCanReadResource(1, nil, rc))
	require.False(t, UserCanAccessResource(2, nil, rc))
	require.True(t, UserCanReadResource(2, nil, rc))
	require.True(t, UserCanAccessResource(9, []portainer.TeamID{3}, rc))
	require.True(t, UserCanReadResource(9, []portainer.TeamID{3}, rc))
	require.False(t, UserCanAccessResource(9, []portainer.TeamID{4}, rc))
	require.True(t, UserCanReadResource(9, []portainer.TeamID{4}, rc))
}

func TestPublicResourceIsReadableAndWritable(t *testing.T) {
	t.Parallel()

	rc := &portainer.ResourceControl{Public: true}
	require.True(t, UserCanAccessResource(1, nil, rc))
	require.True(t, UserCanReadResource(1, nil, rc))
}

func TestLegacyUnsetAccessLevelRemainsReadWrite(t *testing.T) {
	t.Parallel()

	rc := &portainer.ResourceControl{
		UserAccesses: []portainer.UserResourceAccess{{UserID: 1}},
		TeamAccesses: []portainer.TeamResourceAccess{{TeamID: 2}},
	}
	require.True(t, UserCanAccessResource(1, nil, rc))
	require.True(t, UserCanAccessResource(9, []portainer.TeamID{2}, rc))
}
