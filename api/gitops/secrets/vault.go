package secrets

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
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

type vaultStatusError struct {
	operation  string
	statusCode int
}

func (err *vaultStatusError) Error() string {
	return fmt.Sprintf("vault secret %s failed with status %d", err.operation, err.statusCode)
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
	if config.Authentication.Method != "token" {
		return fmt.Errorf("unsupported vault authentication method %q", config.Authentication.Method)
	}
	if strings.TrimSpace(config.Authentication.Token) == "" {
		return fmt.Errorf("vault token is required")
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

	if !isHealthyVaultStatus(resp.StatusCode) {
		return fmt.Errorf("vault health check failed with status %d", resp.StatusCode)
	}

	return testVaultToken(ctx, config)
}

func testVaultToken(ctx context.Context, config *portainer.VaultConfig) error {
	endpoint, err := vaultURL(config.Address, "v1/auth/token/lookup-self")
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

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("vault token validation failed with status %d", resp.StatusCode)
	}

	return nil
}

func isHealthyVaultStatus(statusCode int) bool {
	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return true
	}

	// Vault reports healthy standby and replication modes with non-2xx status
	// codes unless the health endpoint is configured with custom status codes.
	switch statusCode {
	case http.StatusTooManyRequests, 472, 473:
		return true
	default:
		return false
	}
}

func ResolveVaultSecret(ctx context.Context, config *portainer.VaultConfig, secretPath, key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("vault secret key is required")
	}

	values, err := ResolveVaultSecretValues(ctx, config, secretPath)
	if err != nil {
		return "", err
	}

	value, ok := values[key]
	if !ok {
		return "", fmt.Errorf("vault secret key %q was not found", key)
	}

	return value, nil
}

func ResolveVaultSecretValues(ctx context.Context, config *portainer.VaultConfig, secretPath string) (map[string]string, error) {
	if config == nil {
		return nil, fmt.Errorf("vault configuration is required")
	}

	if strings.TrimSpace(secretPath) == "" {
		return nil, fmt.Errorf("vault secret path is required")
	}

	values, err := readVaultSecretValues(ctx, config, secretPath)
	if err == nil {
		return values, nil
	}

	if !isVaultStatusError(err, http.StatusNotFound) {
		return nil, err
	}

	values, folderErr := resolveVaultSecretFolderValues(ctx, config, secretPath)
	if folderErr != nil {
		return nil, fmt.Errorf("vault secret path %q was not found as a secret and folder expansion failed; make sure the path includes the KV mount, for example kv/app: %w", normalizeVaultSecretPath(secretPath), folderErr)
	}

	return values, nil
}

func readVaultSecretValues(ctx context.Context, config *portainer.VaultConfig, secretPath string) (map[string]string, error) {
	apiPath := vaultSecretAPIPath(config.KVVersion, secretPath)
	endpoint, err := vaultURL(config.Address, apiPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	applyVaultHeaders(req, config)

	resp, err := NewVaultClient(config.TLSSkipVerify).httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &vaultStatusError{operation: "read", statusCode: resp.StatusCode}
	}

	var payload struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	values := payload.Data
	if config.KVVersion == 2 {
		if nested, ok := payload.Data["data"].(map[string]any); ok {
			values = nested
		}
	}

	resolved := make(map[string]string, len(values))
	for key, value := range values {
		switch v := value.(type) {
		case string:
			resolved[key] = v
		default:
			resolved[key] = fmt.Sprint(v)
		}
	}

	return resolved, nil
}

func resolveVaultSecretFolderValues(ctx context.Context, config *portainer.VaultConfig, secretPath string) (map[string]string, error) {
	keys, err := listVaultSecretKeys(ctx, config, secretPath)
	if err != nil {
		return nil, err
	}

	resolved := make(map[string]string)
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || strings.HasSuffix(key, "/") {
			continue
		}

		childPath := vaultSecretChildPath(secretPath, key)
		values, err := readVaultSecretValues(ctx, config, childPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read Vault secret %q: %w", childPath, err)
		}

		mergeVaultFolderSecretValues(resolved, key, values)
	}

	return resolved, nil
}

func listVaultSecretKeys(ctx context.Context, config *portainer.VaultConfig, secretPath string) ([]string, error) {
	apiPath := vaultSecretListAPIPath(config.KVVersion, secretPath)
	endpoint, err := vaultURL(config.Address, apiPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "LIST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	applyVaultHeaders(req, config)

	resp, err := NewVaultClient(config.TLSSkipVerify).httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
		_ = resp.Body.Close()
		return listVaultSecretKeysWithGET(ctx, config, endpoint)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &vaultStatusError{operation: "folder list", statusCode: resp.StatusCode}
	}

	return decodeVaultSecretKeys(resp)
}

func listVaultSecretKeysWithGET(ctx context.Context, config *portainer.VaultConfig, endpoint string) ([]string, error) {
	listEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	query := listEndpoint.Query()
	query.Set("list", "true")
	listEndpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listEndpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	applyVaultHeaders(req, config)

	resp, err := NewVaultClient(config.TLSSkipVerify).httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &vaultStatusError{operation: "folder list", statusCode: resp.StatusCode}
	}

	return decodeVaultSecretKeys(resp)
}

func decodeVaultSecretKeys(resp *http.Response) ([]string, error) {
	var payload struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload.Data.Keys, nil
}

func mergeVaultFolderSecretValues(resolved map[string]string, secretName string, values map[string]string) {
	if len(values) == 1 {
		for _, value := range values {
			resolved[secretName] = value
		}
		return
	}

	for key, value := range values {
		resolved[secretName+"_"+key] = value
	}
}

func vaultSecretChildPath(secretPath, key string) string {
	return path.Join(normalizeVaultSecretPath(secretPath), strings.Trim(strings.TrimSuffix(key, "/"), "/"))
}

func isVaultStatusError(err error, statusCode int) bool {
	var statusErr *vaultStatusError
	return errors.As(err, &statusErr) && statusErr.statusCode == statusCode
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
	secretPath = normalizeVaultSecretPath(secretPath)
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

func vaultSecretListAPIPath(kvVersion int, secretPath string) string {
	secretPath = normalizeVaultSecretPath(secretPath)
	if kvVersion != 2 {
		return path.Join("v1", secretPath)
	}

	parts := strings.SplitN(secretPath, "/", 2)
	if len(parts) == 1 {
		return path.Join("v1", parts[0], "metadata")
	}

	switch {
	case parts[1] == "metadata" || strings.HasPrefix(parts[1], "metadata/"):
		return path.Join("v1", secretPath)
	case parts[1] == "data":
		return path.Join("v1", parts[0], "metadata")
	case strings.HasPrefix(parts[1], "data/"):
		return path.Join("v1", parts[0], "metadata", strings.TrimPrefix(parts[1], "data/"))
	default:
		return path.Join("v1", parts[0], "metadata", parts[1])
	}
}

func normalizeVaultSecretPath(secretPath string) string {
	return strings.Trim(strings.TrimSpace(secretPath), "/")
}

func applyVaultHeaders(req *http.Request, config *portainer.VaultConfig) {
	if config.Authentication.Method == "token" && config.Authentication.Token != "" {
		req.Header.Set("X-Vault-Token", config.Authentication.Token)
	}
	if config.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", config.Namespace)
	}
}
