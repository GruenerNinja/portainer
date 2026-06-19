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
		"-k",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks/1/portainer-compose-unpacker",
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

func TestBuildUndeployCmdKeepFiles(t *testing.T) {
	t.Parallel()

	cmd := buildUndeployCmd(testRemoteStack(), testKeepFilesOptions(), nil, nil)

	assert.Equal(t, []string{
		"undeploy",
		"-k",
		"https://github.com/example/stack",
		"my-stack",
		"/mnt/stacks/1/portainer-compose-unpacker",
		"docker-compose.yml",
	}, cmd)
}

func TestBuildSwarmDeployCmdKeepFiles(t *testing.T) {
	t.Parallel()

	cmd := buildSwarmDeployCmd(testRemoteStack(), testKeepFilesOptions(), nil, nil)

	assert.Equal(t, []string{
		"swarm-deploy",
		"-k",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks/1/portainer-compose-unpacker",
		"docker-compose.yml",
	}, cmd)
}

func TestBuildSwarmUndeployCmdKeepFiles(t *testing.T) {
	t.Parallel()

	cmd := buildSwarmUndeployCmd(testRemoteStack(), testKeepFilesOptions(), nil, nil)

	assert.Equal(t, []string{
		"swarm-undeploy",
		"-k",
		"my-stack",
		"/mnt/stacks/1/portainer-compose-unpacker",
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
		composeDestination: "/mnt/stacks/1/portainer-compose-unpacker",
		gitConfig: &gittypes.RepoConfig{
			URL:           "https://github.com/example/stack",
			ReferenceName: "refs/heads/main",
		},
	}
}
