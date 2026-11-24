package cli

import (
	"regexp"
	"runtime/debug"
)

// Boundary: which build this is -- the version mygo stamps in, the one the
// toolchain recorded, and the pseudo-version it is tidied to.

// packagedVersion is the version mygo.json announces, passed in by the build
// scripts with -ldflags -X. It is not the same number as the one the toolchain
// stamps into the binary: that one says which build answered, this one says
// which release it is meant to be.
var packagedVersion string

// version is what `wt-go version` and the report call this build: which one
// answered, and, in a build a script stamped, the release it was packaged as.
func (r Report) version() string { return versionLabel(r.Version, r.Packaged) }

// buildInfo is the version the toolchain stamped into the exe and the commit
// it was built from: a report says which build answered.
func buildInfo() (version, revision string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown", ""
	}
	version = info.Main.Version
	if version == "" || version == "(devel)" {
		version = "devel"
	}
	version = tidyVersion(version)
	modified := false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if modified && revision != "" {
		revision += "+dirty"
	}
	return version, revision
}

// pseudo is the version `go build` stamps into a build made from a commit
// that no tag names: v0.1.1-0.20261110153946-e9f374cbc6b0. A report keeps the
// part that says something and drops the rest, the commit being printed
// beside it already.
var pseudo = regexp.MustCompile(`^(v[0-9][^-]*)-(0\.[0-9]{14}|[0-9]{14})-[0-9a-f]+`)

// tidyVersion turns a pseudo-version into v0.1.1-dev.
func tidyVersion(v string) string {
	if m := pseudo.FindStringSubmatch(v); m != nil {
		return m[1] + "-dev"
	}
	return v
}
