package ports

type Scm interface {
	Download(repositoryUrl string, branch string, repositoryPath string) error
}
