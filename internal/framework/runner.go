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

// RunCmd executes a command with the specified flags and message
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

	// Execute the command
	result := execCommand(command, flags)

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

// execCommand executes the actual command
func execCommand(command string, flags *CmdFlags) *CmdResult {
	// Debug the command being executed
	Debug(fmt.Sprintf("Executing command: %s", command))

	// For shell commands like test, we need to use sh -c to handle quotes properly
	var cmd *exec.Cmd
	if strings.Contains(command, "test ") || strings.Contains(command, "\"") {
		// Use shell for commands with quotes or test
		cmd = exec.Command("sh", "-c", command)
		Debug(fmt.Sprintf("Using shell: sh -c '%s'", command))
	} else {
		// Parse command into parts for simple commands
		parts := strings.Fields(command)
		if len(parts) == 0 {
			return &CmdResult{
				ExitCode: 1,
				Success:  false,
				Error:    "empty command",
			}
		}
		cmd = exec.Command(parts[0], parts[1:]...)
		Debug(fmt.Sprintf("Command parts: %v", parts))
	}

	var output strings.Builder
	var errorOutput strings.Builder

	if flags.PrintOutput && flags.DecorateOutput {
		// If we need to decorate output, we need to capture and process it
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
	messageLength := 99
	format := fmt.Sprintf("%s %-*s %s",
		AddEmphasisGray(fmt.Sprintf("[%s]", GetEntrypointScript())),
		messageLength,
		message,
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
