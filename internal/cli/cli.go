package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sorinlg/tf-manage2/internal/config"
	"github.com/sorinlg/tf-manage2/internal/terraform"
)

// Version information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

// SetVersionInfo sets the version information for the CLI
func SetVersionInfo(v, c, d, b string) {
	version = v
	commit = c
	date = d
	builtBy = b
}

// Execute is the main CLI entry point
func Execute() error {
	args := os.Args[1:]

	if len(args) == 0 {
		return showUsage()
	}

	// Handle version flag
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("tf-manage2 version %s\n", version)
		if commit != "none" {
			fmt.Printf("  commit: %s\n", commit)
		}
		if date != "unknown" {
			fmt.Printf("  built: %s\n", date)
		}
		if builtBy != "unknown" {
			fmt.Printf("  built by: %s\n", builtBy)
		}
		return nil
	}

	// Handle help flag
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		return showHelp()
	}

	// Handle completion commands
	if len(args) >= 1 && args[0] == "__complete" {
		return handleCompletion(args[1:])
	}

	// Handle config commands
	if len(args) >= 1 && args[0] == "config" {
		return handleConfigCommand(args[1:])
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Parse command arguments
	cmd, err := parseCommand(args)
	if err != nil {
		return err
	}

	// Create terraform manager
	tfm := terraform.NewManager(cfg)

	// Execute the command
	err = tfm.Execute(cmd)

	// Check if this is an exit code error and exit with the specific code
	if exitCodeErr, ok := err.(*terraform.ExitCodeError); ok {
		// For exit code errors, we want to preserve the specific exit code
		os.Exit(exitCodeErr.ExitCode)
	}

	return err
}

// Command represents a tf-manage command
type Command = terraform.Command

func parseCommand(args []string) (*terraform.Command, error) {
	if len(args) < 5 {
		return nil, fmt.Errorf("insufficient arguments")
	}

	if len(args) > 6 {
		return nil, fmt.Errorf("too many arguments")
	}

	cmd := &terraform.Command{
		Product:        args[0],
		Module:         args[1],
		Env:            args[2],
		ModuleInstance: args[3],
	}

	// Parse action and action flags
	actionRaw := args[4]
	actionParts := strings.Fields(actionRaw)
	if len(actionParts) > 0 {
		cmd.Action = actionParts[0]
		if len(actionParts) > 1 {
			cmd.ActionFlags = strings.Join(actionParts[1:], " ")
		}
	}

	// Optional workspace override
	if len(args) == 6 {
		cmd.Workspace = strings.TrimPrefix(args[5], "workspace=")
	}

	return cmd, nil
}

func showUsage() error {
	binaryName := os.Args[0]
	return fmt.Errorf("Usage: %s <product> <module> <env> <module_instance> <action> [workspace]", binaryName)
}

func showHelp() error {
	fmt.Printf(`tf-manage2 - Terraform workspace manager

USAGE:
    tf <product> <module> <env> <module_instance> <action> [workspace]
    tf config <command>

ARGUMENTS:
    product           Product name
    module            Terraform module name
    env               Environment (dev, staging, prod, etc.)
    module_instance   Module instance identifier
    action            Terraform action (init, plan, apply, destroy, etc.)
    workspace         Optional workspace override (format: workspace=name)

CONFIGURATION COMMANDS:
    tf config convert       Convert legacy .tfm.conf to .tfm.yaml
    tf config init yaml     Create new .tfm.yaml configuration
    tf config init legacy   Create new .tfm.conf configuration (deprecated)
    tf config validate      Validate current configuration

EXAMPLES:
    tf product1 sample_module dev instance_x init
    tf product1 sample_module dev instance_x plan
    tf product1 sample_module dev instance_x apply
    tf product1 sample_module dev instance_x destroy
    tf product1 sample_module dev instance_x plan workspace=custom

FLAGS:
    -h, --help        Show this help message
    -v, --version     Show version information

ENVIRONMENT VARIABLES:
    TF_EXEC_MODE_OVERRIDE=1    Force unattended mode (auto-approve)

CONFIGURATION:
    tf-manage2 supports both legacy (.tfm.conf) and modern (.tfm.yaml) formats.
    The legacy format is deprecated and will be removed in v3.0.
    Use 'tf config convert' to migrate to the new format.

For more information, see: https://github.com/sorinlg/tf-manage2
`)
	return nil
}

// handleCompletion handles bash completion requests
func handleCompletion(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("completion command required")
	}

	// Try to load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config fails to load, we're likely not in a tf-manage workspace
		// Don't output completion suggestions but also don't error
		// The bash completion script will handle this gracefully
		return nil
	}

	// Create completion handler
	completion := NewCompletion(cfg)

	switch args[0] {
	case "products":
		return completion.SuggestProducts()
	case "modules":
		return completion.SuggestModules()
	case "environments":
		if len(args) < 3 {
			return nil // Silently fail if not enough args
		}
		return completion.SuggestEnvironments(args[1], args[2])
	case "configs":
		if len(args) < 4 {
			return nil // Silently fail if not enough args
		}
		return completion.SuggestConfigs(args[1], args[2], args[3])
	case "actions":
		return completion.SuggestActions()
	case "workspace":
		return completion.SuggestWorkspace()
	case "repo":
		return completion.SuggestRepo()
	case "config":
		return completion.SuggestConfigCommands()
	case "config_init":
		return completion.SuggestConfigInitFormats()
	default:
		return nil // Silently fail for unknown commands
	}
}

// handleConfigCommand handles config-related commands
func handleConfigCommand(args []string) error {
	if len(args) == 0 {
		return showConfigHelp()
	}

	// Handle help flags
	if args[0] == "--help" || args[0] == "-h" {
		return showConfigHelp()
	}

	switch args[0] {
	case "convert":
		return handleConfigConvert()
	case "init":
		if len(args) < 2 {
			return fmt.Errorf("usage: tf config init <format>\nformats: yaml, legacy")
		}
		return handleConfigInit(args[1])
	case "validate":
		return handleConfigValidate()
	default:
		return fmt.Errorf("unknown config command: %s\nRun 'tf config --help' for usage", args[0])
	}
}

// handleConfigConvert converts legacy .tfm.conf to .tfm.yaml
func handleConfigConvert() error {
	projectDir, err := findProjectDir()
	if err != nil {
		return fmt.Errorf("failed to find project directory: %w", err)
	}

	return config.ConvertLegacyToYAML(projectDir)
}

// handleConfigInit creates a new configuration file
func handleConfigInit(format string) error {
	projectDir, err := findProjectDir()
	if err != nil {
		return fmt.Errorf("failed to find project directory: %w", err)
	}

	switch format {
	case "yaml":
		configPath := filepath.Join(projectDir, ".tfm.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("YAML config file already exists at %s", configPath)
		}

		projectName := filepath.Base(projectDir)
		cfg := &config.Config{
			ConfigVersion: "2.0",
			RepoName:      projectName,
			EnvRelPath:    "terraform/environments",
			ModuleRelPath: "terraform/modules",
		}

		if err := config.WriteYAMLConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to create YAML config: %w", err)
		}

		fmt.Printf("✅ Created YAML configuration at %s\n", configPath)
		return nil

	case "legacy":
		configPath := filepath.Join(projectDir, ".tfm.conf")
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("legacy config file already exists at %s", configPath)
		}

		projectName := filepath.Base(projectDir)
		content := fmt.Sprintf(`#!/bin/bash
export __tfm_repo_name='%s'
export __tfm_env_rel_path='terraform/environments'
export __tfm_module_rel_path='terraform/modules'
`, projectName)

		if err := os.WriteFile(configPath, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed to create legacy config: %w", err)
		}

		fmt.Printf("⚠️  Created legacy configuration at %s\n", configPath)
		fmt.Printf("   Note: Legacy format is deprecated. Consider using 'tf config init yaml' instead.\n")
		return nil

	default:
		return fmt.Errorf("unknown format: %s\nSupported formats: yaml, legacy", format)
	}
}

// handleConfigValidate validates the current configuration
func handleConfigValidate() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	fmt.Printf("✅ Configuration is valid\n")
	fmt.Printf("   Config file: %s\n", cfg.ConfigPath)
	fmt.Printf("   Repository:  %s\n", cfg.RepoName)
	fmt.Printf("   Environments: %s\n", cfg.EnvRelPath)
	fmt.Printf("   Modules:     %s\n", cfg.ModuleRelPath)

	if cfg.ConfigVersion != "" {
		fmt.Printf("   Version:     %s\n", cfg.ConfigVersion)
	} else {
		fmt.Printf("   Version:     legacy (consider migrating with 'tf config convert')\n")
	}

	return nil
}

// showConfigHelp shows help for config commands
func showConfigHelp() error {
	fmt.Printf(`tf-manage2 config commands

USAGE:
    tf config <command>

COMMANDS:
    convert     Convert legacy .tfm.conf to .tfm.yaml format
    init        Create a new configuration file (yaml|legacy)
    validate    Validate the current configuration

EXAMPLES:
    tf config convert              # Convert .tfm.conf to .tfm.yaml
    tf config init yaml           # Create new .tfm.yaml file
    tf config init legacy         # Create new .tfm.conf file
    tf config validate            # Check current configuration

MIGRATION:
    The legacy .tfm.conf format is deprecated and will be removed in v3.0.
    Use 'tf config convert' to migrate to the new YAML format.

For more information, see: https://github.com/sorinlg/tf-manage2
`)
	return nil
}

// findProjectDir finds the git repository root directory
// This is a duplicate of the function in config package to avoid circular imports
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
