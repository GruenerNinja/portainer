package exec

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy"
	"github.com/stretchr/testify/require"
)

func TestFetchEndpointProxy(t *testing.T) {
	t.Run("local unix endpoint bypasses proxy", func(t *testing.T) {
		url, proxyServer, err := fetchEndpointProxy(&proxy.Manager{}, &portainer.Endpoint{
			URL: "unix:///var/run/docker.sock",
		})

		require.NoError(t, err)
		require.Empty(t, url)
		require.Nil(t, proxyServer)
	})

	t.Run("direct docker endpoint bypasses proxy", func(t *testing.T) {
		url, proxyServer, err := fetchEndpointProxy(&proxy.Manager{}, &portainer.Endpoint{
			Type: portainer.DockerEnvironment,
			URL:  "tcp://10.0.0.5:2376",
		})

		require.NoError(t, err)
		require.Equal(t, "tcp://10.0.0.5:2376", url)
		require.Nil(t, proxyServer)
	})
}
