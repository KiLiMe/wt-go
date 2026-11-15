package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// conptyManifest is Scripts/conpty.json: the NuGet package the ConPTY beside
// the executable comes from, and the bytes each file is meant to be.
type conptyManifest struct {
	Package     string            `json:"package"`
	Version     string            `json:"version"`
	NupkgSHA256 string            `json:"nupkg_sha256"`
	Files       map[string]string `json:"files"`
}

// TestConPTYFilesMatchTheManifest pins the ConPTY the app ships beside itself.
// The files are Microsoft's, and a build that swapped one would still start,
// only in a console nobody tested: the manifest is what a later version is
// checked against, by downloading its package and hashing it again.
func TestConPTYFilesMatchTheManifest(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("Scripts", "conpty.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m conptyManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Scripts/conpty.json: %v", err)
	}
	if m.Package == "" || m.Version == "" || len(m.NupkgSHA256) != 64 {
		t.Fatalf("Scripts/conpty.json names no package, version and package hash")
	}
	if len(m.Files) == 0 {
		t.Fatal("Scripts/conpty.json lists no file")
	}
	for name, want := range m.Files {
		got, err := os.ReadFile(filepath.FromSlash(name))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		sum := sha256.Sum256(got)
		if have := hex.EncodeToString(sum[:]); have != want {
			t.Errorf("%s holds %s, want %s", name, have, want)
		}
	}
}
