package output

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ColorsEnabled returns true if terminal colors should be used.
// Respects NO_COLOR environment variable (https://no-color.org/)
func ColorsEnabled() bool {
	_, noColor := os.LookupEnv("NO_COLOR")
	if noColor {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// ANSI color codes
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

// Symbols for CLI output (ASCII-compatible)
const (
	SymbolSuccess = "+"
	SymbolError   = "x"
	SymbolWarning = "!"
	SymbolInfo    = "*"
	SymbolArrow   = "->"
	SymbolBullet  = "-"
)

// Bold returns text in bold (or plain if colors disabled)
func Bold(text string) string {
	if !ColorsEnabled() {
		return text
	}
	return fmt.Sprintf("%s%s%s", bold, text, reset)
}

// Dim returns text in dim style (or plain if colors disabled)
func Dim(text string) string {
	if !ColorsEnabled() {
		return text
	}
	return fmt.Sprintf("%s%s%s", dim, text, reset)
}

// Success returns text styled for success messages
func Success(text string) string {
	if !ColorsEnabled() {
		return text
	}
	return fmt.Sprintf("%s%s%s", green, text, reset)
}

// Error returns text styled for error messages
func Error(text string) string {
	if !ColorsEnabled() {
		return text
	}
	return fmt.Sprintf("%s%s%s", red, text, reset)
}

// Warning returns text styled for warning messages
func Warning(text string) string {
	if !ColorsEnabled() {
		return text
	}
	return fmt.Sprintf("%s%s%s", yellow, text, reset)
}

// Info returns text styled for informational messages
func Info(text string) string {
	if !ColorsEnabled() {
		return text
	}
	return fmt.Sprintf("%s%s%s", cyan, text, reset)
}

// Header returns text styled as a section header
func Header(text string) string {
	if !ColorsEnabled() {
		return text
	}
	return fmt.Sprintf("%s%s%s%s", bold, white, text, reset)
}

// Secondary returns text in dim cyan for secondary information
func Secondary(text string) string {
	if !ColorsEnabled() {
		return text
	}
	return fmt.Sprintf("%s%s%s%s", dim, cyan, text, reset)
}

// PrintHeader prints a bold section header
func PrintHeader(text string) {
	fmt.Println(Header(text))
}

// PrintSuccess prints a success message with checkmark
func PrintSuccess(message string) {
	fmt.Printf("%s %s\n", SymbolSuccess, Success(message))
}

// PrintError prints an error message with X symbol
func PrintError(message string) {
	fmt.Printf("%s %s\n", SymbolError, Error(message))
}

// PrintWarning prints a warning message with ! symbol
func PrintWarning(message string) {
	fmt.Printf("%s %s\n", SymbolWarning, Warning(message))
}

// PrintInfo prints an info message with * symbol
func PrintInfo(message string) {
	fmt.Printf("%s %s\n", SymbolInfo, Info(message))
}

// PrintStep prints a step being executed with arrow
func PrintStep(message string) {
	fmt.Printf("  %s %s\n", SymbolArrow, message)
}

// PrintSecondary prints secondary/supplementary information
func PrintSecondary(message string) {
	fmt.Printf("  %s %s\n", SymbolArrow, Secondary(message))
}
