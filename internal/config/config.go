package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// Config represents the tf-manage configuration
type Config struct {
	RepoName      string `json:"repo_name"      yaml:"repo_name"`
	EnvRelPath    string `json:"env_rel_path"   yaml:"env_rel_path"`
	ModuleRelPath string `json:"module_rel_path" yaml:"module_rel_path"`
	ProjectDir    string `json:"project_dir"    yaml:"-"`
	ConfigPath    string `json:"config_path"    yaml:"-"`

	// Version tracking for migration and compatibility
	ConfigVersion string `json:"config_version" yaml:"config_version,omitempty"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		EnvRelPath:    "terraform/environments",
		ModuleRelPath: "terraform/modules",
	}
}

// LoadConfig loads the tf-manage configuration from either .tfm.yaml or .tfm.conf file
func LoadConfig() (*Config, error) {
	projectDir, err := findProjectDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find project directory: %w", err)
	}

	config := DefaultConfig()
	config.ProjectDir = projectDir

	// Try YAML format first (new format)
	yamlConfigPath := filepath.Join(projectDir, ".tfm.yaml")
	if _, err := os.Stat(yamlConfigPath); err == nil {
		config.ConfigPath = yamlConfigPath
		if err := parseYAMLConfigFile(yamlConfigPath, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config file %s: %w", yamlConfigPath, err)
		}
	} else {
		// Fall back to legacy format
		legacyConfigPath := filepath.Join(projectDir, ".tfm.conf")
		config.ConfigPath = legacyConfigPath

		// Check if legacy config file exists
		if _, err := os.Stat(legacyConfigPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found. Create either:\n%s\n\nOR (recommended new format):\n%s",
				generateLegacyConfigSnippet(projectDir), generateYAMLConfigSnippet(projectDir))
		}

		// Parse the legacy config file and show deprecation notice
		if err := parseLegacyConfigFile(legacyConfigPath, config); err != nil {
			return nil, fmt.Errorf("failed to parse legacy config file %s: %w", legacyConfigPath, err)
		}

		// Show deprecation notice for legacy format
		showDeprecationNotice()
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.RepoName == "" {
		return fmt.Errorf("repo_name is required")
	}
	if c.EnvRelPath == "" {
		return fmt.Errorf("env_rel_path is required")
	}
	if c.ModuleRelPath == "" {
		return fmt.Errorf("module_rel_path is required")
	}
	return nil
}

// GetModulePath returns the absolute path to the modules directory
func (c *Config) GetModulePath() string {
	return filepath.Join(c.ProjectDir, c.ModuleRelPath)
}

// GetEnvPath returns the absolute path to the environments directory
func (c *Config) GetEnvPath() string {
	return filepath.Join(c.ProjectDir, c.EnvRelPath)
}

// findProjectDir finds the git repository root directory
func findProjectDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for .git directory
	dir := cwd
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not in a git repository")
}

// parseConfigFile parses the .tfm.conf file which uses bash export syntax
// Deprecated: Use parseYAMLConfigFile for new configurations
func parseConfigFile(configPath string, config *Config) error {
	return parseLegacyConfigFile(configPath, config)
}

// parseLegacyConfigFile parses the .tfm.conf file which uses bash export syntax
func parseLegacyConfigFile(configPath string, config *Config) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Skip shebang
		if strings.HasPrefix(line, "#!") {
			continue
		}

		// Parse export statements
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
		}

		// Split on first '=' to handle values with '=' in them
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes from value
		value = strings.Trim(value, `'"`)

		// Map bash variable names to config fields
		switch key {
		case "__tfm_repo_name":
			config.RepoName = value
		case "__tfm_env_rel_path":
			config.EnvRelPath = value
		case "__tfm_module_rel_path":
			config.ModuleRelPath = value
		}
	}

	return scanner.Err()
}

// parseYAMLConfigFile parses the .tfm.yaml file
func parseYAMLConfigFile(configPath string, config *Config) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	// Set version if not specified
	if config.ConfigVersion == "" {
		config.ConfigVersion = "2.0"
	}

	// Validate config version
	if err := ValidateConfigVersion(config.ConfigVersion); err != nil {
		return err
	}

	return nil
}

// showDeprecationNotice displays a deprecation warning for legacy .tfm.conf format
func showDeprecationNotice() {
	fmt.Fprintf(os.Stderr, "\n⚠️  DEPRECATION NOTICE: Legacy .tfm.conf format detected\n")
	fmt.Fprintf(os.Stderr, "   The bash export format (.tfm.conf) is deprecated and will be removed in v2.0\n")
	fmt.Fprintf(os.Stderr, "   Please migrate to the new YAML format (.tfm.yaml)\n")
	fmt.Fprintf(os.Stderr, "   Run 'tf config convert' to automatically migrate your configuration\n\n")
}

// generateLegacyConfigSnippet generates a sample .tfm.conf file content
func generateLegacyConfigSnippet(projectDir string) string {
	projectName := filepath.Base(projectDir)
	return fmt.Sprintf(`cat > %s/.tfm.conf <<-EOF
#!/bin/bash
export __tfm_repo_name='%s'
export __tfm_env_rel_path='terraform/environments'
export __tfm_module_rel_path='terraform/modules'
EOF`, projectDir, projectName)
}

// generateYAMLConfigSnippet generates a sample .tfm.yaml file content
func generateYAMLConfigSnippet(projectDir string) string {
	projectName := filepath.Base(projectDir)
	return fmt.Sprintf(`cat > %s/.tfm.yaml <<-EOF
# tf-manage2 configuration file
# For documentation, see: https://github.com/sorinlg/tf-manage2

config_version: "2.0"
repo_name: "%s"
env_rel_path: "terraform/environments"
module_rel_path: "terraform/modules"
EOF`, projectDir, projectName)
}
