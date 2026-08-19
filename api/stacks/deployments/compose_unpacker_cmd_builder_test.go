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

func TestBuildDeployCmdSourceDir(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.sourceDir = "tmc-proxy"
	stack := testRemoteStack()
	stack.EntryPoint = "tmc-proxy/docker-compose.yml"

	cmd := buildDeployCmd(stack, opts, nil, nil)

	assert.Equal(t, []string{
		"deploy",
		"--flat",
		"-k",
		"--source-dir",
		"tmc-proxy",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks",
		"tmc-proxy/docker-compose.yml",
	}, cmd)
}

func TestBuildDeployCmdDeploymentFileOptions(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.sourceDir = "tmc-proxy"
	opts.deploymentDir = ".deployed"
	opts.cleanupDeploymentFiles = true
	stack := testRemoteStack()
	stack.EntryPoint = "tmc-proxy/docker-compose.yml"

	cmd := buildDeployCmd(stack, opts, nil, nil)

	assert.Equal(t, []string{
		"deploy",
		"--flat",
		"-k",
		"--source-dir",
		"tmc-proxy",
		"--deployment-dir",
		".deployed",
		"--cleanup-deployment-files",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks",
		"tmc-proxy/docker-compose.yml",
	}, cmd)
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

func TestBuildComposeStartCmdSourceDir(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.sourceDir = "tmc-proxy"
	stack := testRemoteStack()
	stack.EntryPoint = "tmc-proxy/docker-compose.yml"

	cmd := buildComposeStartCmd(stack, opts, nil, nil)

	assert.Equal(t, []string{
		"deploy",
		"--flat",
		"-k",
		"--source-dir",
		"tmc-proxy",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks",
		"tmc-proxy/docker-compose.yml",
	}, cmd)
}

func TestBuildComposeStartCmdDeploymentFileOptions(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.deploymentDir = ".deployed"
	opts.cleanupDeploymentFiles = true

	cmd := buildComposeStartCmd(testRemoteStack(), opts, nil, nil)

	assert.Contains(t, cmd, "--deployment-dir")
	assert.Contains(t, cmd, ".deployed")
	assert.Contains(t, cmd, "--cleanup-deployment-files")
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

func TestBuildSwarmDeployCmdSourceDir(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.sourceDir = "tmc-proxy"
	stack := testRemoteStack()
	stack.EntryPoint = "tmc-proxy/docker-compose.yml"

	cmd := buildSwarmDeployCmd(stack, opts, nil, nil)

	assert.Equal(t, []string{
		"swarm-deploy",
		"--flat",
		"-k",
		"--source-dir",
		"tmc-proxy",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks",
		"tmc-proxy/docker-compose.yml",
	}, cmd)
}

func TestBuildSwarmDeployCmdPreservesRelativePathsWithRepullAndForceRedeploy(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.sourceDir = "tmc-proxy"
	opts.deploymentDir = ".deployed"
	opts.pullImage = true
	opts.forceRecreate = true
	opts.prune = true
	stack := testRemoteStack()
	stack.EntryPoint = "tmc-proxy/docker-compose.yml"

	cmd := buildSwarmDeployCmd(stack, opts, nil, nil)
	assert.Contains(t, cmd, "--source-dir")
	assert.Contains(t, cmd, "tmc-proxy")
	assert.Contains(t, cmd, "--deployment-dir")
	assert.Contains(t, cmd, ".deployed")
	assert.Contains(t, cmd, "-f")
	assert.Contains(t, cmd, "--force-recreate")
	assert.Contains(t, cmd, "-r")
}

func TestBuildSwarmDeployCmdDeploymentFileOptions(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.deploymentDir = ".deployed"
	opts.cleanupDeploymentFiles = true

	cmd := buildSwarmDeployCmd(testRemoteStack(), opts, nil, nil)

	assert.Contains(t, cmd, "--deployment-dir")
	assert.Contains(t, cmd, ".deployed")
	assert.Contains(t, cmd, "--cleanup-deployment-files")
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

func TestBuildSwarmStartCmdSourceDir(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.sourceDir = "tmc-proxy"
	stack := testRemoteStack()
	stack.EntryPoint = "tmc-proxy/docker-compose.yml"

	cmd := buildSwarmStartCmd(stack, opts, nil, nil)

	assert.Equal(t, []string{
		"swarm-deploy",
		"-f",
		"-r",
		"--flat",
		"-k",
		"--source-dir",
		"tmc-proxy",
		"https://github.com/example/stack",
		"refs/heads/main",
		"my-stack",
		"/mnt/stacks",
		"tmc-proxy/docker-compose.yml",
	}, cmd)
}

func TestBuildSwarmStartCmdDeploymentFileOptions(t *testing.T) {
	t.Parallel()

	opts := testKeepFilesOptions()
	opts.deploymentDir = ".deployed"
	opts.cleanupDeploymentFiles = true

	cmd := buildSwarmStartCmd(testRemoteStack(), opts, nil, nil)

	assert.Contains(t, cmd, "--deployment-dir")
	assert.Contains(t, cmd, ".deployed")
	assert.Contains(t, cmd, "--cleanup-deployment-files")
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
