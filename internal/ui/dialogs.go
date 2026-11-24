package ui

// Boundary: the models of the three dialogs, which the window opens one at
// a time. Drawing is rename.go, about.go and confirm.go.

// Rename is the dialog that renames a tab.
type Rename struct {
	// Open is whether the dialog shows; the dialog closes it.
	Open bool
	// Tab is the tab the dialog renames.
	Tab int
	// Text is what the field holds.
	Text string
}

// About is the dialog that says what the program is: the version it runs, what
// it was built from, what it runs on and the files it leaves, and what it is
// licensed under.
type About struct {
	// Open is whether the dialog shows; the dialog closes it.
	Open bool
	// Version is the build as `wt-go version` says it, such as
	// "v0.3.1-dev (packaged 0.3.0)", and Revision the commit it was built
	// from, "" in a build with no history.
	Version  string
	Revision string
	// Go and Platform are what it runs on.
	Go       string
	Platform string
	// Mygo is the mygo version the window and the terminals are built on:
	// which build of the framework a bug report is about matters as much as
	// which build of the program.
	Mygo string
	// Portable is whether this copy keeps its files in a data directory beside
	// the executable rather than under the user's application data: it decides
	// where the two paths below point, and a person pasting a report needs it
	// to read them.
	Portable bool
	// Terminal is where the terminals' console comes from — TerminalApp,
	// TerminalEnv or TerminalInherent — and TerminalPath the dll it is, ""
	// for the console the machine comes with. mygo picks a ConPTY the first
	// time it makes a terminal, so it is read once at startup: a later read
	// would name a dll the terminals in front were not made in.
	Terminal     string
	TerminalPath string
	// Settings and Log are where settings.json and wt-go.log are, the first
	// thing `wt-go diag` is asked about: a person sent to look at a file by a
	// bug report should not have to open a console to find it.
	Settings string
	Log      string
	// Project is the address of the program's page, which the dialog opens.
	Project string
	// Note is what went wrong the last time the dialog opened the page, which
	// it says in place of the silent button a desktop without a browser
	// leaves.
	Note string
}

// Confirm is the question the window asks before it does something it cannot
// take back, as closing the settings with changes it has not saved.
type Confirm struct {
	// Open is false once the dialog is dismissed, which the window takes
	// as no.
	Open  bool
	Title string
	Text  string
	Yes   string
	No    string
}

// Where the terminals' console comes from, which the about dialog says.
const (
	// TerminalApp is the ConPTY shipped beside the program, in its conpty
	// folder, which is what a build unpacks to.
	TerminalApp = "app"
	// TerminalEnv is the ConPTY that MYGO_CONPTY names, which is where a person
	// points the window at a build of their own.
	TerminalEnv = "env"
	// TerminalInherent is the console the machine comes with: Windows' own, or
	// the pty a system gives every program.
	TerminalInherent = "inherent"
)
