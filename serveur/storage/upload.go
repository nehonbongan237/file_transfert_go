package storage

import (
	"io"
	"os"
	"path/filepath"
)

func SaveFile(group, filename string, reader io.Reader) error {
	dir := filepath.Join("storage", group)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}
