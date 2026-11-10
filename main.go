// Command wtgo is a terminal emulator: a window of tabs, each a shell.
package main

import (
	"log"

	"wtgo/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
