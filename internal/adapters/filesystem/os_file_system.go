package filesystem

import (
	"dx/internal/ports"
	"fmt"
	"os"
	"path/filepath"
)

type OsFileSystem struct{}

func ProvideOsFileSystem() *OsFileSystem {
	return &OsFileSystem{}
}

func (f *OsFileSystem) ReadFile(path string) ([]byte, error) {
	if len(path) > 0 && path[:1] == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	return os.ReadFile(path)
}

func (f *OsFileSystem) WriteFile(path string, content []byte, accessMode ports.AccessMode) error {
	if len(path) > 0 && path[:1] == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	err := f.EnsureDirExists(path)
	if err != nil {
		return fmt.Errorf("failed to ensure directory exists: %w", err)
	}

	if err := os.WriteFile(path, content, getOsFileModeForAccessMode(accessMode)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (f *OsFileSystem) EnsureDirExists(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), getOsFileModeForAccessMode(ports.ReadWriteExecute)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

func (f *OsFileSystem) FileExists(path string) (bool, error) {
	if len(path) > 0 && path[:1] == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false, fmt.Errorf("failed to get user home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check if file exists: %w", err)
}

func getOsFileModeForAccessMode(accessMode ports.AccessMode) os.FileMode {
	switch accessMode {
	case ports.ReadWrite:
		return 0600
	case ports.ReadWriteExecute:
		return 0700
	case ports.ReadAllWriteOwner:
		return 0644
	default:
		return 0600
	}
}
