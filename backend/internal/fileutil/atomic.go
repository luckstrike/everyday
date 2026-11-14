// Package fileutil holds small file helpers shared by the packages that keep their
// data in files.
package fileutil

import (
	"os"
	"path/filepath"
)

// WriteAtomic writes data to path so that a crash, or a reader at the wrong moment,
// never sees a half-written file: it writes a temporary file in the same directory,
// gives it perm, and renames it over the target. Missing parent directories are
// created readable by the owner only.
func WriteAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".write-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
