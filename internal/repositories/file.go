package repositories

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileRepository struct {
	dataDir string
}

func NewFileRepository(dataDir string) (*FileRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dataDir, err)
	}

	return &FileRepository{dataDir: dataDir}, nil
}

func (r *FileRepository) pathForID(id string) string {
	prefix := id[:2]
	dir := filepath.Join(r.dataDir, prefix)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}

	return filepath.Join(dir, id+".bin")
}

func (r *FileRepository) Save(id string, src io.Reader) error {
	dstPath := r.pathForID(id)
	if dstPath == "" {
		return os.ErrInvalid
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", dstPath, err)
	}
	defer func() { _ = dst.Close() }()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("copy to file %s: %w", dstPath, err)
	}

	return nil
}

func (r *FileRepository) Get(id string) (*os.File, error) {
	path := r.pathForID(id)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", path, err)
	}

	return file, nil
}

func (r *FileRepository) Delete(id string) error {
	path := r.pathForID(id)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove file %s: %w", path, err)
	}

	return nil
}

func (r *FileRepository) DataDir() string {
	return r.dataDir
}
