package deployments

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildDeployCmdKeepFiles(t *testing.T) {
	t.Parallel()

	cmd := buildDeployCmd(testRemoteStack(), testKeepFilesOptions(), nil, nil)

	assert.Equal(t, []string{
		"deploy",
		"--flat",
		"-k",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks",
		"docker-compose.yml",
	}, cmd)
}

func TestBuildDeployCmdDoesNotKeepFilesByDefault(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.keepFiles = false
	cmd := buildDeployCmd(testRemoteStack(), opts, nil, nil)

	assert.NotContains(t, cmd, "-k")
}

func TestBuildDeployCmdDoesNotUseFlatByDefault(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.flat = false
	cmd := buildDeployCmd(testRemoteStack(), opts, nil, nil)

	assert.NotContains(t, cmd, "--flat")
}

func TestBuildUndeployCmdKeepFiles(t *testing.T) {
	t.Parallel()

	cmd := buildUndeployCmd(testRemoteStack(), testKeepFilesOptions(), nil, nil)

	assert.Equal(t, []string{
		"undeploy",
		"--flat",
		"-k",
		"https://github.com/example/stack",
		"my-stack",
		"/mnt/stacks",
		"docker-compose.yml",
	}, cmd)
}

func TestBuildSwarmDeployCmdKeepFiles(t *testing.T) {
	t.Parallel()

	cmd := buildSwarmDeployCmd(testRemoteStack(), testKeepFilesOptions(), nil, nil)

	assert.Equal(t, []string{
		"swarm-deploy",
		"--flat",
		"-k",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks",
		"docker-compose.yml",
	}, cmd)
}

func TestBuildSwarmUndeployCmdKeepFiles(t *testing.T) {
	t.Parallel()

	cmd := buildSwarmUndeployCmd(testRemoteStack(), testKeepFilesOptions(), nil, nil)

	assert.Equal(t, []string{
		"swarm-undeploy",
		"--flat",
		"-k",
		"my-stack",
		"/mnt/stacks",
	}, cmd)
}

func testRemoteStack() *portainer.Stack {
	return &portainer.Stack{
		Name:       "my-stack",
		EntryPoint: "docker-compose.yml",
	}
}

func testKeepFilesOptions() unpackerCmdBuilderOptions {
	return unpackerCmdBuilderOptions{
		keepFiles:          true,
		flat:               true,
		composeDestination: "/mnt/stacks",
		gitConfig: &gittypes.RepoConfig{
			URL:           "https://github.com/example/stack",
			ReferenceName: "refs/heads/main",
		},
	}
}
