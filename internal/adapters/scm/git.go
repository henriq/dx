package scm

import (
	"fmt"
	"os"
)

type Git struct {
	gitClient *GitClient
	// Track unique repo+branch combinations to avoid duplicate clones
	cloned map[string]bool
}

func ProvideGit(gitClient *GitClient) *Git {
	return &Git{
		gitClient: gitClient,
		cloned:    make(map[string]bool),
	}
}

func (g *Git) Download(repositoryUrl string, ref string, repositoryPath string) error {
	repoKey := repositoryPath + ":" + ref
	if !g.cloned[repoKey] {
		if g.gitClient.ContainsRepository(repositoryPath) {
			err := g.gitClient.UpdateOriginUrl(repositoryPath, repositoryUrl)
			if err != nil {
				return err
			}

			err = g.gitClient.FetchRefFromOrigin(repositoryPath, ref)
			if err != nil {
				return err
			}

			currentRef, err := g.gitClient.GetCurrentRef(repositoryPath)
			if err != nil {
				return err
			}

			if currentRef != ref {
				err = g.gitClient.Checkout(repositoryPath, ref)
				if err != nil {
					return err
				}
			}

			if g.gitClient.IsBranch(repositoryPath, ref) {
				localRevision, err := g.gitClient.GetRevisionForCommit(repositoryPath, ref)
				if err != nil {
					return err
				}

				originRevision, err := g.gitClient.GetRevisionForCommit(
					repositoryPath,
					fmt.Sprintf("origin/%s", ref),
				)
				if err != nil {
					return err
				}

				if originRevision != localRevision {
					err = g.gitClient.ResetToCommit(repositoryPath, fmt.Sprintf("origin/%s", ref))
					if err != nil {
						return err
					}
				}
			}
		} else {
			// Create the destination directory if it doesn't exist
			if err := os.MkdirAll(repositoryPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create destination directory: %v", err)
			}
			err := g.gitClient.Download(repositoryPath, ref, repositoryUrl)
			if err != nil {
				return err
			}
		}
	}

	g.cloned[repoKey] = true
	return nil
}
