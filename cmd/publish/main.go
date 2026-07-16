// Command publish uploads a packaged release to the Factorio Mod Portal.
package main

import (
	"fmt"
	"log/slog"
	"os"

	gopher "github.com/siutsin/factorio-gopher/internal"
)

var exitProcess = os.Exit

func main() {
	exitProcess(run(os.Args[1:], gopher.Publish))
}

func run(args []string, publish func([]string) (string, int, error)) int {
	message, exitCode, err := publish(args)
	if err != nil {
		slog.Error("publish failed", "err", err)
		return exitCode
	}
	fmt.Println(message)
	return 0
}
