package docker

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDockerOverviewRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{name: "containers collection", method: http.MethodGet, path: "/v1.52/containers/json", expected: true},
		{name: "images collection", method: http.MethodGet, path: "/images/json", expected: true},
		{name: "networks collection", method: http.MethodGet, path: "/networks", expected: true},
		{name: "nodes collection", method: http.MethodGet, path: "/nodes", expected: true},
		{name: "container inspect", method: http.MethodGet, path: "/containers/abc/json", expected: false},
		{name: "network inspect", method: http.MethodGet, path: "/networks/abc", expected: false},
		{name: "configuration collection", method: http.MethodGet, path: "/configs", expected: false},
		{name: "secrets collection", method: http.MethodGet, path: "/secrets", expected: false},
		{name: "container mutation", method: http.MethodPost, path: "/containers/json", expected: false},
		{name: "system configuration", method: http.MethodGet, path: "/info", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(test.method, "http://docker"+test.path, nil)
			require.NoError(t, err)
			assert.Equal(t, test.expected, isDockerOverviewRequest(request))
		})
	}
}
