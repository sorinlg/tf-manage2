package framework

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CmdFlags represents the configuration flags for command execution
type CmdFlags struct {
	Strict          bool   // Whether to exit on command failure
	PrintCmd        bool   // Whether to print the command being executed
	DecorateOutput  bool   // Whether to decorate command output
	PrintOutput     bool   // Whether to print command output
	PrintMessage    bool   // Whether to print the message
	PrintStatus     bool   // Whether to print status indicators
	PrintOutcome    bool   // Whether to print outcome (done/continuing...)
	StrictMessage   string // Message to show in strict mode on failure
	NoStrictMessage string // Message to show in non-strict mode on failure
	ValidExitCodes  []int  // List of valid exit codes (default: [0])
}

// DefaultCmdFlags returns the default command flags
func DefaultCmdFlags() *CmdFlags {
	return &CmdFlags{
		Strict:          false,
		PrintCmd:        false,
		DecorateOutput:  false,
		PrintOutput:     true,
		PrintMessage:    true,
		PrintStatus:     true,
		PrintOutcome:    false,
		StrictMessage:   "aborting...",
		NoStrictMessage: "continuing...",
		ValidExitCodes:  []int{0},
	}
}

// CmdResult represents the result of a command execution
type CmdResult struct {
	ExitCode int
	Success  bool
	Output   string
	Error    string
}

// RunCmd executes a system command with the specified flags and message
func RunCmd(command, message string, flags *CmdFlags, failMessage ...string) *CmdResult {
	if flags == nil {
		flags = DefaultCmdFlags()
	}

	// Print the message if enabled
	if flags.PrintMessage {
		Info(message)
	}

	// Print the command if enabled
	if flags.PrintCmd {
		Info(command)
	}

	// Execute the system command
	result := execSystemCommand(command, flags)

	// Parse and display status
	parseStatus(message, result, flags, failMessage...)

	return result
}

// RunCmdSilent executes a command silently (no output)
func RunCmdSilent(command, message string, failMessage ...string) *CmdResult {
	flags := DefaultCmdFlags()
	flags.PrintOutput = false
	flags.PrintMessage = false
	flags.PrintStatus = false
	flags.PrintOutcome = false

	return RunCmd(command, message, flags, failMessage...)
}

// RunCmdStrict executes a command in strict mode (exits on failure)
func RunCmdStrict(command, message string, failMessage ...string) *CmdResult {
	flags := DefaultCmdFlags()
	flags.Strict = true
	flags.PrintOutput = false
	flags.PrintMessage = false
	flags.PrintOutcome = false

	return RunCmd(command, message, flags, failMessage...)
}

// RunCmdSilentStrict executes a command silently and in strict mode
func RunCmdSilentStrict(command, message string, failMessage ...string) *CmdResult {
	flags := DefaultCmdFlags()
	flags.Strict = true
	flags.PrintOutput = false
	flags.PrintMessage = false
	flags.PrintStatus = false
	flags.PrintOutcome = false

	return RunCmd(command, message, flags, failMessage...)
}

// RunCmdInteractive executes a command that requires user interaction (stdin)
// It automatically disables decoration to ensure stdin works properly
func RunCmdInteractive(command, message string, failMessage ...string) *CmdResult {
	flags := DefaultCmdFlags()
	flags.DecorateOutput = false // Disable decoration for interactive commands
	flags.PrintOutput = true     // Keep output enabled for user feedback

	return RunCmd(command, message, flags, failMessage...)
}

// execSystemCommand executes the actual system command
func execSystemCommand(command string, flags *CmdFlags) *CmdResult {
	// Debug the command being executed
	Debug(fmt.Sprintf("Executing command: %s", command))

	// Always use shell execution for consistent quote and escape handling
	if strings.TrimSpace(command) == "" {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Error:    "empty command",
		}
	}

	cmd := exec.Command("sh", "-c", command)
	Debug(fmt.Sprintf("Using shell: sh -c '%s'", command))

	var output strings.Builder
	var errorOutput strings.Builder

	if flags.PrintOutput && flags.DecorateOutput {
		// If we need to decorate output, we need to capture and process it
		// However, for interactive commands, we should not use decoration
		// as it breaks stdin interaction
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return &CmdResult{
				ExitCode: 1,
				Success:  false,
				Error:    err.Error(),
			}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return &CmdResult{
				ExitCode: 1,
				Success:  false,
				Error:    err.Error(),
			}
		}

		// Note: stdin is not connected when using pipes for decoration
		// This means decorated output is incompatible with interactive commands

		if err := cmd.Start(); err != nil {
			return &CmdResult{
				ExitCode: 1,
				Success:  false,
				Error:    err.Error(),
			}
		}

		// Read and decorate stdout
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				output.WriteString(line + "\n")
				if flags.PrintOutput {
					decoratedLine := AddEmphasisBlue(fmt.Sprintf("[%s]", "cmd")) + " " + line
					fmt.Println(decoratedLine)
				}
			}
		}()

		// Read stderr
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				errorOutput.WriteString(line + "\n")
				if flags.PrintOutput {
					decoratedLine := AddEmphasisRed(fmt.Sprintf("[%s]", "err")) + " " + line
					fmt.Println(decoratedLine)
				}
			}
		}()

		err = cmd.Wait()

		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1
			}
		}

		// Check if exit code is valid
		success := false
		for _, validCode := range flags.ValidExitCodes {
			if exitCode == validCode {
				success = true
				break
			}
		}

		return &CmdResult{
			ExitCode: exitCode,
			Success:  success,
			Output:   output.String(),
			Error:    errorOutput.String(),
		}

	} else {
		// Standard execution
		var err error
		if flags.PrintOutput {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin // Enable interactive input
		}
		err = cmd.Run()

		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1
			}
		}

		// Check if exit code is valid
		success := false
		for _, validCode := range flags.ValidExitCodes {
			if exitCode == validCode {
				success = true
				break
			}
		}

		return &CmdResult{
			ExitCode: exitCode,
			Success:  success,
			Output:   output.String(),
			Error:    errorOutput.String(),
		}
	}
}

// CommandType represents the type of command to execute
type CommandType int

const (
	SystemCommand CommandType = iota
	NativeCommand
)

// NativeFunc represents a native Go function that can be executed
type NativeFunc func() *CmdResult

// RunNative executes a native Go function with the specified flags and message
func RunNative(nativeFunc NativeFunc, message string, flags *CmdFlags, failMessage ...string) *CmdResult {
	if flags == nil {
		flags = DefaultCmdFlags()
	}

	// Print the message if enabled
	if flags.PrintMessage {
		Info(message)
	}

	// Execute the native function
	result := nativeFunc()

	// Parse and display status
	parseStatus(message, result, flags, failMessage...)

	return result
}

// Enhanced native functions with better error reporting

// TestDir checks if a directory exists (replacement for "test -d")
func TestDir(path string) *CmdResult {
	Debug(fmt.Sprintf("Native directory check: %s", path))

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return &CmdResult{
			ExitCode: 0,
			Success:  true,
			Output:   fmt.Sprintf("Directory exists: %s", path),
			Error:    "",
		}
	} else {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Output:   "",
			Error:    fmt.Sprintf("Directory does not exist: %s", path),
		}
	}
}

// TestFile checks if a file exists (replacement for "test -f")
func TestFile(path string) *CmdResult {
	Debug(fmt.Sprintf("Native file check: %s", path))

	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return &CmdResult{
			ExitCode: 0,
			Success:  true,
			Output:   fmt.Sprintf("File exists: %s", path),
			Error:    "",
		}
	} else {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Output:   "",
			Error:    fmt.Sprintf("File does not exist: %s", path),
		}
	}
}

// TestNotEmpty checks if a string is not empty (replacement for "test ! -z")
func TestNotEmpty(value string) *CmdResult {
	Debug(fmt.Sprintf("Native non-empty check: '%s'", value))

	if value != "" {
		return &CmdResult{
			ExitCode: 0,
			Success:  true,
			Output:   fmt.Sprintf("String is not empty: '%s'", value),
			Error:    "",
		}
	} else {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Output:   "",
			Error:    "String is empty",
		}
	}
}

// Helper functions that return NativeFunc for easier usage

// NativeTestDir returns a NativeFunc that checks if a directory exists
func NativeTestDir(path string) NativeFunc {
	return func() *CmdResult {
		return TestDir(path)
	}
}

// NativeTestFile returns a NativeFunc that checks if a file exists
func NativeTestFile(path string) NativeFunc {
	return func() *CmdResult {
		return TestFile(path)
	}
}

// NativeTestNotEmpty returns a NativeFunc that checks if a string is not empty
func NativeTestNotEmpty(value string) NativeFunc {
	return func() *CmdResult {
		return TestNotEmpty(value)
	}
}

// parseStatus displays the status of command execution
func parseStatus(message string, result *CmdResult, flags *CmdFlags, failMessage ...string) {
	if !flags.PrintStatus {
		return
	}

	// Prepare status indicators
	var statusIndicator string
	var outcomeMessage string

	if result.Success {
		statusIndicator = fmt.Sprintf("[ %s ]", AddEmphasisGreen(CheckMark))
		outcomeMessage = "(done)"
	} else {
		statusIndicator = fmt.Sprintf("[ %s ]", AddEmphasisRed(CrossMark))
		if flags.Strict {
			outcomeMessage = fmt.Sprintf("(%s)", AddEmphasisRed(flags.StrictMessage))
		} else {
			outcomeMessage = fmt.Sprintf("(%s)", AddEmphasisRed(flags.NoStrictMessage))
		}
	}

	// Format the message with proper spacing (similar to bash version)
	// Calculate padding based on visual length (excluding ANSI codes)
	entrypoint := fmt.Sprintf("[%s]", GetEntrypointScript())
	entrypointWithColor := AddEmphasisGray(entrypoint)

	// Calculate the actual visual width needed
	messageVisualLength := getVisualLength(message)
	entrypointVisualLength := getVisualLength(entrypoint) // Use uncolored version for length
	statusVisualLength := 5                               // "[ ✓ ]" or "[ ✗ ]"

	// Total target width minus the parts we know
	totalWidth := 120
	paddingWidth := totalWidth - entrypointVisualLength - 1 - messageVisualLength - 1 - statusVisualLength

	// Ensure minimum padding
	if paddingWidth < 1 {
		paddingWidth = 1
	}

	format := fmt.Sprintf("%s %s%*s %s",
		entrypointWithColor,
		message,
		paddingWidth, "",
		statusIndicator)

	if flags.PrintOutcome && outcomeMessage != "" {
		format += " " + outcomeMessage
	}

	fmt.Println(format)

	// Handle failure
	if !result.Success {
		if len(failMessage) > 0 && failMessage[0] != "" {
			Error(failMessage[0])
		}

		if flags.Strict {
			os.Exit(result.ExitCode)
		}
	}
}
