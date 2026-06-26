package stacks

import (
	"testing"

	gittypes "github.com/portainer/portainer/api/git/types"

	"github.com/stretchr/testify/assert"
)

func TestRedeployGitCredentialsUsesStoredGitConfigAuthentication(t *testing.T) {
	t.Parallel()

	username, password := redeployGitCredentials(&gittypes.RepoConfig{
		Authentication: &gittypes.GitAuthentication{
			Username: "git-user",
			Password: "git-token",
		},
	}, stackGitRedeployPayload{})

	assert.Equal(t, "git-user", username)
	assert.Equal(t, "git-token", password)
}

func TestRedeployGitCredentialsPayloadOverridesStoredAuthentication(t *testing.T) {
	t.Parallel()

	username, password := redeployGitCredentials(&gittypes.RepoConfig{
		Authentication: &gittypes.GitAuthentication{
			Username: "stored-user",
			Password: "stored-token",
		},
	}, stackGitRedeployPayload{
		RepositoryAuthentication: true,
		RepositoryUsername:       "payload-user",
		RepositoryPassword:       "payload-token",
	})

	assert.Equal(t, "payload-user", username)
	assert.Equal(t, "payload-token", password)
}

func TestRedeployGitCredentialsKeepsStoredValuesWhenPayloadOmitsThem(t *testing.T) {
	t.Parallel()

	username, password := redeployGitCredentials(&gittypes.RepoConfig{
		Authentication: &gittypes.GitAuthentication{
			Username: "stored-user",
			Password: "stored-token",
		},
	}, stackGitRedeployPayload{
		RepositoryAuthentication: true,
	})

	assert.Equal(t, "stored-user", username)
	assert.Equal(t, "stored-token", password)
}

func TestRedeployGitCredentialsReturnsEmptyValuesWithoutAuthentication(t *testing.T) {
	t.Parallel()

	username, password := redeployGitCredentials(&gittypes.RepoConfig{}, stackGitRedeployPayload{})

	assert.Empty(t, username)
	assert.Empty(t, password)
}
