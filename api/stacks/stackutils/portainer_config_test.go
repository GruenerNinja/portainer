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
	assert.Equal(t, []string{"docker-compose.yml", "compose.override.yml"}, config.Compose.Files)
}

func TestLoadPortainerStackConfigMissingFile(t *testing.T) {
	t.Parallel()

	config, found, err := LoadPortainerStackConfig(t.TempDir())

	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, config)
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

func writePortainerStackConfig(t *testing.T, projectPath string, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(projectPath, PortainerStackConfigFile), []byte(content), 0644))
}
