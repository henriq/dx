package ports

import "io"

// CommandRunner executes shell commands and returns their output.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
	RunWithEnv(name string, env []string, args ...string) ([]byte, error)
	RunInDir(dir, name string, args ...string) ([]byte, error)
	RunWithEnvInDir(dir string, env []string, name string, args ...string) ([]byte, error)
	RunWithStdin(stdin io.Reader, name string, args ...string) ([]byte, error)
	// RunInteractive executes a command with stdin, stdout, and stderr connected
	// to the terminal for interactive use. Returns error if command fails.
	RunInteractive(name string, args ...string) error
}
