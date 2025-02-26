package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type Storage struct {
	dir string
}

func NewStorage(dir string) *Storage {
	return &Storage{dir: dir}
}

func (s *Storage) CreateFile(filename string) (*os.File, error) {
	path := filepath.Join(s.dir, filename)
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	return file, nil
}

func (s *Storage) ReadFile(filename string) (*os.File, error) {
	path := filepath.Join(s.dir, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	return file, nil
}

func (s *Storage) ListFiles() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}
	return entries, nil
}
