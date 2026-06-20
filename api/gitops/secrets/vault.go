package secrets

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	portainer "github.com/portainer/portainer/api"
)

type VaultClient struct {
	httpClient *http.Client
}

func NewVaultClient(tlsSkipVerify bool) *VaultClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &VaultClient{
		httpClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		},
	}
}

func TestVaultConnection(ctx context.Context, config *portainer.VaultConfig) error {
	if config == nil {
		return fmt.Errorf("vault configuration is required")
	}

	endpoint, err := vaultURL(config.Address, "v1/sys/health")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	applyVaultHeaders(req, config)

	resp, err := NewVaultClient(config.TLSSkipVerify).httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return nil
	}

	return fmt.Errorf("vault health check failed with status %d", resp.StatusCode)
}

func ResolveVaultSecret(ctx context.Context, config *portainer.VaultConfig, secretPath, key string) (string, error) {
	if config == nil {
		return "", fmt.Errorf("vault configuration is required")
	}

	if strings.TrimSpace(secretPath) == "" {
		return "", fmt.Errorf("vault secret path is required")
	}

	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("vault secret key is required")
	}

	apiPath := vaultSecretAPIPath(config.KVVersion, secretPath)
	endpoint, err := vaultURL(config.Address, apiPath)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	applyVaultHeaders(req, config)

	resp, err := NewVaultClient(config.TLSSkipVerify).httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("vault secret read failed with status %d", resp.StatusCode)
	}

	var payload struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	values := payload.Data
	if config.KVVersion == 2 {
		if nested, ok := payload.Data["data"].(map[string]any); ok {
			values = nested
		}
	}

	value, ok := values[key]
	if !ok {
		return "", fmt.Errorf("vault secret key %q was not found", key)
	}

	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprint(v), nil
	}
}

func vaultURL(address, apiPath string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", fmt.Errorf("vault address is required")
	}

	base, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("invalid vault address: %w", err)
	}
	if base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("vault address must include scheme and host")
	}

	base.Path = path.Join(base.Path, strings.TrimLeft(apiPath, "/"))
	return base.String(), nil
}

func vaultSecretAPIPath(kvVersion int, secretPath string) string {
	secretPath = strings.Trim(strings.TrimSpace(secretPath), "/")
	if kvVersion != 2 {
		return path.Join("v1", secretPath)
	}

	parts := strings.SplitN(secretPath, "/", 2)
	if len(parts) == 1 {
		return path.Join("v1", parts[0], "data")
	}

	if parts[1] == "data" || strings.HasPrefix(parts[1], "data/") {
		return path.Join("v1", secretPath)
	}

	return path.Join("v1", parts[0], "data", parts[1])
}

func applyVaultHeaders(req *http.Request, config *portainer.VaultConfig) {
	if config.Authentication.Method == "token" && config.Authentication.Token != "" {
		req.Header.Set("X-Vault-Token", config.Authentication.Token)
	}
	if config.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", config.Namespace)
	}
}
