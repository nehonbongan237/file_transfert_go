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

func DeleteFile(filename string) error {
	return os.Remove("storage/" + filename)
}
func ListAllFiles() ([]string, error) {
    baseDir := "storage"
    var allFiles []string

    groups, err := os.ReadDir(baseDir)
    if err != nil {
        return nil, err
    }

    for _, g := range groups {
        if g.IsDir() {
            entries, _ := os.ReadDir(baseDir + "/" + g.Name())
            for _, f := range entries {
                if !f.IsDir() {
                    allFiles = append(allFiles, g.Name()+"/"+f.Name())
                }
            }
        }
    }
    return allFiles, nil
}

