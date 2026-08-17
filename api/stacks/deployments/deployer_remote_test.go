package deployments

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
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

	found, err := applyRelativePathStackConfig(stack, &opts)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "tmc-proxy", opts.sourceDir)
	assert.Equal(t, ".deployed", opts.deploymentDir)
	assert.True(t, opts.cleanupDeploymentFiles)
	assert.Equal(t, "/data/compose/tmc-proxy", opts.composeDestination)

	stack.FilesystemPath = "/data/compose/tmc-proxy"
	opts = unpackerCmdBuilderOptions{}
	found, err = applyRelativePathStackConfig(stack, &opts)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "/data/compose/tmc-proxy", opts.composeDestination)
}

func TestApplyRelativePathStackConfigReportsMissingConfig(t *testing.T) {
	t.Parallel()

	stack := &portainer.Stack{
		ID:                  12,
		ProjectPath:         t.TempDir(),
		WorkflowID:          1,
		SupportRelativePath: true,
		FilesystemPath:      "/data/compose",
		EntryPoint:          "docker-compose.yml",
	}
	opts := unpackerCmdBuilderOptions{}

	found, err := applyRelativePathStackConfig(stack, &opts)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestRequiresPortainerStackConfig(t *testing.T) {
	t.Parallel()

	assert.True(t, requiresPortainerStackConfig(OperationDeploy))
	assert.True(t, requiresPortainerStackConfig(OperationComposeStart))
	assert.True(t, requiresPortainerStackConfig(OperationSwarmDeploy))
	assert.True(t, requiresPortainerStackConfig(OperationSwarmStart))
	assert.False(t, requiresPortainerStackConfig(OperationUndeploy))
	assert.False(t, requiresPortainerStackConfig(OperationComposeStop))
	assert.False(t, requiresPortainerStackConfig(OperationSwarmUndeploy))
	assert.False(t, requiresPortainerStackConfig(OperationSwarmStop))
}

func TestGetUnpackerImage(t *testing.T) {
	t.Setenv(composeUnpackerImageEnvVar, "")

	assert.Equal(t, defaultUnpackerImage, getUnpackerImage())

	t.Setenv(composeUnpackerImageEnvVar, "example.com/custom/compose-unpacker:test")

	assert.Equal(t, "example.com/custom/compose-unpacker:test", getUnpackerImage())
}

func Test_unpackerImageRegistryAuth(t *testing.T) {
	t.Parallel()
	_, store := datastore.MustNewTestStore(t, false, true)

	d := NewStackDeployer(nil, nil, nil, nil, store)

	err := store.Registry().Create(&portainer.Registry{
		ID:             1,
		Type:           portainer.CustomRegistry,
		URL:            "myregistry.example.com",
		Authentication: true,
		Username:       "someuser",
		Password:       "somepass",
	})
	require.NoError(t, err, "failed to create a test registry")

	err = store.Registry().Create(&portainer.Registry{
		ID:   2,
		Type: portainer.CustomRegistry,
		URL:  "noauth.example.com",
	})
	require.NoError(t, err, "failed to create a test registry")

	// an unparsable image name is rejected before any registry lookup happens
	auth, err := d.unpackerImageRegistryAuth("INVALID IMAGE NAME")
	require.Error(t, err)
	require.Empty(t, auth)

	// no registry matches this image, so callers must fall back to an unauthenticated pull
	auth, err = d.unpackerImageRegistryAuth("unmatched.example.com/portainer/compose-unpacker:2.0.0")
	require.Error(t, err)
	require.Empty(t, auth)

	// the matching registry has authentication disabled, so callers must fall back to an unauthenticated pull
	auth, err = d.unpackerImageRegistryAuth("noauth.example.com/portainer/compose-unpacker:2.0.0")
	require.Error(t, err)
	require.Empty(t, auth)

	// the matching registry has authentication enabled, so callers get an encoded basic-auth header
	auth, err = d.unpackerImageRegistryAuth("myregistry.example.com/portainer/compose-unpacker:2.0.0")
	require.NoError(t, err)
	require.NotEmpty(t, auth)

	decoded, err := base64.StdEncoding.DecodeString(auth)
	require.NoError(t, err)
	require.Contains(t, string(decoded), "someuser")
	require.Contains(t, string(decoded), "somepass")
}

func Test_resolveUnpackerRegistryAuth(t *testing.T) {
	t.Parallel()
	_, store := datastore.MustNewTestStore(t, false, true)

	d := NewStackDeployer(nil, nil, nil, nil, store)

	// no registry matches, so resolveUnpackerRegistryAuth swallows the error and returns an empty auth
	require.Empty(t, d.resolveUnpackerRegistryAuth("unmatched.example.com/portainer/compose-unpacker:2.0.0"))

	err := store.Registry().Create(&portainer.Registry{
		ID:             1,
		Type:           portainer.CustomRegistry,
		URL:            "resolveauth.example.com",
		Authentication: true,
		Username:       "someuser",
		Password:       "somepass",
	})
	require.NoError(t, err, "failed to create a test registry")

	auth := d.resolveUnpackerRegistryAuth("resolveauth.example.com/portainer/compose-unpacker:2.0.0")
	require.NotEmpty(t, auth)

	decoded, err := base64.StdEncoding.DecodeString(auth)
	require.NoError(t, err)
	require.Contains(t, string(decoded), "someuser")
}
