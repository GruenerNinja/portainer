package deployments

import (
	"os"
	"path/filepath"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/stacks/stackutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteComposeDestination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stack    *portainer.Stack
		expected string
	}{
		{
			name: "standard stack uses project path",
			stack: &portainer.Stack{
				ID:          12,
				ProjectPath: "/data/compose/12",
			},
			expected: "/data/compose/12/portainer-compose-unpacker",
		},
		{
			name: "relative path git stack uses target filesystem path",
			stack: &portainer.Stack{
				ID:                  12,
				ProjectPath:         "/data/compose/12",
				WorkflowID:          1,
				SupportRelativePath: true,
				FilesystemPath:      "/mnt/stacks",
			},
			expected: "/mnt/stacks",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, remoteComposeDestination(test.stack, ""))
		})
	}
}

func TestRemoteComposeDestinationAppendsTargetNameOnce(t *testing.T) {
	t.Parallel()

	stack := &portainer.Stack{
		ID:                  12,
		ProjectPath:         "/data/compose/12",
		WorkflowID:          1,
		SupportRelativePath: true,
		FilesystemPath:      "/data/compose",
	}

	assert.Equal(t, "/data/compose/tmc-proxy", remoteComposeDestination(stack, "tmc-proxy"))

	stack.FilesystemPath = "/data/compose/tmc-proxy"
	assert.Equal(t, "/data/compose/tmc-proxy", remoteComposeDestination(stack, "tmc-proxy"))
}

func TestApplyRelativePathStackConfigSetsSourceDirAndDestination(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	configDir := filepath.Join(projectPath, "tmc-proxy")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, stackutils.PortainerStackConfigFile), []byte(`version: 1
deploy:
  mode: flat
  targetName: tmc-proxy
  deploymentDir: .deployed
  cleanupDeploymentFiles: true
compose:
  files:
    - docker-compose.yml
`), 0644))

	stack := &portainer.Stack{
		ID:                  12,
		ProjectPath:         projectPath,
		WorkflowID:          1,
		SupportRelativePath: true,
		FilesystemPath:      "/data/compose",
		EntryPoint:          "tmc-proxy/docker-compose.yml",
	}
	opts := unpackerCmdBuilderOptions{}

	require.NoError(t, applyRelativePathStackConfig(stack, &opts))
	assert.Equal(t, "tmc-proxy", opts.sourceDir)
	assert.Equal(t, ".deployed", opts.deploymentDir)
	assert.True(t, opts.cleanupDeploymentFiles)
	assert.Equal(t, "/data/compose/tmc-proxy", opts.composeDestination)

	stack.FilesystemPath = "/data/compose/tmc-proxy"
	opts = unpackerCmdBuilderOptions{}
	require.NoError(t, applyRelativePathStackConfig(stack, &opts))
	assert.Equal(t, "/data/compose/tmc-proxy", opts.composeDestination)
}

func TestGetUnpackerImage(t *testing.T) {
	t.Setenv(composeUnpackerImageEnvVar, "")

	assert.Equal(t, defaultUnpackerImage, getUnpackerImage())

	t.Setenv(composeUnpackerImageEnvVar, "example.com/custom/compose-unpacker:test")

	assert.Equal(t, "example.com/custom/compose-unpacker:test", getUnpackerImage())
}
