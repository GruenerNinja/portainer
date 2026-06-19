package stackutils

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func Test_GetStackFilePaths(t *testing.T) {
	t.Parallel()
	stack := &portainer.Stack{
		ProjectPath: "/tmp/stack/1",
		EntryPoint:  "file-one.yml",
	}

	t.Run("stack doesn't have additional files", func(t *testing.T) {
		expected := []string{"/tmp/stack/1/file-one.yml"}
		assert.ElementsMatch(t, expected, GetStackFilePaths(stack, true))
	})

	t.Run("stack has additional files", func(t *testing.T) {
		stack.AdditionalFiles = []string{"file-two.yml", "file-three.yml"}
		expected := []string{"/tmp/stack/1/file-one.yml", "/tmp/stack/1/file-two.yml", "/tmp/stack/1/file-three.yml"}
		assert.ElementsMatch(t, expected, GetStackFilePaths(stack, true))
	})
}

func TestIsRelativePathStack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stack    *portainer.Stack
		expected bool
	}{
		{
			name:     "nil stack",
			expected: false,
		},
		{
			name: "non git stack",
			stack: &portainer.Stack{
				SupportRelativePath: true,
				FilesystemPath:      "/mnt",
			},
			expected: false,
		},
		{
			name: "relative path disabled",
			stack: &portainer.Stack{
				WorkflowID:     1,
				FilesystemPath: "/mnt",
			},
			expected: false,
		},
		{
			name: "filesystem path missing",
			stack: &portainer.Stack{
				WorkflowID:          1,
				SupportRelativePath: true,
			},
			expected: false,
		},
		{
			name: "git stack with relative path enabled",
			stack: &portainer.Stack{
				WorkflowID:          1,
				SupportRelativePath: true,
				FilesystemPath:      "/mnt",
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, IsRelativePathStack(test.stack))
		})
	}
}
