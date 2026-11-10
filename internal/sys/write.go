package sys

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeFileAtomic writes data to path by writing it next to path and renaming
// it into place, so that a reader sees either the old file or the new one, not
// half of each. If path is a symbolic link, the file it points to is written
// and the link is left in place. An existing file keeps its permissions; a
// new one is created with perm.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	target := path
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		target = resolved
	}
	if fi, err := os.Stat(target); err == nil {
		perm = fi.Mode().Perm()
	}
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("sys: create %s: %w", dir, err)
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return fmt.Errorf("sys: create a temporary file in %s: %w", dir, err)
	}
	tmp := f.Name()
	renamed := false
	defer func() {
		f.Close()
		if !renamed {
			os.Remove(tmp)
		}
	}()
	if err := f.Chmod(perm); err != nil {
		return fmt.Errorf("sys: write %s: %w", path, err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("sys: write %s: %w", path, err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sys: write %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("sys: write %s: %w", path, err)
	}
	if err := os.Rename(tmp, target); err != nil {
		return fmt.Errorf("sys: replace %s: %w", target, err)
	}
	renamed = true
	return nil
}
