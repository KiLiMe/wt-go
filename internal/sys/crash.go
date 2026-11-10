package sys

import (
	"fmt"
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
	path := filepath.Join(dir, "crash-"+now.Format("20060102-150405")+".log")
	body := fmt.Sprintf("%s\n%s\n\n%s\n", now.Format(time.RFC3339), reason, stack)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", fmt.Errorf("sys: write %s: %w", path, err)
	}
	return path, nil
}
