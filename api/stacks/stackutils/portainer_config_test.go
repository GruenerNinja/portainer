package stackutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPortainerStackConfig(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	writePortainerStackConfig(t, projectPath, `
version: 1
deploy:
  mode: flat
  targetName: tmc-proxy
  deploymentDir: ./.deployed
  cleanupDeploymentFiles: true
compose:
  files:
    - ./docker-compose.yml
    - compose.override.yml
`)

	config, found, err := LoadPortainerStackConfig(projectPath)
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, config)

	assert.Equal(t, 1, config.Version)
	assert.Equal(t, "flat", config.Deploy.Mode)
	assert.Equal(t, "tmc-proxy", config.Deploy.TargetName)
	assert.Equal(t, ".deployed", config.Deploy.DeploymentDir)
	assert.True(t, config.Deploy.CleanupDeploymentFiles)
	assert.Equal(t, []string{"docker-compose.yml", "compose.override.yml"}, config.Compose.Files)
}

func TestLoadPortainerStackConfigMissingFile(t *testing.T) {
	t.Parallel()

	config, found, err := LoadPortainerStackConfig(t.TempDir())

	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, config)
}

func TestLoadPortainerStackConfigForFileLoadsConfigNextToComposeFile(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	writePortainerStackConfig(t, filepath.Join(projectPath, "tmc-proxy"), `
version: 1
deploy:
  mode: flat
  targetName: tmc-proxy
compose:
  files:
    - docker-compose.yml
    - compose.override.yml
`)

	config, found, err := LoadPortainerStackConfigForFile(projectPath, "tmc-proxy/docker-compose.yml")
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, config)

	assert.Equal(t, "tmc-proxy", config.ConfigDir)
	assert.Equal(t, []string{"tmc-proxy/docker-compose.yml", "tmc-proxy/compose.override.yml"}, config.Compose.Files)
}

func TestLoadPortainerStackConfigForFileHandlesLeadingSlashComposeFile(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	writePortainerStackConfig(t, filepath.Join(projectPath, "tmc-proxy"), `
version: 1
compose:
  files:
    - docker-compose.yml
`)

	config, found, err := LoadPortainerStackConfigForFile(projectPath, "/tmc-proxy/docker-compose.yml")
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, config)

	assert.Equal(t, "tmc-proxy", config.ConfigDir)
	assert.Equal(t, []string{"tmc-proxy/docker-compose.yml"}, config.Compose.Files)
}

func TestLoadPortainerStackConfigForFileRejectsMultipleConfigs(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	writePortainerStackConfig(t, projectPath, "version: 1\n")
	writePortainerStackConfig(t, filepath.Join(projectPath, "tmc-proxy"), "version: 1\n")

	_, _, err := LoadPortainerStackConfigForFile(projectPath, "tmc-proxy/docker-compose.yml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one Portainer stack config file")
}

func TestLoadPortainerStackConfigRejectsUnsafeTargetName(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	writePortainerStackConfig(t, projectPath, `
version: 1
deploy:
  mode: flat
  targetName: ../etc
`)

	_, _, err := LoadPortainerStackConfig(projectPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deploy.targetName")
}

func TestLoadPortainerStackConfigRejectsUnsafeComposeFile(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	writePortainerStackConfig(t, projectPath, `
version: 1
compose:
  files:
    - ../docker-compose.yml
`)

	_, _, err := LoadPortainerStackConfig(projectPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "compose.files")
}

func TestLoadPortainerStackConfigRejectsUnsafeDeploymentDir(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	writePortainerStackConfig(t, projectPath, `
version: 1
deploy:
  mode: flat
  deploymentDir: ../.deployed
`)

	_, _, err := LoadPortainerStackConfig(projectPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deploy.deploymentDir")
}

func writePortainerStackConfig(t *testing.T, projectPath string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(projectPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, PortainerStackConfigFile), []byte(content), 0644))
}
