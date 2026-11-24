package sys

// Boundary: settings.json -- where it lives, and the three things the window
// does with it. A file of the settings directory is configfile.go's to read and
// write.

// settingsName is the file a person edits to change what the window does.
const settingsName = "settings.json"

// SettingsPath is the file the settings live in: settings.json in ConfigDir().
func SettingsPath() (string, error) { return configFile(settingsName) }

// ReadSettings is settings.json, and whether there was one. A window that
// finds none starts with the defaults and writes them, so the file is there
// to edit.
func ReadSettings() ([]byte, bool, error) { return readConfigFile(settingsName) }

// WriteSettings makes the settings directory and writes settings.json.
func WriteSettings(b []byte) error { return writeConfigFile(settingsName, b) }
