package scm

import (
	"fmt"
	"path/filepath"
	"strings"

	"dx/internal/ports"
)

type GitClient struct {
	commandRunner ports.CommandRunner
	fileSystem    ports.FileSystem
}

func ProvideGitClient(commandRunner ports.CommandRunner, fileSystem ports.FileSystem) *GitClient {
	return &GitClient{
		commandRunner: commandRunner,
		fileSystem:    fileSystem,
	}
}

func (g *GitClient) ContainsRepository(repositoryPath string) bool {
	exists, err := g.fileSystem.FileExists(filepath.Join(repositoryPath, ".git", "HEAD"))
	return err == nil && exists
}

func (g *GitClient) UpdateOriginUrl(repositoryPath string, originUrl string) error {
	output, err := g.commandRunner.RunInDir(repositoryPath, "git", "remote", "set-url", "origin", originUrl)
	if err != nil {
		return fmt.Errorf("failed to update git remote URL: %v\n%s", err, string(output))
	}

	return nil
}

func (g *GitClient) FetchRefFromOrigin(repositoryPath string, branch string) error {
	output, err := g.commandRunner.RunInDir(repositoryPath, "git", "fetch", "origin", "-f", branch)
	if err != nil {
		return fmt.Errorf("failed to fetch from remote: %v\n%s", err, string(output))
	}

	return nil
}

func (g *GitClient) GetCurrentRef(repositoryPath string) (string, error) {
	output, err := g.commandRunner.RunInDir(repositoryPath, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *GitClient) Checkout(repositoryPath string, commit string) error {
	output, err := g.commandRunner.RunInDir(repositoryPath, "git", "checkout", commit)
	if err != nil {
		return fmt.Errorf("failed to checkout %s: %v\n%s", commit, err, string(output))
	}

	return nil
}

func (g *GitClient) IsBranch(repositoryPath string, branch string) bool {
	_, err := g.commandRunner.RunInDir(repositoryPath, "git", "rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+branch)
	return err == nil
}

func (g *GitClient) GetRevisionForCommit(repositoryPath string, commit string) (string, error) {
	output, err := g.commandRunner.RunInDir(repositoryPath, "git", "rev-parse", commit)
	if err != nil {
		return "", fmt.Errorf("failed to get origin revision: %v\n%s", err, string(output))
	}

	return string(output), nil
}

func (g *GitClient) ResetToCommit(repositoryPath string, commit string) error {
	output, err := g.commandRunner.RunInDir(repositoryPath, "git", "reset", "--hard", commit)
	if err != nil {
		return fmt.Errorf("failed to reset to %s: %v\n%s", commit, err, string(output))
	}

	return nil
}

func (g *GitClient) Download(repositoryPath string, branch string, repositoryUrl string) error {
	output, err := g.commandRunner.Run("git", "clone", repositoryUrl, "--branch", branch, repositoryPath)
	if err != nil {
		return fmt.Errorf("failed to clone %s: %v\n%s", repositoryUrl, err, string(output))
	}

	return nil
}
