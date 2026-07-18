// Command gopher regenerates the procedural gopher sprite sheets and
// installs/uninstalls the mod into Factorio's mods directory.
//
// Usage: gopher {running|shadow|sheets|all|install|uninstall}
package main

import (
	"io"
	"os"

	gopher "github.com/siutsin/factorio-gopher/internal"
)

var (
	exitProcess             = os.Exit
	standardError io.Writer = os.Stderr
)

func main() {
	exitProcess(gopher.CLI(os.Args[1:], standardError))
}
