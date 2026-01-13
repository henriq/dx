package command_runner

import (
	"io"
	"os"
	"os/exec"

	"dx/internal/ports"
)

// OsCommandRunner executes shell commands using os/exec.
type OsCommandRunner struct{}

func ProvideOsCommandRunner() *OsCommandRunner {
	return &OsCommandRunner{}
}

func (r *OsCommandRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

func (r *OsCommandRunner) RunWithEnv(name string, env []string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...) // Extend environment instead of replacing
	return cmd.CombinedOutput()
}

func (r *OsCommandRunner) RunInDir(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func (r *OsCommandRunner) RunWithStdin(stdin io.Reader, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	return cmd.CombinedOutput()
}

func (r *OsCommandRunner) RunInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var _ ports.CommandRunner = (*OsCommandRunner)(nil)
