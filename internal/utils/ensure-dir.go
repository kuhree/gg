package utils

import (
	"os"
	"path/filepath"
)

func EnsureDir(filename string) error {
	dir := filepath.Dir(filename)
	return os.MkdirAll(dir, 0755)
}
