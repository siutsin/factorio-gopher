// Mod install / uninstall: symlink the mod source dir into Factorio's
// per-OS mods folder. Cross-platform replacement for the previous bash
// scripts.

package gopher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Package-level filesystem hooks. Default to the real os calls; tests
// swap them out via withFSHook to exercise error branches that would
// otherwise require chmod tricks or platform-specific failure modes.
//
// Concurrency contract: these are package globals. A test that swaps a
// hook MUST NOT call t.Parallel(), and parallel tests MUST NOT depend
// on the default value while a swap is in flight. The same rule applies
// to defaultModsDirFn in cli.go and to SetFrameSize in testseam.go.
var (
	fsMkdirAll  = os.MkdirAll
	fsRemoveAll = os.RemoveAll
	fsSymlink   = os.Symlink
	fsLstat     = os.Lstat
	fsRemove    = os.Remove
	fsReadDir   = os.ReadDir
	fsReadlink  = os.Readlink
)

// modInfo holds the fields of mod/info.json that Install needs to derive the
// mods-folder entry name (`<name>_<version>`).
type modInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// DefaultModsDir returns Factorio's default mods directory for the host OS.
// Wraps modsDirFor so production code never has to care about runtime.GOOS.
func DefaultModsDir() (string, error) {
	return modsDirFor(runtime.GOOS, os.UserHomeDir, os.Getenv)
}

// modsDirFor resolves Factorio's mods directory for an explicit goos value.
// Takes the home and env lookups as functions so tests can inject without
// mutating process state.
func modsDirFor(goos string, homeFn func() (string, error), getenv func(string) string) (string, error) {
	switch goos {
	case "darwin":
		home, err := homeFn()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "factorio", "mods"), nil
	case "windows":
		appdata := getenv("APPDATA")
		if appdata == "" {
			return "", errors.New("APPDATA not set")
		}
		return filepath.Join(appdata, "Factorio", "mods"), nil
	default:
		home, err := homeFn()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".factorio", "mods"), nil
	}
}

// Install symlinks modSrc into modsDir as `<name>_<version>`, after first
// removing any stale symlinks for older versions of the same mod that still
// point at modSrc. The target directory is created if missing.
func Install(modSrc, modsDir string) error {
	info, err := readModInfo(modSrc)
	if err != nil {
		return err
	}
	if err := fsMkdirAll(modsDir, 0o750); err != nil {
		return fmt.Errorf("create mods dir %s: %w", modsDir, err)
	}

	link := filepath.Join(modsDir, fmt.Sprintf("%s_%s", info.Name, info.Version))
	if err := removeStaleLinks(modsDir, info.Name, modSrc, link); err != nil {
		return err
	}
	if err := fsRemoveAll(link); err != nil {
		return fmt.Errorf("remove existing %s: %w", link, err)
	}
	if err := fsSymlink(modSrc, link); err != nil {
		return fmt.Errorf("symlink %s -> %s: %w", link, modSrc, err)
	}

	slog.Info("linked", "from", link, "to", modSrc)
	return nil
}

// Uninstall removes the `<name>_<version>` symlink in modsDir if it exists
// and is a symlink. Anything else (regular file, real directory) is left
// alone so users don't accidentally lose a hand-installed mod copy.
func Uninstall(modSrc, modsDir string) error {
	info, err := readModInfo(modSrc)
	if err != nil {
		return err
	}
	link := filepath.Join(modsDir, fmt.Sprintf("%s_%s", info.Name, info.Version))

	stat, err := fsLstat(link)
	if errors.Is(err, fs.ErrNotExist) {
		slog.Info("not present", "path", link)
		return nil
	}
	if err != nil {
		return err
	}
	if stat.Mode()&os.ModeSymlink == 0 {
		slog.Warn("not a symlink, leaving alone", "path", link)
		return nil
	}
	if err := fsRemove(link); err != nil {
		return err
	}
	slog.Info("unlinked", "path", link)
	return nil
}

// readModInfo loads the name+version pair Factorio uses to identify a mod.
func readModInfo(modSrc string) (modInfo, error) {
	path := filepath.Join(modSrc, "info.json")
	data, err := os.ReadFile(path) //nolint:gosec // operator-supplied mod source path; not attacker-controlled
	if err != nil {
		return modInfo{}, fmt.Errorf("read %s: %w", path, err)
	}
	var info modInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return modInfo{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if info.Name == "" || info.Version == "" {
		return modInfo{}, fmt.Errorf("%s missing name or version", path)
	}
	return info, nil
}

// removeStaleLinks deletes any other `<name>_*` entries in modsDir that are
// symlinks pointing at modSrc, leaving the one at keepLink. Factorio refuses
// to load duplicate name_version entries, so old version symlinks must go.
func removeStaleLinks(modsDir, name, modSrc, keepLink string) error {
	entries, err := fsReadDir(modsDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	prefix := name + "_"
	for _, e := range entries {
		full := filepath.Join(modsDir, e.Name())
		if full == keepLink || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		stat, err := fsLstat(full)
		if err != nil || stat.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target, err := fsReadlink(full)
		if err != nil || target != modSrc {
			continue
		}
		if err := fsRemove(full); err != nil {
			return err
		}
		slog.Info("removed stale symlink", "path", full)
	}
	return nil
}
