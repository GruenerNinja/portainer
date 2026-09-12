package auth

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"

	"github.com/stretchr/testify/require"
)

func TestResolveOAuthTeamMemberships(t *testing.T) {
	settings := &portainer.OAuthSettings{
		DefaultTeamID: 3,
		TeamMemberships: portainer.OAuthTeamMembershipSettings{
			OAuthClaimName: "groups",
			OAuthClaimMappings: []portainer.OAuthClaimMapping{
				{ClaimValRegex: `^portainer-admins$`, Team: 1},
				{ClaimValRegex: `^portainer-(agentic|automation)$`, Team: 2},
			},
		},
	}

	t.Run("maps every matching claim and does not add the default", func(t *testing.T) {
		desired, managed, err := resolveOAuthTeamMemberships(settings, map[string]any{
			"groups": []any{"portainer-admins", "portainer-agentic", 42},
		})
		require.NoError(t, err)
		require.Equal(t, map[portainer.TeamID]struct{}{1: {}, 2: {}}, desired)
		require.Equal(t, map[portainer.TeamID]struct{}{1: {}, 2: {}, 3: {}}, managed)
	})

	t.Run("uses the default team when the claim is absent", func(t *testing.T) {
		desired, managed, err := resolveOAuthTeamMemberships(settings, nil)
		require.NoError(t, err)
		require.Equal(t, map[portainer.TeamID]struct{}{3: {}}, desired)
		require.Equal(t, map[portainer.TeamID]struct{}{1: {}, 2: {}, 3: {}}, managed)
	})

	t.Run("rejects an invalid regular expression", func(t *testing.T) {
		invalidSettings := *settings
		invalidSettings.TeamMemberships.OAuthClaimMappings = []portainer.OAuthClaimMapping{
			{ClaimValRegex: `[`, Team: 1},
		}
		_, _, err := resolveOAuthTeamMemberships(&invalidSettings, nil)
		require.Error(t, err)
	})
}

func TestSyncOAuthTeamMemberships(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, false)
	handler := &Handler{DataStore: store}

	existing := []*portainer.TeamMembership{
		{UserID: 7, TeamID: 1, Role: portainer.TeamLeader},
		{UserID: 7, TeamID: 2, Role: portainer.TeamMember},
		{UserID: 7, TeamID: 9, Role: portainer.TeamMember},
	}
	for _, membership := range existing {
		require.NoError(t, store.TeamMembership().Create(membership))
	}

	settings := &portainer.OAuthSettings{
		OAuthAutoMapTeamMemberships: true,
		TeamMemberships: portainer.OAuthTeamMembershipSettings{
			OAuthClaimName: "groups",
			OAuthClaimMappings: []portainer.OAuthClaimMapping{
				{ClaimValRegex: `^admin$`, Team: 1},
				{ClaimValRegex: `^agentic$`, Team: 2},
				{ClaimValRegex: `^readonly$`, Team: 3},
			},
		},
	}

	require.NoError(t, handler.syncOAuthTeamMemberships(7, settings, map[string]any{
		"groups": []any{"admin", "readonly"},
	}))

	memberships, err := store.TeamMembership().TeamMembershipsByUserID(7)
	require.NoError(t, err)
	require.Len(t, memberships, 3)

	rolesByTeam := make(map[portainer.TeamID]portainer.MembershipRole, len(memberships))
	for _, membership := range memberships {
		rolesByTeam[membership.TeamID] = membership.Role
	}
	require.Equal(t, map[portainer.TeamID]portainer.MembershipRole{
		1: portainer.TeamLeader,
		3: portainer.TeamMember,
		9: portainer.TeamMember,
	}, rolesByTeam)
}
