package ports

// TerminalInput provides methods for reading user input from the terminal.
type TerminalInput interface {
	// ReadPassword prompts for a password and returns the input without echoing to the terminal.
	ReadPassword(prompt string) (string, error)
	// IsTerminal returns true if stdin is connected to a terminal.
	IsTerminal() bool
}
