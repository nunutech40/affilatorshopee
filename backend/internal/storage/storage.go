package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Storage interface {
	Save(productID, filename string, reader io.Reader, maxSize int64) (string, int64, error)
	Open(relativePath string) (*os.File, error)
	Delete(relativePath string) error
}

type LocalStorage struct{ root string }

func NewLocalStorage(root string) (*LocalStorage, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("storage path is required")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &LocalStorage{root: root}, nil
}

func (s *LocalStorage) Save(productID, filename string, reader io.Reader, maxSize int64) (string, int64, error) {
	if filepath.Base(filename) != filename || strings.Contains(filename, "..") {
		return "", 0, fmt.Errorf("invalid filename")
	}
	relative := filepath.Join("products", productID, filename)
	absolute := filepath.Join(s.root, relative)
	if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
		return "", 0, err
	}
	temporary, err := os.CreateTemp(filepath.Dir(absolute), ".upload-*")
	if err != nil {
		return "", 0, err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	limited := io.LimitReader(reader, maxSize+1)
	size, err := io.Copy(temporary, limited)
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", 0, err
	}
	if size > maxSize {
		return "", 0, fmt.Errorf("file exceeds maximum size")
	}
	if err := os.Chmod(temporaryName, 0o640); err != nil {
		return "", 0, err
	}
	if err := os.Rename(temporaryName, absolute); err != nil {
		return "", 0, err
	}
	return filepath.ToSlash(relative), size, nil
}

func (s *LocalStorage) Open(relativePath string) (*os.File, error) {
	clean := filepath.Clean(relativePath)
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) || clean == ".." {
		return nil, fmt.Errorf("invalid storage path")
	}
	return os.Open(filepath.Join(s.root, clean))
}

func (s *LocalStorage) Delete(relativePath string) error {
	clean := filepath.Clean(relativePath)
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) || clean == ".." {
		return fmt.Errorf("invalid storage path")
	}
	return os.Remove(filepath.Join(s.root, clean))
}
