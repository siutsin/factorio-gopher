// Command build regenerates the procedural gopher sprite sheets and
// installs/uninstalls the mod into Factorio's mods directory.
//
// Usage: build {running|shadow|sheets|all|install|uninstall}
package main

import (
	"os"

	gopher "github.com/siutsin/factorio-gopher/internal"
)

func main() {
	os.Exit(gopher.CLI(os.Args[1:], os.Stderr))
}
