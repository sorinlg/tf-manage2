package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// ConvertLegacyToYAML converts a legacy .tfm.conf file to the new .tfm.yaml format
func ConvertLegacyToYAML(projectDir string) error {
	legacyPath := filepath.Join(projectDir, ".tfm.conf")
	yamlPath := filepath.Join(projectDir, ".tfm.yaml")

	// Check if legacy config exists
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return fmt.Errorf("legacy config file not found at %s", legacyPath)
	}

	// Check if YAML config already exists
	if _, err := os.Stat(yamlPath); err == nil {
		return fmt.Errorf("YAML config file already exists at %s", yamlPath)
	}

	// Load legacy config
	config := DefaultConfig()
	config.ProjectDir = projectDir
	config.ConfigPath = legacyPath

	if err := parseLegacyConfigFile(legacyPath, config); err != nil {
		return fmt.Errorf("failed to parse legacy config: %w", err)
	}

	// Set version for new format
	config.ConfigVersion = "2.0"

	// Convert to YAML
	if err := WriteYAMLConfig(yamlPath, config); err != nil {
		return fmt.Errorf("failed to write YAML config: %w", err)
	}

	fmt.Printf("✅ Successfully converted configuration to YAML format\n")
	fmt.Printf("   Legacy: %s\n", legacyPath)
	fmt.Printf("   New:    %s\n", yamlPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Test the new configuration: tf --version\n")
	fmt.Printf("2. Remove the legacy file: rm %s\n", legacyPath)

	return nil
}

// WriteYAMLConfig writes a Config struct to a YAML file
func WriteYAMLConfig(configPath string, config *Config) error {
	// Create a clean config struct for YAML output (excluding runtime fields)
	yamlConfig := struct {
		ConfigVersion string `yaml:"config_version"`
		RepoName      string `yaml:"repo_name"`
		EnvRelPath    string `yaml:"env_rel_path"`
		ModuleRelPath string `yaml:"module_rel_path"`
	}{
		ConfigVersion: config.ConfigVersion,
		RepoName:      config.RepoName,
		EnvRelPath:    config.EnvRelPath,
		ModuleRelPath: config.ModuleRelPath,
	}

	data, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return err
	}

	// Add header comment
	header := `# tf-manage2 configuration file
# For documentation, see: https://github.com/sorinlg/tf-manage2

`

	return os.WriteFile(configPath, append([]byte(header), data...), 0644)
}

// ValidateConfigVersion checks if the config version is supported
func ValidateConfigVersion(version string) error {
	switch version {
	case "", "2.0":
		return nil
	default:
		return fmt.Errorf("unsupported config version: %s (supported: 2.0)", version)
	}
}
