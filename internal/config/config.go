package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the tf-manage configuration
type Config struct {
	RepoName      string `json:"repo_name"`
	EnvRelPath    string `json:"env_rel_path"`
	ModuleRelPath string `json:"module_rel_path"`
	ProjectDir    string `json:"project_dir"`
	ConfigPath    string `json:"config_path"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		EnvRelPath:    "terraform/environments",
		ModuleRelPath: "terraform/modules",
	}
}

// LoadConfig loads the tf-manage configuration from .tfm.conf file
func LoadConfig() (*Config, error) {
	projectDir, err := findProjectDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find project directory: %w", err)
	}

	configPath := filepath.Join(projectDir, ".tfm.conf")
	config := DefaultConfig()
	config.ProjectDir = projectDir
	config.ConfigPath = configPath

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s. Create it with:\n%s",
			configPath, generateConfigSnippet(projectDir))
	}

	// Parse the config file
	if err := parseConfigFile(configPath, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
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
func parseConfigFile(configPath string, config *Config) error {
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

// generateConfigSnippet generates a sample .tfm.conf file content
func generateConfigSnippet(projectDir string) string {
	projectName := filepath.Base(projectDir)
	return fmt.Sprintf(`cat > %s/.tfm.conf <<-EOF
#!/bin/bash
export __tfm_repo_name='%s'
export __tfm_env_rel_path='terraform/environments'
export __tfm_module_rel_path='terraform/modules'
EOF`, projectDir, projectName)
}
