package sys

import (
	"os"
	"strings"
)

// ExpandPath expands what a path in the settings writes for the machine: a
// leading ~ stands for the user's home directory, and %NAME%, $NAME and
// ${NAME} stand for the value of the variable NAME. A variable that is not
// set is left as written, so that a path is never silently emptied.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") && (len(path) == 1 || path[1] == '/' || path[1] == '\\') {
		if home, err := os.UserHomeDir(); err == nil {
			path = home + path[1:]
		}
	}
	return expandEnv(path)
}

// expandEnv is path with its variables expanded, and the ones it does not
// know kept as they are written.
func expandEnv(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			if name, n := varAtPercent(s[i:]); name != "" {
				if v, ok := os.LookupEnv(name); ok {
					b.WriteString(v)
					i += n
					continue
				}
			}
		case '$':
			if name, n := varAtDollar(s[i:]); name != "" {
				if v, ok := os.LookupEnv(name); ok {
					b.WriteString(v)
					i += n
					continue
				}
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// varAtPercent is the name in the %NAME% that starts s, and the length of the
// whole placeholder, when s starts with one.
func varAtPercent(s string) (string, int) {
	end := strings.IndexByte(s[1:], '%')
	if end < 0 {
		return "", 0
	}
	name := s[1 : 1+end]
	if !varName(name) {
		return "", 0
	}
	return name, end + 2
}

// varAtDollar is the name in the $NAME or ${NAME} that starts s, and the
// length of the whole placeholder, when s starts with one.
func varAtDollar(s string) (string, int) {
	if len(s) < 2 {
		return "", 0
	}
	if s[1] == '{' {
		end := strings.IndexByte(s[2:], '}')
		if end < 0 {
			return "", 0
		}
		name := s[2 : 2+end]
		if !varName(name) {
			return "", 0
		}
		return name, end + 3
	}
	n := 0
	for n+1 < len(s) && isVarByte(s[n+1]) {
		n++
	}
	if n == 0 {
		return "", 0
	}
	return s[1 : 1+n], n + 1
}

// varName reports whether s is the name of a variable.
func varName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isVarByte(s[i]) {
			return false
		}
	}
	return true
}

// isVarByte reports whether c may be a byte of a variable name.
func isVarByte(c byte) bool {
	return c == '_' ||
		'a' <= c && c <= 'z' ||
		'A' <= c && c <= 'Z' ||
		'0' <= c && c <= '9'
}
