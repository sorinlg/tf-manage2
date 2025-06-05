package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sorinlg/tf-manage2/internal/config"
)

// Completion provides bash completion functionality
type Completion struct {
	config *config.Config
}

// NewCompletion creates a new completion handler
func NewCompletion(cfg *config.Config) *Completion {
	return &Completion{
		config: cfg,
	}
}

// SuggestProducts lists available products from environments directory
func (c *Completion) SuggestProducts() error {
	envPath := c.config.GetEnvPath()

	entries, err := os.ReadDir(envPath)
	if err != nil {
		// If directory doesn't exist, suggest creating it
		return fmt.Errorf("environment path does not exist: %s", envPath)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no products found in: %s", envPath)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Println(entry.Name())
		}
	}
	return nil
}

// SuggestModules lists available modules from modules directory
func (c *Completion) SuggestModules() error {
	modulePath := c.config.GetModulePath()

	entries, err := os.ReadDir(modulePath)
	if err != nil {
		return fmt.Errorf("module path does not exist: %s", modulePath)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no modules found in: %s", modulePath)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Println(entry.Name())
		}
	}
	return nil
}

// SuggestEnvironments lists available environments for a given product and module
func (c *Completion) SuggestEnvironments(product, module string) error {
	// First check if the product exists
	productPath := filepath.Join(c.config.GetEnvPath(), product)
	if _, err := os.Stat(productPath); os.IsNotExist(err) {
		return fmt.Errorf("product path does not exist: %s", productPath)
	}

	// Then check if the module exists
	modulePath := filepath.Join(c.config.GetModulePath(), module)
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return fmt.Errorf("module path does not exist: %s", modulePath)
	}

	entries, err := os.ReadDir(productPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %s", productPath)
	}

	var environments []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this environment has the specified module
		moduleDirPath := filepath.Join(productPath, entry.Name(), module)
		if _, err := os.Stat(moduleDirPath); err == nil {
			environments = append(environments, entry.Name())
		}
	}

	if len(environments) == 0 {
		fmt.Fprintf(os.Stderr, "Search pattern %s/*/<%s> is empty\nYou must create entries first\n", productPath, module)
		return fmt.Errorf("no environments found for product %s and module %s", product, module)
	}

	for _, env := range environments {
		fmt.Println(env)
	}
	return nil
}

// SuggestConfigs lists available configuration files for a given product, env, and module
func (c *Completion) SuggestConfigs(product, env, module string) error {
	// First check if the product exists
	productPath := filepath.Join(c.config.GetEnvPath(), product)
	if _, err := os.Stat(productPath); os.IsNotExist(err) {
		return fmt.Errorf("product path does not exist: %s", productPath)
	}

	// Then check if the environment exists
	envPath := filepath.Join(productPath, env)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("environment path does not exist: %s", envPath)
	}

	// Then check if the module exists
	modulePath := filepath.Join(c.config.GetModulePath(), module)
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return fmt.Errorf("module path does not exist: %s", modulePath)
	}

	configPath := filepath.Join(c.config.GetEnvPath(), product, env, module)

	entries, err := os.ReadDir(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config directory: %s", configPath)
	}

	// Filter for .tfvars files and exclude .tfplan files
	var configs []string
	tfvarsRegex := regexp.MustCompile(`^(.+)\.tfvars$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip .tfplan files
		if strings.Contains(entry.Name(), ".tfplan") {
			continue
		}

		// Match .tfvars files
		if matches := tfvarsRegex.FindStringSubmatch(entry.Name()); matches != nil {
			configs = append(configs, matches[1]) // Return without .tfvars extension
		}
	}

	if len(configs) == 0 {
		return fmt.Errorf("no config files found in: %s", configPath)
	}

	for _, config := range configs {
		fmt.Println(config)
	}
	return nil
}

// SuggestActions lists available terraform actions
func (c *Completion) SuggestActions() error {
	actions := []string{
		"init", "plan", "apply", "apply_plan", "destroy", "output",
		"get", "workspace", "providers", "import", "taint", "untaint",
		"state", "refresh", "validate", "fmt", "format", "show",
	}

	for _, action := range actions {
		fmt.Println(action)
	}
	return nil
}

// SuggestWorkspace suggests workspace override format
func (c *Completion) SuggestWorkspace() error {
	fmt.Println("workspace=default")
	return nil
}

// SuggestRepo suggests the repository name (for compatibility with original bash completion)
func (c *Completion) SuggestRepo() error {
	// Extract repository name from project directory
	repoName := filepath.Base(c.config.ProjectDir)
	fmt.Println(repoName)
	return nil
}

// SuggestConfigCommands lists available config subcommands
func (c *Completion) SuggestConfigCommands() error {
	commands := []string{
		"convert", "init", "validate",
	}

	for _, cmd := range commands {
		fmt.Println(cmd)
	}
	return nil
}

// SuggestConfigInitFormats lists available config init formats
func (c *Completion) SuggestConfigInitFormats() error {
	formats := []string{
		"yaml", "legacy",
	}

	for _, format := range formats {
		fmt.Println(format)
	}
	return nil
}
