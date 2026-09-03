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

type VaultTokenInfo struct {
	TTL            int64 `json:"ttl"`
	CreationTTL    int64 `json:"creation_ttl"`
	Period         int64 `json:"period"`
	ExplicitMaxTTL int64 `json:"explicit_max_ttl"`
	Renewable      bool  `json:"renewable"`
}

type VaultTokenRenewalResult struct {
	Renewed bool
	TTL     int64
	Period  int64
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

	_, err := tryVaultAddresses(config, func(address string) (struct{}, error) {
		return struct{}{}, testVaultConnectionAtAddress(ctx, config, address)
	})
	return err
}

func testVaultConnectionAtAddress(ctx context.Context, config *portainer.VaultConfig, address string) error {
	endpoint, err := vaultURL(address, "v1/sys/health")
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

	return testVaultToken(ctx, config, address)
}

func testVaultToken(ctx context.Context, config *portainer.VaultConfig, address string) error {
	endpoint, err := vaultURL(address, "v1/auth/token/lookup-self")
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

func LookupVaultToken(ctx context.Context, config *portainer.VaultConfig) (VaultTokenInfo, error) {
	if err := validateVaultTokenConfig(config); err != nil {
		return VaultTokenInfo{}, err
	}

	return tryVaultAddresses(config, func(address string) (VaultTokenInfo, error) {
		endpoint, err := vaultURL(address, "v1/auth/token/lookup-self")
		if err != nil {
			return VaultTokenInfo{}, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return VaultTokenInfo{}, err
		}
		applyVaultHeaders(req, config)

		resp, err := NewVaultClient(config.TLSSkipVerify).httpClient.Do(req)
		if err != nil {
			return VaultTokenInfo{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return VaultTokenInfo{}, fmt.Errorf("vault token lookup failed with status %d", resp.StatusCode)
		}

		var payload struct {
			Data VaultTokenInfo `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return VaultTokenInfo{}, fmt.Errorf("failed to decode Vault token lookup response: %w", err)
		}

		return payload.Data, nil
	})
}

func RenewVaultToken(ctx context.Context, config *portainer.VaultConfig) (VaultTokenInfo, error) {
	if err := validateVaultTokenConfig(config); err != nil {
		return VaultTokenInfo{}, err
	}

	return tryVaultAddresses(config, func(address string) (VaultTokenInfo, error) {
		endpoint, err := vaultURL(address, "v1/auth/token/renew-self")
		if err != nil {
			return VaultTokenInfo{}, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, http.NoBody)
		if err != nil {
			return VaultTokenInfo{}, err
		}
		applyVaultHeaders(req, config)

		resp, err := NewVaultClient(config.TLSSkipVerify).httpClient.Do(req)
		if err != nil {
			return VaultTokenInfo{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return VaultTokenInfo{}, fmt.Errorf("vault token renewal failed with status %d", resp.StatusCode)
		}

		var payload struct {
			Auth struct {
				LeaseDuration int64 `json:"lease_duration"`
				Renewable     bool  `json:"renewable"`
			} `json:"auth"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return VaultTokenInfo{}, fmt.Errorf("failed to decode Vault token renewal response: %w", err)
		}

		return VaultTokenInfo{
			TTL:       payload.Auth.LeaseDuration,
			Renewable: payload.Auth.Renewable,
		}, nil
	})
}

func RenewVaultTokenIfNeeded(ctx context.Context, config *portainer.VaultConfig) (VaultTokenRenewalResult, error) {
	info, err := LookupVaultToken(ctx, config)
	if err != nil {
		return VaultTokenRenewalResult{}, err
	}

	result := VaultTokenRenewalResult{TTL: info.TTL, Period: info.Period}
	if !info.Renewable || info.Period <= 0 {
		return result, nil
	}

	renewalThreshold := (info.Period + 1) / 2
	if info.TTL > renewalThreshold {
		return result, nil
	}

	renewedInfo, err := RenewVaultToken(ctx, config)
	if err != nil {
		return VaultTokenRenewalResult{}, err
	}

	result.Renewed = true
	result.TTL = renewedInfo.TTL
	return result, nil
}

func validateVaultTokenConfig(config *portainer.VaultConfig) error {
	if config == nil {
		return fmt.Errorf("vault configuration is required")
	}
	if config.Authentication.Method != "token" {
		return fmt.Errorf("unsupported vault authentication method %q", config.Authentication.Method)
	}
	if strings.TrimSpace(config.Authentication.Token) == "" {
		return fmt.Errorf("vault token is required")
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
	return tryVaultAddresses(config, func(address string) (map[string]string, error) {
		return readVaultSecretValuesAtAddress(ctx, config, address, secretPath)
	})
}

func readVaultSecretValuesAtAddress(ctx context.Context, config *portainer.VaultConfig, address, secretPath string) (map[string]string, error) {
	apiPath := vaultSecretAPIPath(config.KVVersion, secretPath)
	endpoint, err := vaultURL(address, apiPath)
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
	return tryVaultAddresses(config, func(address string) ([]string, error) {
		return listVaultSecretKeysAtAddress(ctx, config, address, secretPath)
	})
}

func listVaultSecretKeysAtAddress(ctx context.Context, config *portainer.VaultConfig, address, secretPath string) ([]string, error) {
	apiPath := vaultSecretListAPIPath(config.KVVersion, secretPath)
	endpoint, err := vaultURL(address, apiPath)
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
	if err == nil {
		return false
	}

	if statusErr, ok := err.(*vaultStatusError); ok {
		return statusErr.statusCode == statusCode
	}

	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, nestedErr := range joined.Unwrap() {
			if isVaultStatusError(nestedErr, statusCode) {
				return true
			}
		}
		return false
	}

	return isVaultStatusError(errors.Unwrap(err), statusCode)
}

func tryVaultAddresses[T any](config *portainer.VaultConfig, operation func(address string) (T, error)) (T, error) {
	var zero T
	addresses := vaultAddressCandidates(config)
	if len(addresses) == 0 {
		return zero, fmt.Errorf("vault address is required")
	}

	failures := make([]error, 0, len(addresses))
	for _, address := range addresses {
		result, err := operation(address)
		if err == nil {
			return result, nil
		}
		if len(addresses) == 1 {
			return zero, err
		}
		failures = append(failures, fmt.Errorf("%s: %w", address, err))
	}

	return zero, fmt.Errorf("all configured Vault addresses failed: %w", errors.Join(failures...))
}

func vaultAddressCandidates(config *portainer.VaultConfig) []string {
	if config == nil {
		return nil
	}

	addresses := make([]string, 0, 2)
	for _, address := range []string{config.InternalAddress, config.Address} {
		address = strings.TrimSpace(address)
		if address == "" || (len(addresses) > 0 && addresses[0] == address) {
			continue
		}
		addresses = append(addresses, address)
	}

	return addresses
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
