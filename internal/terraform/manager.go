package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sorinlg/tf-manage2/internal/config"
	"github.com/sorinlg/tf-manage2/internal/framework"
)

// Manager handles terraform operations with tf-manage conventions
type Manager struct {
	config *config.Config
}

// NewManager creates a new terraform manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// Command represents a terraform command to execute
type Command struct {
	Product        string
	Module         string
	Env            string
	ModuleInstance string
	Action         string
	ActionFlags    string
	Workspace      string
}

// Execute runs the terraform command with tf-manage conventions
func (m *Manager) Execute(cmd *Command) error {
	framework.Info(fmt.Sprintf("Detected exec mode: %s", m.detectExecMode()))

	// Validate the command
	if err := m.validateCommand(cmd); err != nil {
		return err
	}

	// Compute paths
	paths := m.computePaths(cmd)

	// Generate workspace name
	workspaceName := m.generateWorkspace(cmd, paths)

	framework.Info("*** Terraform ***")
	framework.Info(fmt.Sprintf("Running from \"%s\"", paths.ModulePath))

	// Change to module directory
	if err := os.Chdir(paths.ModulePath); err != nil {
		return fmt.Errorf("failed to change to module directory %s: %w", paths.ModulePath, err)
	}

	framework.Info(fmt.Sprintf("Executing terraform %s", cmd.Action))

	// For init command, run terraform init first, then handle workspace
	if cmd.Action == "init" {
		if err := m.terraformInit(cmd, paths); err != nil {
			return fmt.Errorf("failed to initialize terraform: %w", err)
		}

		// After init, ensure workspace exists and is selected
		if err := m.ensureWorkspace(workspaceName); err != nil {
			return fmt.Errorf("failed to ensure workspace: %w", err)
		}

		// Init is complete, return
		return nil
	}

	// For all other commands, ensure workspace exists first, then execute
	if err := m.ensureWorkspace(workspaceName); err != nil {
		return fmt.Errorf("failed to ensure workspace: %w", err)
	}

	// Execute the terraform command
	return m.executeTerraformAction(cmd, paths, workspaceName)
}

// Paths holds all the computed paths for the command
type Paths struct {
	ModulePath    string
	EnvPath       string
	ModuleEnvPath string
	VarFile       string
	PlanFile      string
}

func (m *Manager) validateCommand(cmd *Command) error {
	// Check product exists
	productPath := filepath.Join(m.config.GetEnvPath(), cmd.Product)

	// Debug output
	framework.Debug(fmt.Sprintf("Project dir: %s", m.config.ProjectDir))
	framework.Debug(fmt.Sprintf("Env rel path: %s", m.config.EnvRelPath))
	framework.Debug(fmt.Sprintf("Env full path: %s", m.config.GetEnvPath()))
	framework.Debug(fmt.Sprintf("Product path: %s", productPath))

	flags := framework.DefaultCmdFlags()
	flags.PrintOutput = false
	flags.PrintMessage = false

	result := framework.RunNative(
		framework.NativeTestDir(productPath),
		fmt.Sprintf("Checking product %s is valid", framework.AddEmphasisBlue(cmd.Product)),
		flags,
		fmt.Sprintf("Product path \"%s\" was not found!", framework.AddEmphasisBlue(productPath)),
	)

	if !result.Success {
		return fmt.Errorf("product validation failed")
	}

	// Check repo is valid
	if m.config.RepoName == "" {
		framework.Error("Repo name is empty!")
		return fmt.Errorf("repo validation failed")
	}

	result = framework.RunNative(
		framework.NativeTestNotEmpty(m.config.RepoName),
		fmt.Sprintf("Checking repo %s is valid", framework.AddEmphasisBlue(m.config.RepoName)),
		flags,
		"Component is empty. Make sure the first argument is set to a non-null string",
	)

	if !result.Success {
		return fmt.Errorf("repo validation failed")
	}

	// Check module exists
	modulePath := filepath.Join(m.config.GetModulePath(), cmd.Module)
	result = framework.RunNative(
		framework.NativeTestDir(modulePath),
		fmt.Sprintf("Checking module %s exists", framework.AddEmphasisBlue(cmd.Module)),
		flags,
		fmt.Sprintf("Module path \"%s\" was not found!", framework.AddEmphasisBlue(modulePath)),
	)

	if !result.Success {
		return fmt.Errorf("module validation failed")
	}

	// Check environment exists
	envPath := filepath.Join(m.config.GetEnvPath(), cmd.Product, cmd.Env)
	result = framework.RunNative(
		framework.NativeTestDir(envPath),
		fmt.Sprintf("Checking environment %s exists", framework.AddEmphasisBlue(cmd.Env)),
		flags,
		fmt.Sprintf("Environment path \"%s\" was not found!", framework.AddEmphasisBlue(envPath)),
	)

	if !result.Success {
		return fmt.Errorf("environment validation failed")
	}

	// Check config file exists
	varFile := filepath.Join(envPath, cmd.Module, cmd.ModuleInstance+".tfvars")
	result = framework.RunNative(
		framework.NativeTestFile(varFile),
		fmt.Sprintf("Checking config %s.tfvars exists", framework.AddEmphasisBlue(cmd.ModuleInstance)),
		flags,
		fmt.Sprintf("Config file \"%s\" was not found!", framework.AddEmphasisBlue(varFile)),
	)

	if !result.Success {
		return fmt.Errorf("config validation failed")
	}

	return nil
}

func (m *Manager) computePaths(cmd *Command) *Paths {
	modulePath := filepath.Join(m.config.GetModulePath(), cmd.Module)
	envPath := filepath.Join(m.config.GetEnvPath(), cmd.Product, cmd.Env)
	moduleEnvPath := filepath.Join(envPath, cmd.Module)
	varFile := filepath.Join(moduleEnvPath, cmd.ModuleInstance+".tfvars")
	planFile := filepath.Join(moduleEnvPath, cmd.ModuleInstance+".tfvars.tfplan")

	return &Paths{
		ModulePath:    modulePath,
		EnvPath:       envPath,
		ModuleEnvPath: moduleEnvPath,
		VarFile:       varFile,
		PlanFile:      planFile,
	}
}

func (m *Manager) generateWorkspace(cmd *Command, paths *Paths) string {
	// Replace forward slashes with double underscores in env path
	envSanitized := strings.ReplaceAll(cmd.Env, "/", "__")

	// Generate workspace name: {product}.{repo}.{module}.{env}.{module_instance}
	workspace := fmt.Sprintf("%s.%s.%s.%s.%s",
		cmd.Product,
		m.config.RepoName,
		cmd.Module,
		envSanitized,
		cmd.ModuleInstance,
	)

	return workspace
}

func (m *Manager) detectExecMode() string {
	if os.Getenv("TF_EXEC_MODE_OVERRIDE") != "" {
		return "unattended"
	}

	user := os.Getenv("USER")
	if user == "jenkins" {
		return "unattended"
	}

	return "operator"
}

func (m *Manager) ensureWorkspace(workspaceName string) error {
	// Check if workspace exists
	flags := framework.DefaultCmdFlags()
	flags.PrintOutput = false
	flags.PrintMessage = false

	result := framework.RunCmd(
		"terraform workspace list",
		"Checking available workspaces",
		flags,
	)

	if !result.Success {
		return fmt.Errorf("failed to list workspaces")
	}

	// Check if workspace already exists (look for exact match with proper formatting)
	lines := strings.Split(result.Output, "\n")
	for _, line := range lines {
		// Workspace list shows current workspace with "*" and others with spaces
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "* ") {
			line = strings.TrimPrefix(line, "* ")
		}
		if line == workspaceName {
			// Workspace exists, select it if not already current
			if !strings.HasPrefix(strings.TrimSpace(line), "* ") {
				return m.selectWorkspace(workspaceName)
			}
			// Already current workspace
			return nil
		}
	}

	// Create new workspace (terraform workspace new also selects it)
	flags = framework.DefaultCmdFlags()
	flags.ValidExitCodes = []int{0, 1} // Allow exit code 1 for already existing workspace

	result = framework.RunCmd(
		fmt.Sprintf("terraform workspace new %s", workspaceName),
		fmt.Sprintf("Creating workspace %s", framework.AddEmphasisBlue(workspaceName)),
		flags,
		fmt.Sprintf("Failed to create workspace %s", workspaceName),
	)

	// If workspace already exists, try to select it
	if !result.Success && strings.Contains(result.Error, "already exists") {
		return m.selectWorkspace(workspaceName)
	}

	if !result.Success {
		return fmt.Errorf("failed to create workspace %s", workspaceName)
	}

	return nil
}

func (m *Manager) selectWorkspace(workspaceName string) error {
	result := framework.RunCmd(
		fmt.Sprintf("terraform workspace select %s", workspaceName),
		fmt.Sprintf("Selecting workspace %s", framework.AddEmphasisBlue(workspaceName)),
		framework.DefaultCmdFlags(),
		fmt.Sprintf("Failed to select workspace %s", workspaceName),
	)

	if !result.Success {
		return fmt.Errorf("failed to select workspace %s", workspaceName)
	}

	return nil
}

func (m *Manager) executeTerraformAction(cmd *Command, paths *Paths, workspaceName string) error {
	switch cmd.Action {
	case "init":
		return m.terraformInit(cmd, paths)
	case "plan":
		return m.terraformPlan(cmd, paths)
	case "apply":
		return m.terraformApply(cmd, paths)
	case "destroy":
		return m.terraformDestroy(cmd, paths)
	case "output":
		return m.terraformOutput(cmd, paths)
	case "import":
		return m.terraformImport(cmd, paths)
	case "taint":
		return m.terraformTaint(cmd, paths)
	case "untaint":
		return m.terraformUntaint(cmd, paths)
	case "state":
		return m.terraformState(cmd, paths)
	case "refresh":
		return m.terraformRefresh(cmd, paths)
	case "validate":
		return m.terraformValidate(cmd, paths)
	case "fmt", "format":
		return m.terraformFormat(cmd, paths)
	case "show":
		return m.terraformShow(cmd, paths)
	default:
		return fmt.Errorf("unsupported terraform action: %s", cmd.Action)
	}
}

func (m *Manager) terraformInit(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform init"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Initializing terraform",
		framework.DefaultCmdFlags(),
		"Terraform init failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform init failed")
	}

	return nil
}

func (m *Manager) terraformPlan(cmd *Command, paths *Paths) error {
	terraformCmd := fmt.Sprintf("terraform plan -var-file=\"%s\" -out=\"%s\"", paths.VarFile, paths.PlanFile)
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Planning terraform changes",
		framework.DefaultCmdFlags(),
		"Terraform plan failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform plan failed")
	}

	return nil
}

func (m *Manager) terraformApply(cmd *Command, paths *Paths) error {
	var terraformCmd string

	// Check if plan file exists
	if _, err := os.Stat(paths.PlanFile); err == nil {
		// Use plan file
		terraformCmd = fmt.Sprintf("terraform apply \"%s\"", paths.PlanFile)
	} else {
		// Apply directly with var file
		terraformCmd = fmt.Sprintf("terraform apply -var-file=\"%s\"", paths.VarFile)
		if cmd.ActionFlags != "" {
			terraformCmd += " " + cmd.ActionFlags
		}
	}

	result := framework.RunCmd(
		terraformCmd,
		"Applying terraform changes",
		framework.DefaultCmdFlags(),
		"Terraform apply failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform apply failed")
	}

	return nil
}

func (m *Manager) terraformDestroy(cmd *Command, paths *Paths) error {
	terraformCmd := fmt.Sprintf("terraform destroy -var-file=\"%s\"", paths.VarFile)
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Destroying terraform resources",
		framework.DefaultCmdFlags(),
		"Terraform destroy failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform destroy failed")
	}

	return nil
}

func (m *Manager) terraformOutput(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform output"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Getting terraform outputs",
		framework.DefaultCmdFlags(),
		"Terraform output failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform output failed")
	}

	return nil
}

func (m *Manager) terraformImport(cmd *Command, paths *Paths) error {
	terraformCmd := fmt.Sprintf("terraform import -var-file=\"%s\"", paths.VarFile)
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Importing terraform resource",
		framework.DefaultCmdFlags(),
		"Terraform import failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform import failed")
	}

	return nil
}

func (m *Manager) terraformTaint(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform taint"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Tainting terraform resource",
		framework.DefaultCmdFlags(),
		"Terraform taint failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform taint failed")
	}

	return nil
}

func (m *Manager) terraformUntaint(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform untaint"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Untainting terraform resource",
		framework.DefaultCmdFlags(),
		"Terraform untaint failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform untaint failed")
	}

	return nil
}

func (m *Manager) terraformState(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform state"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Managing terraform state",
		framework.DefaultCmdFlags(),
		"Terraform state command failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform state command failed")
	}

	return nil
}

func (m *Manager) terraformRefresh(cmd *Command, paths *Paths) error {
	terraformCmd := fmt.Sprintf("terraform refresh -var-file=\"%s\"", paths.VarFile)
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Refreshing terraform state",
		framework.DefaultCmdFlags(),
		"Terraform refresh failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform refresh failed")
	}

	return nil
}

func (m *Manager) terraformValidate(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform validate"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Validating terraform configuration",
		framework.DefaultCmdFlags(),
		"Terraform validate failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform validate failed")
	}

	return nil
}

func (m *Manager) terraformFormat(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform fmt"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Formatting terraform files",
		framework.DefaultCmdFlags(),
		"Terraform fmt failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform fmt failed")
	}

	return nil
}

func (m *Manager) terraformShow(cmd *Command, paths *Paths) error {
	var terraformCmd string

	// Check if plan file exists
	if _, err := os.Stat(paths.PlanFile); err == nil {
		// Show plan file
		terraformCmd = fmt.Sprintf("terraform show \"%s\"", paths.PlanFile)
	} else {
		// Show current state
		terraformCmd = "terraform show"
	}

	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Showing terraform state/plan",
		framework.DefaultCmdFlags(),
		"Terraform show failed",
	)

	if !result.Success {
		return fmt.Errorf("terraform show failed")
	}

	return nil
}
