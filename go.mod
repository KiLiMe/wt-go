module wtgo

go 1.27.3

require (
	github.com/ebitengine/purego v0.11.1
	github.com/egoist/mygo v0.8.3
	golang.org/x/sys v0.48.0
)

require (
	github.com/ProtonMail/go-crypto v1.4.1 // indirect
	github.com/cloudflare/circl v1.6.2 // indirect
	github.com/go-text/typesetting v0.3.5 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	go.digitalxero.dev/rpm v0.3.4 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/image v0.46.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

tool github.com/egoist/mygo/cmd/mygo

// Our fork of mygo: a terminal behind a ConPTY writes no reply at all, and keys
// go as the plain forms while a program has CSI u on (ADR-72, ADR-57).
replace github.com/egoist/mygo => github.com/KiLiMe/mygo v0.9.1-0.20261114072300-2632a920569b
