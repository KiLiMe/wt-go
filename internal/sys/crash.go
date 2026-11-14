package sys

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// WriteCrashLog writes what stopped the program next to the settings file, and
// answers with the path it wrote. A window that dies leaves something to read
// instead of a window that is gone.
func WriteCrashLog(reason string, stack []byte) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("sys: create %s: %w", dir, err)
	}
	now := time.Now()
	base := "crash-" + now.Format("20060102-150405")
	body := fmt.Sprintf("%s\n%s\n\n%s\n", now.Format(time.RFC3339), reason, stack)
	// O_EXCL: two crashes in one second would otherwise write the same name,
	// and the second would overwrite the first instead of joining it. The
	// suffix keeps both, and the name still matches IsLogPath's crash-*.log.
	for i := 0; i < crashTries; i++ {
		path := filepath.Join(dir, base+crashSuffix(i)+".log")
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("sys: write %s: %w", path, err)
		}
		_, werr := f.WriteString(body)
		cerr := f.Close()
		if werr != nil {
			return "", fmt.Errorf("sys: write %s: %w", path, werr)
		}
		if cerr != nil {
			return "", fmt.Errorf("sys: write %s: %w", path, cerr)
		}
		return path, nil
	}
	return "", fmt.Errorf("sys: write %s: %w", filepath.Join(dir, base), fs.ErrExist)
}

// crashTries is how many names one second's crash may take before the log is
// given up on: eight crashes in one second is a loop, and the disk is fuller
// with each one.
const crashTries = 8

// crashSuffix is the number appended to a crash log's name, empty for the
// first so that a single crash keeps the name earlier versions wrote.
func crashSuffix(i int) string {
	if i == 0 {
		return ""
	}
	return fmt.Sprintf("-%d", i)
}
