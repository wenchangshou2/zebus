package utils

import (
	"os"
	"path/filepath"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
func GetFullPath(path string) (string, error) {
	fullExecPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	dir, _ := filepath.Split(fullExecPath)
	return filepath.Join(dir, path), nil
}
