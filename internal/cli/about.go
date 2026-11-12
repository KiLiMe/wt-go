package cli

import "runtime"

// Self is what the program says about itself without reading the machine: the
// version the toolchain stamped, the commit it was built from, the release that
// version is meant to be, and what it runs on. A Report carries the same lines
// beside what `wt-go diag` read from the disk; a window shows these.
type Self struct {
	Version  string `json:"version"`
	Revision string `json:"revision,omitempty"`
	Packaged string `json:"packaged,omitempty"`
	Go       string `json:"go"`
	Platform string `json:"platform"`
}

// About is the Self of the running program.
func About() Self {
	version, revision := buildInfo()
	return Self{
		Version:  version,
		Revision: revision,
		Packaged: packagedVersion,
		Go:       runtime.Version(),
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// Label is the version with the release it is meant to be, as `wt-go version`
// and the About dialog say it: "v0.3.1-dev (packaged 0.3.0)".
func (s Self) Label() string { return versionLabel(s.Version, s.Packaged) }

// versionLabel is Label for the two numbers on their own, so that a Report and
// a Self cannot say the same build two ways.
func versionLabel(version, packaged string) string {
	if packaged == "" {
		return version
	}
	return version + " (packaged " + packaged + ")"
}
