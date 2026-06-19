package deployments

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
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
			expected: "/mnt/stacks/12/portainer-compose-unpacker",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, remoteComposeDestination(test.stack))
		})
	}
}
