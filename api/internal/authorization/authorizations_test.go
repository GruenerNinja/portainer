package authorization

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func TestDefaultEndpointAuthorizationsForOperatorRole(t *testing.T) {
	t.Parallel()

	authorizations := DefaultEndpointAuthorizationsForOperatorRole(false)

	assert.True(t, authorizations[portainer.OperationDockerContainerStart])
	assert.True(t, authorizations[portainer.OperationDockerContainerStop])
	assert.True(t, authorizations[portainer.OperationDockerContainerRestart])
	assert.True(t, authorizations[portainer.OperationDockerContainerExec])
	assert.False(t, authorizations[portainer.OperationDockerContainerCreate])
	assert.False(t, authorizations[portainer.OperationDockerContainerUpdate])
	assert.False(t, authorizations[portainer.OperationDockerContainerDelete])
	assert.False(t, authorizations[portainer.OperationDockerServiceCreate])
	assert.False(t, authorizations[portainer.OperationDockerServiceUpdate])
	assert.False(t, authorizations[portainer.OperationDockerServiceDelete])
}

func TestDefaultEndpointAuthorizationsForOperatorRoleWithVolumeBrowsing(t *testing.T) {
	t.Parallel()

	authorizations := DefaultEndpointAuthorizationsForOperatorRole(true)

	assert.True(t, authorizations[portainer.OperationDockerAgentBrowseGet])
	assert.True(t, authorizations[portainer.OperationDockerAgentBrowseList])
	assert.False(t, authorizations[portainer.OperationDockerAgentBrowseDelete])
	assert.False(t, authorizations[portainer.OperationDockerAgentBrowsePut])
	assert.False(t, authorizations[portainer.OperationDockerAgentBrowseRename])
}
