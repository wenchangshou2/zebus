package utils

import (
	"os"
	"path/filepath"
)

func IsExist(path string)bool{
	_,err:=os.Stat(path)
	return err==nil||os.IsExist(err)
}
func GetFullPath(path string) (string, error) {
	fullexecpath, err := os.Executable()
	if err != nil {
		return "", err
	}

	dir, _ := filepath.Split(fullexecpath)
	return filepath.Join(dir, path), nil
}