package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sorinlg/tf-manage2/internal/config"
	"github.com/sorinlg/tf-manage2/internal/framework"
)

// ExitCodeError represents an error that carries a specific exit code
type ExitCodeError struct {
	Message  string
	ExitCode int
}

func (e *ExitCodeError) Error() string {
	return e.Message
}

// NewExitCodeError creates a new error with the specified exit code
func NewExitCodeError(message string, exitCode int) *ExitCodeError {
	return &ExitCodeError{
		Message:  message,
		ExitCode: exitCode,
	}
}

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

	// Show Terraform CLI version in the banner
	ver := getTerraformVersion()
	if ver != "unknown" && !strings.HasPrefix(ver, "v") {
		ver = "v" + ver
	}
	framework.Info(fmt.Sprintf("*** Terraform %s ***", ver))
	framework.Info(fmt.Sprintf("Running from \"%s\"", paths.ModulePath))

	// Change to module directory
	if err := os.Chdir(paths.ModulePath); err != nil {
		return fmt.Errorf("failed to change to module directory %s: %w", paths.ModulePath, err)
	}

	framework.Info(fmt.Sprintf("Executing terraform %s", cmd.Action))

	// Check terraform workspace exists and is active
	// Skip workspace validation for workspace, init, and fmt commands (matching bash __tf_controller logic)
	if cmd.Action != "workspace" && cmd.Action != "init" && cmd.Action != "fmt" {
		if err := m.ensureWorkspace(workspaceName); err != nil {
			return fmt.Errorf("failed to ensure workspace: %w", err)
		}
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
	unattended := framework.AddEmphasisRed("unattended")
	interactive := framework.AddEmphasisGreen("operator")

	// Allow explicit override
	if os.Getenv("TF_EXEC_MODE_OVERRIDE") != "" {
		return unattended
	}

	// Check for CI/CD environment variables
	if m.isRunningInCI() {
		return unattended
	}

	// Default to interactive operator mode
	return interactive
}

// isRunningInCI detects if we're running in any popular CI/CD system
func (m *Manager) isRunningInCI() bool {
	// GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return true
	}

	// GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		return true
	}

	// CircleCI
	if os.Getenv("CIRCLECI") == "true" {
		return true
	}

	// Travis CI
	if os.Getenv("TRAVIS") == "true" {
		return true
	}

	// Azure DevOps / Azure Pipelines
	if os.Getenv("TF_BUILD") == "True" {
		return true
	}

	// Jenkins (multiple ways to detect)
	if os.Getenv("JENKINS_URL") != "" || os.Getenv("BUILD_NUMBER") != "" {
		return true
	}

	// Bamboo
	if os.Getenv("bamboo_buildKey") != "" {
		return true
	}

	// TeamCity
	if os.Getenv("TEAMCITY_VERSION") != "" {
		return true
	}

	// Buildkite
	if os.Getenv("BUILDKITE") == "true" {
		return true
	}

	// Drone CI
	if os.Getenv("DRONE") == "true" {
		return true
	}

	// AWS CodeBuild
	if os.Getenv("CODEBUILD_BUILD_ID") != "" {
		return true
	}

	// Generic CI indicator (set by many CI systems)
	if os.Getenv("CI") == "true" || os.Getenv("CI") == "1" {
		return true
	}

	// Fallback: Legacy Jenkins detection by username
	if os.Getenv("USER") == "jenkins" {
		return true
	}

	return false
}

// getTerraformVersion returns the Terraform CLI version found on PATH.
// It first tries `terraform version -json` and falls back to parsing `terraform version` output.
func getTerraformVersion() string {
	// Quiet flags: we only need the output string
	flags := framework.DefaultCmdFlags()
	flags.PrintMessage = false
	flags.PrintOutput = false
	flags.PrintStatus = false
	flags.PrintOutcome = false

	// Preferred: JSON output
	res := framework.RunCmd("terraform version -json", "Detecting Terraform version", flags)
	if res != nil && res.Success {
		var payload struct {
			TerraformVersion string `json:"terraform_version"`
		}
		if err := json.Unmarshal([]byte(res.Output), &payload); err == nil {
			if payload.TerraformVersion != "" {
				return payload.TerraformVersion
			}
		}
	}

	// Fallback: plain text
	res = framework.RunCmd("terraform version", "Detecting Terraform version", flags)
	if res != nil && res.Success {
		// Typical first line: "Terraform v1.9.5" or "Terraform v1.5.7 on darwin_amd64"
		out := strings.TrimSpace(res.Output)
		if idx := strings.IndexByte(out, '\n'); idx >= 0 {
			out = strings.TrimSpace(out[:idx])
		}
		for _, tok := range strings.Fields(out) {
			if strings.HasPrefix(tok, "v") {
				tok = strings.TrimRight(tok, ",")
				return tok
			}
		}
	}

	return "unknown"
}

func (m *Manager) ensureWorkspace(workspaceName string) error {
	// Execute terraform workspace list command directly
	flags := framework.DefaultCmdFlags()
	flags.PrintOutput = false
	flags.PrintMessage = false
	flags.PrintStatus = true
	flags.PrintOutcome = false

	result := framework.RunCmd(
		"terraform workspace list",
		fmt.Sprintf("Checking workspace %s exists", framework.AddEmphasisBlue(workspaceName)),
		flags,
	)

	// Parse the workspace list output
	workspaceExists := false
	for _, line := range strings.Split(result.Output, "\n") {
		// Terraform workspace list format:
		// '* default' (current workspace has asterisk)
		// '  workspace1'
		// '  workspace2'
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "*") {
			trimmedLine = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "*"))
		}

		if trimmedLine == workspaceName {
			workspaceExists = true
			break
		}
	}

	// If workspace doesn't exist, create it
	if !workspaceExists {
		// Create new workspace
		flags = framework.DefaultCmdFlags()
		flags.PrintMessage = true
		flags.PrintStatus = true
		flags.PrintOutcome = false

		result = framework.RunCmd(
			fmt.Sprintf("terraform workspace new %s", workspaceName),
			fmt.Sprintf("Creating workspace %s", framework.AddEmphasisRed(workspaceName)),
			flags,
			"Could not create workspace!",
		)

		if !result.Success {
			return fmt.Errorf("failed to create workspace %s", workspaceName)
		}
	}

	// Select workspace using environment variable (same as bash version)
	os.Setenv("TF_WORKSPACE", workspaceName)
	framework.Info(fmt.Sprintf("Selecting workspace %s", framework.AddEmphasisBlue(workspaceName)))

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
	case "apply_plan":
		return m.terraformApplyPlan(cmd, paths)
	case "destroy":
		return m.terraformDestroy(cmd, paths)
	case "output":
		return m.terraformOutput(cmd, paths)
	case "get":
		return m.terraformGet(cmd, paths)
	case "workspace":
		return m.terraformWorkspace(cmd, paths)
	case "providers":
		return m.terraformProviders(cmd, paths)
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

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformPlan(cmd *Command, paths *Paths) error {
	extraVars := m.generateTfmExtraVars(cmd)
	terraformCmd := fmt.Sprintf("terraform plan -var-file=\"%s\" -out=\"%s\" %s", paths.VarFile, paths.PlanFile, extraVars)
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Planning terraform changes",
		framework.DefaultCmdFlags(),
		"Terraform plan failed",
	)

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformApply(cmd *Command, paths *Paths) error {
	// Apply directly with var file (not using plan file)
	extraVars := m.generateTfmExtraVars(cmd)
	terraformCmd := fmt.Sprintf("terraform apply -var-file=\"%s\" %s", paths.VarFile, extraVars)

	// Add extra arguments in case we're running in "unattended" mode
	if m.detectExecMode() == "unattended" {
		terraformCmd += " -input=false -auto-approve"
	}

	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	// Notify user about the action
	framework.Info("Executing terraform apply")
	framework.Info("This will affect infrastructure resources.")

	var result *framework.CmdResult

	// Use interactive runner for operator mode, regular runner for unattended mode
	if m.detectExecMode() == "unattended" {
		flags := framework.DefaultCmdFlags()
		flags.PrintMessage = false

		result = framework.RunCmd(
			terraformCmd,
			"Applying terraform changes",
			flags,
			"Terraform apply failed",
		)
	} else {
		// Interactive mode - use special interactive runner
		result = framework.RunCmdInteractive(
			terraformCmd,
			"Applying terraform changes",
			"Terraform apply failed",
		)
	}

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformApplyPlan(cmd *Command, paths *Paths) error {
	// Apply using the plan file
	terraformCmd := fmt.Sprintf("terraform apply \"%s\"", paths.PlanFile)

	// Add extra arguments in case we're running in "unattended" mode
	if m.detectExecMode() == "unattended" {
		terraformCmd += " -input=false"
	}

	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	flags := framework.DefaultCmdFlags()
	flags.PrintMessage = false

	// Notify user about the action
	framework.Info("Executing terraform apply")
	framework.Info("This will affect infrastructure resources.")

	result := framework.RunCmd(
		terraformCmd,
		"Applying terraform changes",
		flags,
		"Terraform apply failed",
	)

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformDestroy(cmd *Command, paths *Paths) error {
	extraVars := m.generateTfmExtraVars(cmd)
	terraformCmd := fmt.Sprintf("terraform destroy -var-file=\"%s\" %s", paths.VarFile, extraVars)

	// Add extra arguments in case we're running in "unattended" mode
	if m.detectExecMode() == "unattended" {
		terraformCmd += " -auto-approve"
	}

	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	// Notify user about the action
	framework.Info("Executing terraform destroy")
	framework.Info("This will DESTROY infrastructure resources.")

	var result *framework.CmdResult

	// Use interactive runner for operator mode, regular runner for unattended mode
	if m.detectExecMode() == "unattended" {
		flags := framework.DefaultCmdFlags()
		flags.PrintMessage = false

		result = framework.RunCmd(
			terraformCmd,
			"Destroying terraform resources",
			flags,
			"Terraform destroy failed",
		)
	} else {
		// Interactive mode - use special interactive runner
		result = framework.RunCmdInteractive(
			terraformCmd,
			"Destroying terraform resources",
			"Terraform destroy failed",
		)
	}

	return NewExitCodeError("command failed", result.ExitCode)
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

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformImport(cmd *Command, paths *Paths) error {
	extraVars := m.generateTfmExtraVars(cmd)
	terraformCmd := fmt.Sprintf("terraform import -var-file=\"%s\" %s", paths.VarFile, extraVars)

	// Add extra arguments in case we're running in "unattended" mode
	if m.detectExecMode() == "unattended" {
		terraformCmd += " -input=false -auto-approve"
	}

	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	// Notify user about the action
	framework.Info("Executing terraform import")
	framework.Info("This will affect infrastructure resources.")

	var result *framework.CmdResult

	// Use interactive runner for operator mode, regular runner for unattended mode
	if m.detectExecMode() == "unattended" {
		flags := framework.DefaultCmdFlags()
		flags.PrintMessage = false

		result = framework.RunCmd(
			terraformCmd,
			"Importing terraform resource",
			flags,
			"Terraform import failed",
		)
	} else {
		// Interactive mode - use special interactive runner
		result = framework.RunCmdInteractive(
			terraformCmd,
			"Importing terraform resource",
			"Terraform import failed",
		)
	}

	return NewExitCodeError("command failed", result.ExitCode)
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

	return NewExitCodeError("command failed", result.ExitCode)
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

	return NewExitCodeError("command failed", result.ExitCode)
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

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformRefresh(cmd *Command, paths *Paths) error {
	extraVars := m.generateTfmExtraVars(cmd)
	terraformCmd := fmt.Sprintf("terraform refresh -var-file=\"%s\" %s", paths.VarFile, extraVars)
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Refreshing terraform state",
		framework.DefaultCmdFlags(),
		"Terraform refresh failed",
	)

	return NewExitCodeError("command failed", result.ExitCode)
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

	return NewExitCodeError("command failed", result.ExitCode)
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

	return NewExitCodeError("command failed", result.ExitCode)
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

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformGet(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform get"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Getting terraform modules",
		framework.DefaultCmdFlags(),
		"Terraform get failed",
	)

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformWorkspace(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform workspace"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Managing terraform workspace",
		framework.DefaultCmdFlags(),
		"Terraform workspace command failed",
	)

	return NewExitCodeError("command failed", result.ExitCode)
}

func (m *Manager) terraformProviders(cmd *Command, paths *Paths) error {
	terraformCmd := "terraform providers"
	if cmd.ActionFlags != "" {
		terraformCmd += " " + cmd.ActionFlags
	}

	result := framework.RunCmd(
		terraformCmd,
		"Managing terraform providers",
		framework.DefaultCmdFlags(),
		"Terraform providers command failed",
	)

	return NewExitCodeError("command failed", result.ExitCode)
}

// generateTfmExtraVars creates the terraform variable flags for tf-manage integration
// This matches the bash version's _TFM_EXTRA_VARS functionality
func (m *Manager) generateTfmExtraVars(cmd *Command) string {
	return fmt.Sprintf("-var 'tfm_product=%s' -var 'tfm_repo=%s' -var 'tfm_module=%s' -var 'tfm_env=%s' -var 'tfm_module_instance=%s'",
		cmd.Product,
		m.config.RepoName,
		cmd.Module,
		cmd.Env,
		cmd.ModuleInstance,
	)
}
