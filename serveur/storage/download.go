package storage

import (
	"os"
	"path/filepath"
)

func OpenFile(group, filename string) (*os.File, int64, error) {
	path := filepath.Join("storage", group, filename)

	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, err
	}

	return file, info.Size(), nil
}
