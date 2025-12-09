package framework

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
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

// parseCommand parses a command string into program and arguments
// This handles basic shell-like parsing including quoted strings
func parseCommand(cmdStr string) (string, []string) {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return "", nil
	}

	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(cmdStr); i++ {
		char := cmdStr[i]

		if !inQuotes && (char == '"' || char == '\'') {
			inQuotes = true
			quoteChar = char
		} else if inQuotes && char == quoteChar {
			inQuotes = false
			quoteChar = 0
		} else if !inQuotes && char == ' ' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	if len(parts) == 0 {
		return "", nil
	}

	return parts[0], parts[1:]
}

// execSystemCommand executes the actual system command directly without a shell
func execSystemCommand(command string, flags *CmdFlags) *CmdResult {
	// Debug the command being executed
	Debug(fmt.Sprintf("Executing command: %s", command))

	// Check for empty command
	if strings.TrimSpace(command) == "" {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Error:    "empty command",
		}
	}

	// Parse the command into program and arguments
	program, args := parseCommand(command)
	Debug(fmt.Sprintf("Parsed command: %s %v", program, args))

	if program == "" {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Error:    "empty command",
		}
	}

	cmd := exec.Command(program, args...)
	return execCommand(cmd, flags)
}

// outputLine represents a line of output with its formatting context
type outputLine struct {
	text     string
	isStderr bool
	decorate bool
}

// execCommand is the common execution function for both direct and shell commands
func execCommand(cmd *exec.Cmd, flags *CmdFlags) *CmdResult {
	var output strings.Builder
	var errorOutput strings.Builder

	// For interactive commands, connect pipes differently to handle unbuffered output
	isInteractive := !flags.DecorateOutput

	if isInteractive {
		// Interactive mode: pass through stdout/stderr directly and capture in background
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Start the command
		if err := cmd.Start(); err != nil {
			return &CmdResult{
				ExitCode: 1,
				Success:  false,
				Error:    err.Error(),
			}
		}

		// Wait for the command to complete
		err := cmd.Wait()

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
			Output:   "", // No output captured in interactive mode
			Error:    "",
		}
	}

	// Non-interactive mode: capture stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Error:    err.Error(),
		}
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Error:    err.Error(),
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return &CmdResult{
			ExitCode: 1,
			Success:  false,
			Error:    err.Error(),
		}
	}

	// Use WaitGroups to coordinate goroutines
	var pumpWg sync.WaitGroup  // For stdout/stderr pump goroutines
	var printWg sync.WaitGroup // For output printer goroutine

	// Create buffered channel for async output handling
	// Buffer size prevents blocking when output is faster than printing
	outputChan := make(chan outputLine, 100)

	// Start output printer goroutine (decouples reading from writing)
	printWg.Add(1)
	go func() {
		defer printWg.Done()
		for line := range outputChan {
			if line.decorate {
				if line.isStderr {
					decoratedLine := AddEmphasisRed(fmt.Sprintf("[%s]", "err")) + " " + line.text
					fmt.Println(decoratedLine)
				} else {
					decoratedLine := AddEmphasisBlue(fmt.Sprintf("[%s]", "cmd")) + " " + line.text
					fmt.Println(decoratedLine)
				}
			} else {
				if line.isStderr {
					fmt.Fprintln(os.Stderr, line.text)
				} else {
					fmt.Println(line.text)
				}
			}
		}
	}()

	// Process stdout - pump lines into channel
	pumpWg.Add(1)
	go func() {
		defer pumpWg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			if flags.PrintOutput {
				outputChan <- outputLine{
					text:     line,
					isStderr: false,
					decorate: flags.DecorateOutput,
				}
			}
		}
	}()

	// Process stderr - pump lines into channel
	pumpWg.Add(1)
	go func() {
		defer pumpWg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			errorOutput.WriteString(line + "\n")
			if flags.PrintOutput {
				outputChan <- outputLine{
					text:     line,
					isStderr: true,
					decorate: flags.DecorateOutput,
				}
			}
		}
	}()

	// Wait for the command to complete
	err = cmd.Wait()

	// Wait for all pump goroutines to finish reading
	pumpWg.Wait()

	// Close channel after all pumps are done
	close(outputChan)

	// Wait for the printer goroutine to finish printing
	printWg.Wait()

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
