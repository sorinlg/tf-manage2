package cli

import (
	"fmt"
	"os"
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
	return tfm.Execute(cmd)
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

ARGUMENTS:
    product           Product name
    module            Terraform module name
    env               Environment (dev, staging, prod, etc.)
    module_instance   Module instance identifier
    action            Terraform action (init, plan, apply, destroy, etc.)
    workspace         Optional workspace override (format: workspace=name)

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
	default:
		return nil // Silently fail for unknown commands
	}
}
