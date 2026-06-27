package libstack

import (
	"testing"

	"github.com/docker/cli/cli/flags"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/require"
)

func TestApplyTLSOptions(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		opts := flags.NewClientOptions()

		applyTLSOptions(opts, portainer.TLSConfiguration{})

		require.False(t, opts.TLS)
		require.False(t, opts.TLSVerify)
		require.Nil(t, opts.TLSOptions)
	})

	t.Run("enabled with verification", func(t *testing.T) {
		opts := flags.NewClientOptions()

		applyTLSOptions(opts, portainer.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: false,
			TLSCACertPath: "/tmp/ca.pem",
			TLSCertPath:   "/tmp/cert.pem",
			TLSKeyPath:    "/tmp/key.pem",
		})

		require.True(t, opts.TLS)
		require.True(t, opts.TLSVerify)
		require.NotNil(t, opts.TLSOptions)
		require.Equal(t, "/tmp/ca.pem", opts.TLSOptions.CAFile)
		require.Equal(t, "/tmp/cert.pem", opts.TLSOptions.CertFile)
		require.Equal(t, "/tmp/key.pem", opts.TLSOptions.KeyFile)
		require.False(t, opts.TLSOptions.InsecureSkipVerify)
	})

	t.Run("enabled without verification", func(t *testing.T) {
		opts := flags.NewClientOptions()

		applyTLSOptions(opts, portainer.TLSConfiguration{
			TLSSkipVerify: true,
		})

		require.True(t, opts.TLS)
		require.False(t, opts.TLSVerify)
		require.NotNil(t, opts.TLSOptions)
		require.True(t, opts.TLSOptions.InsecureSkipVerify)
	})
}
