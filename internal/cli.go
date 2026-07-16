// CLI plumbing: flag parsing, subcommand dispatch, and default-path
// resolution for the build binary. Lives in this package so cmd/main.go
// stays a one-line entrypoint and the dispatcher is testable in isolation.

package gopher

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

// CLI parses args, dispatches the requested subcommand, and returns an exit
// code. All progress and errors are emitted via slog (default text handler
// to stderr); stderr is also used for flag parse errors and usage output,
// matching the flag package's convention.
func CLI(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(stderr)
	gfxDir := fs.String("gfx", DefaultGfxDir(), "directory containing gopher-*.png sources")
	modDir := fs.String("mod", DefaultModDir(), "mod source directory containing info.json")
	modsDir := fs.String("mods", "", "Factorio mods directory (defaults to OS-standard location)")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "usage: build [-gfx <dir>] [-mod <dir>] [-mods <dir>] {running|shadow|sheets|knight|all|install|uninstall}") //nolint:errcheck // best-effort write to stderr
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}

	if err := dispatch(fs.Arg(0), *gfxDir, *modDir, *modsDir); err != nil {
		slog.Error("build failed", "command", fs.Arg(0), "err", err)
		return 1
	}
	return 0
}

// dispatch routes the user-supplied subcommand to the matching build or
// install step; "all" runs the build steps in dependency order and stops at
// the first failure.
func dispatch(cmd, gfxDir, modDir, modsDir string) error {
	switch cmd {
	case "running":
		return Run(gfxDir)
	case "shadow":
		return Shadow(gfxDir)
	case "sheets":
		return Sheets(gfxDir)
	case "knight":
		return Knight(gfxDir)
	case "all":
		if err := Run(gfxDir); err != nil {
			return err
		}
		if err := Shadow(gfxDir); err != nil {
			return err
		}
		if err := Sheets(gfxDir); err != nil {
			return err
		}
		return Knight(gfxDir)
	case "install":
		dir, err := resolveModsDir(modsDir)
		if err != nil {
			return err
		}
		return Install(modDir, dir)
	case "uninstall":
		dir, err := resolveModsDir(modsDir)
		if err != nil {
			return err
		}
		return Uninstall(modDir, dir)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

// defaultModsDirFn is the function used by resolveModsDir to compute the
// default Factorio mods directory. Indirected through a var so tests can
// inject a failure to exercise the error branch. Same concurrency contract
// as the fs hooks in install.go: tests that swap it must not run in
// parallel with code that reads the default.
var defaultModsDirFn = DefaultModsDir

// resolveModsDir returns the explicit -mods flag value when set, otherwise
// the OS-standard default from DefaultModsDir.
func resolveModsDir(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	return defaultModsDirFn()
}

// DefaultGfxDir resolves <repo>/mod/graphics relative to this source file.
// Intended for `go run ./cmd` from a clone of the repo; users running a
// built binary should pass `-gfx` explicitly.
func DefaultGfxDir() string {
	return filepath.Join(repoRoot(), "mod", "graphics")
}

// DefaultModDir resolves <repo>/mod relative to this source file. Intended
// for `go run ./cmd` from a clone of the repo; users running a built binary
// should pass `-mod` explicitly.
func DefaultModDir() string {
	return filepath.Join(repoRoot(), "mod")
}

// callerFile returns the source path of the current frame. Indirected
// through a var so tests can inject a non-existent path to exercise the
// repoRoot fallback branch.
var callerFile = func() string {
	_, file, _, _ := runtime.Caller(1)
	return file
}

// repoRoot resolves the repository root from this file's compile-time path.
// Returns "." when the embedded source path doesn't resolve to an existing
// directory (e.g. a binary built with -trimpath, or relocated to another
// machine). Production callers are expected to pass -gfx / -mod explicitly
// in those cases; the fallback keeps the resulting path obviously invalid
// rather than silently pointing at a stale build-host location.
func repoRoot() string {
	root := filepath.Clean(filepath.Join(filepath.Dir(callerFile()), ".."))
	if _, err := os.Stat(root); err != nil {
		return "."
	}
	return root
}
