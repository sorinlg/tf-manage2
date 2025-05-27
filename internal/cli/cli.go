package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/sorinlg/tf-manage2/internal/config"
	"github.com/sorinlg/tf-manage2/internal/terraform"
)

// Execute is the main CLI entry point
func Execute() error {
	args := os.Args[1:]

	if len(args) == 0 {
		return showUsage()
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
