package ports

type AccessMode int

const (
	ReadWrite = iota
	ReadWriteExecute
	ReadAllWriteOwner
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte, accessMode AccessMode) error
	EnsureDirExists(path string) error
	FileExists(path string) (bool, error)
}
