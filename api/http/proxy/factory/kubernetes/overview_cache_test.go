package kubernetes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsKubernetesOverviewRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{name: "cluster pod collection", method: http.MethodGet, path: "/api/v1/pods", expected: true},
		{name: "namespaced pod collection", method: http.MethodGet, path: "/api/v1/namespaces/default/pods", expected: true},
		{name: "deployment collection", method: http.MethodGet, path: "/apis/apps/v1/namespaces/default/deployments", expected: true},
		{name: "node collection", method: http.MethodGet, path: "/kubernetes/api/v1/nodes", expected: true},
		{name: "pod detail", method: http.MethodGet, path: "/api/v1/namespaces/default/pods/web", expected: false},
		{name: "deployment detail", method: http.MethodGet, path: "/apis/apps/v1/namespaces/default/deployments/web", expected: false},
		{name: "configuration collection", method: http.MethodGet, path: "/api/v1/namespaces/default/configmaps", expected: false},
		{name: "secrets collection", method: http.MethodGet, path: "/api/v1/namespaces/default/secrets", expected: false},
		{name: "configuration mutation", method: http.MethodPut, path: "/api/v1/namespaces/default/configmaps/settings", expected: false},
		{name: "watch", method: http.MethodGet, path: "/api/v1/pods?watch=true", expected: false},
		{name: "API discovery", method: http.MethodGet, path: "/apis", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(test.method, "https://kubernetes"+test.path, nil)
			require.NoError(t, err)
			assert.Equal(t, test.expected, isKubernetesOverviewRequest(request))
		})
	}
}
