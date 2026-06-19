package stackutils

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	PortainerStackConfigFile     = ".portainer.yml"
	portainerStackConfigFileYAML = ".portainer.yaml"
	portainerStackConfigVersion  = 1
)

type PortainerStackConfig struct {
	Version int                         `yaml:"version"`
	Deploy  PortainerStackDeployConfig  `yaml:"deploy"`
	Compose PortainerStackComposeConfig `yaml:"compose"`
}

type PortainerStackDeployConfig struct {
	Mode       string `yaml:"mode"`
	TargetName string `yaml:"targetName"`
}

type PortainerStackComposeConfig struct {
	Files []string `yaml:"files"`
}

func LoadPortainerStackConfig(projectPath string) (*PortainerStackConfig, bool, error) {
	configPath, found, err := findPortainerStackConfig(projectPath)
	if err != nil || !found {
		return nil, found, err
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, true, fmt.Errorf("failed to read %s: %w", filepath.Base(configPath), err)
	}

	var config PortainerStackConfig
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return nil, true, fmt.Errorf("failed to parse %s: %w", filepath.Base(configPath), err)
	}

	if err := config.Validate(); err != nil {
		return nil, true, err
	}

	return &config, true, nil
}

func (config *PortainerStackConfig) Validate() error {
	if config.Version != portainerStackConfigVersion {
		return fmt.Errorf("%s version must be %d", PortainerStackConfigFile, portainerStackConfigVersion)
	}

	config.Deploy.Mode = strings.TrimSpace(config.Deploy.Mode)
	if config.Deploy.Mode != "" && config.Deploy.Mode != "flat" {
		return fmt.Errorf("%s deploy.mode must be flat", PortainerStackConfigFile)
	}

	config.Deploy.TargetName = strings.TrimSpace(config.Deploy.TargetName)
	if config.Deploy.TargetName != "" {
		if err := validatePortainerTargetName(config.Deploy.TargetName); err != nil {
			return err
		}
	}

	for i, file := range config.Compose.Files {
		cleaned, err := cleanPortainerComposeFilePath(file)
		if err != nil {
			return err
		}

		config.Compose.Files[i] = cleaned
	}

	return nil
}

func (config PortainerStackConfig) HasDeployConfig() bool {
	return config.Deploy.Mode != "" || config.Deploy.TargetName != ""
}

func findPortainerStackConfig(projectPath string) (string, bool, error) {
	var foundPath string
	for _, name := range []string{PortainerStackConfigFile, portainerStackConfigFileYAML} {
		configPath := filepath.Join(projectPath, name)
		if _, err := os.Stat(configPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return "", false, fmt.Errorf("failed to inspect %s: %w", name, err)
		}

		if foundPath != "" {
			return "", false, fmt.Errorf("only one Portainer stack config file is allowed")
		}

		foundPath = configPath
	}

	return foundPath, foundPath != "", nil
}

func validatePortainerTargetName(targetName string) error {
	if path.IsAbs(targetName) || filepath.VolumeName(targetName) != "" {
		return fmt.Errorf("%s deploy.targetName must be a relative folder name", PortainerStackConfigFile)
	}

	if targetName == "." || targetName == ".." || strings.ContainsAny(targetName, `/\`) {
		return fmt.Errorf("%s deploy.targetName must be a single folder name", PortainerStackConfigFile)
	}

	return nil
}

func cleanPortainerComposeFilePath(filePath string) (string, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return "", fmt.Errorf("%s compose.files entries cannot be empty", PortainerStackConfigFile)
	}

	if path.IsAbs(filePath) || filepath.VolumeName(filePath) != "" || strings.Contains(filePath, `\`) {
		return "", fmt.Errorf("%s compose.files entries must be relative repository paths", PortainerStackConfigFile)
	}

	cleaned := path.Clean(filePath)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("%s compose.files entries cannot traverse outside the repository", PortainerStackConfigFile)
	}

	return cleaned, nil
}
