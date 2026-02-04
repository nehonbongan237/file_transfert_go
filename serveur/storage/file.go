package storage

import (
	"os"
	"path/filepath"
)

func ListFiles(group string) ([]string, error) {
	dir := filepath.Join("storage", group)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	return files, nil
}
