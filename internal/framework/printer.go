package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Color constants for ANSI escape codes
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Gray    = "\033[30;1m"
)

// Visual indicators
const (
	CheckMark = "\u2713" // ✓
	CrossMark = "\u2717" // ✗
)

// Color formatting functions
func AddEmphasisBlue(text string) string {
	return Blue + text + Reset
}

func AddEmphasisRed(text string) string {
	return Red + text + Reset
}

func AddEmphasisGreen(text string) string {
	return Green + text + Reset
}

func AddEmphasisMagenta(text string) string {
	return Magenta + text + Reset
}

func AddEmphasisGray(text string) string {
	return Gray + text + Reset
}

// GetEntrypointScript returns the name of the main executable
func GetEntrypointScript() string {
	if len(os.Args) > 0 {
		return filepath.Base(os.Args[0])
	}
	return "tf"
}

// Info prints an info message with consistent formatting
func Info(message string) {
	format := AddEmphasisGray(fmt.Sprintf("[%s]", GetEntrypointScript())) + " %s\n"
	fmt.Fprintf(os.Stderr, format, message)
}

// Error prints an error message with consistent formatting
func Error(message string) {
	format := AddEmphasisRed(fmt.Sprintf("[%s]", GetEntrypointScript())) + " %s\n"
	fmt.Fprintf(os.Stderr, format, message)
}

// Debug prints a debug message (only if debug is enabled)
func Debug(message string) {
	if os.Getenv("TFM_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", message)
	}
}

// stripAnsiCodes removes ANSI escape sequences from a string
func stripAnsiCodes(str string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(str, "")
}

// getVisualLength returns the visual length of a string (excluding ANSI codes)
func getVisualLength(str string) int {
	return len(stripAnsiCodes(str))
}
