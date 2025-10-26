package scm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitClient struct{}

func ProvideGitClient() *GitClient {
	return &GitClient{}
}

func (g *GitClient) ContainsRepository(repositoryPath string) bool {
	_, err := os.Stat(filepath.Join(repositoryPath, ".git", "HEAD"))
	return err == nil
}

func (g *GitClient) UpdateOriginUrl(repositoryPath string, originUrl string) error {
	output := strings.Builder{}
	setUrlCmd := exec.Command("git", "remote", "set-url", "origin", originUrl)
	setUrlCmd.Dir = repositoryPath
	setUrlCmd.Stdout = &output
	setUrlCmd.Stderr = &output
	if err := setUrlCmd.Run(); err != nil {
		return fmt.Errorf("failed to update git remote URL: %v\n%s", err, output.String())
	}

	return nil
}

func (g *GitClient) FetchRefFromOrigin(repositoryPath string, branch string) error {
	output := strings.Builder{}
	fetchCmd := exec.Command("git", "fetch", "origin", "-f", branch)
	fetchCmd.Dir = repositoryPath
	fetchCmd.Stdout = &output
	fetchCmd.Stderr = &output
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch from remote: %v\n%s", err, output.String())
	}

	return nil
}

func (g *GitClient) GetCurrentRef(repositoryPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repositoryPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output[:len(output)-1]), nil // trim newline
}

func (g *GitClient) Checkout(repositoryPath string, commit string) error {
	output := strings.Builder{}
	checkoutCmd := exec.Command("git", "checkout", commit)
	checkoutCmd.Dir = repositoryPath
	checkoutCmd.Stdout = &output
	checkoutCmd.Stderr = &output
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout %s: %v\n%s", commit, err, output.String())
	}

	return nil
}

func (g *GitClient) IsBranch(repositoryPath string, branch string) bool {
	cmd := exec.Command(
		"git",
		"rev-parse",
		"--verify",
		"--quiet",
		"refs/remotes/origin/"+branch,
	)

	output := strings.Builder{}
	cmd.Dir = repositoryPath
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func (g *GitClient) GetRevisionForCommit(repositoryPath string, commit string) (string, error) {
	output := strings.Builder{}
	originRevCmd := exec.Command("git", "rev-parse", commit)
	originRevCmd.Dir = repositoryPath
	originRevCmd.Stderr = &output
	originRev, err := originRevCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get origin revision: %v\n%s", err, output.String())
	}

	return string(originRev), nil
}

func (g *GitClient) ResetToCommit(repositoryPath string, commit string) error {
	output := strings.Builder{}

	resetCmd := exec.Command("git", "reset", "--hard", commit)
	resetCmd.Dir = repositoryPath
	resetCmd.Stdout = &output
	resetCmd.Stderr = &output

	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to %s: %v\n%s", commit, err, output.String())
	}

	return nil
}

func (g *GitClient) Download(repositoryPath string, branch string, repositoryUrl string) error {
	cloneCmd := exec.Command("git", "clone", repositoryUrl, "--branch", branch, repositoryPath)

	output := strings.Builder{}
	cloneCmd.Stdout = &output
	cloneCmd.Stderr = &output

	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone %s: %v\n%s", repositoryUrl, err, output.String())
	}

	return nil
}
