package stacks

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func TestComposeGitPayload_ValidateWithSourceID_URLNotRequired(t *testing.T) {
	t.Parallel()
	payload := &composeStackFromGitRepositoryPayload{
		Name:     "mystack",
		SourceID: portainer.SourceID(1),
		// RepositoryURL intentionally omitted
	}

	err := payload.Validate(nil)
	assert.NoError(t, err)
}

func TestComposeGitPayload_ValidateWithoutSourceID_URLRequired(t *testing.T) {
	t.Parallel()
	payload := &composeStackFromGitRepositoryPayload{
		Name: "mystack",
		// SourceID and RepositoryURL both omitted
	}

	err := payload.Validate(nil)
	assert.Error(t, err)
}

func TestComposeGitPayload_ValidateRelativePathRequiresFilesystemPath(t *testing.T) {
	t.Parallel()
	payload := &composeStackFromGitRepositoryPayload{
		Name:                "mystack",
		SourceID:            portainer.SourceID(1),
		SupportRelativePath: true,
	}

	err := payload.Validate(nil)
	assert.ErrorContains(t, err, "filesystem path")
}
