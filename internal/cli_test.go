package gopher

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeAllSources fills dir with the gopher and knight per-direction stub
// PNGs that the build pipeline expects to load. Pixels are
// uniformly transparent because the CLI tests only care about exit codes,
// not sheet contents.
func writeAllSources(t *testing.T, dir string) {
	t.Helper()
	for _, d := range []string{"n", "ne", "e", "se", "s", "sw", "w", "nw"} {
		writeStubPNG(t, filepath.Join(dir, "gopher-"+d+".png"), frameSize, frameSize)
		writeStubPNG(t, filepath.Join(dir, "knight-"+d+".png"), frameSize, frameSize)
	}
}

// writeStubPNG writes a transparent w×h PNG to path. Used as a minimal source
// fixture by table-driven dispatch tests.
func writeStubPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	require.NoError(t, savePNG(path, img))
}

// cliSetup builds a gfx dir for a table-driven case and returns its path.
type cliSetup func(t *testing.T) string

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errBoom }

// emptyDir returns a fresh temp dir with no source PNGs, simulating a
// brand-new repo where the load step is expected to fail.
func emptyDir(t *testing.T) string { t.Helper(); return t.TempDir() }

// filledDir returns a temp dir prepopulated with the eight stub sources so
// the build pipeline can run end-to-end successfully.
func filledDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeAllSources(t, dir)
	return dir
}

// filledDirBlocking returns a setup that prepopulates sources and then plants
// a directory at the named output path so the build step's save call fails
// with EISDIR.
func filledDirBlocking(name string) cliSetup {
	return func(t *testing.T) string {
		t.Helper()
		dir := filledDir(t)
		path := filepath.Join(dir, name)
		writeStubPNG(t, path, 1, 1)
		require.NoError(t, os.Remove(path))
		require.NoError(t, os.Mkdir(path, 0o700))
		return dir
	}
}

// filledMod returns a setup that creates a mod source dir with info.json so
// install/uninstall have something to symlink. Tests read both <root>/mod
// and <root>/mods from the returned root.
func filledMod(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "mod"), 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "mod", "info.json"),
		[]byte(`{"name":"gopher","version":"0.0.1"}`),
		0o600,
	))
	return root
}

// TestCLI covers every flag-parsing branch and every dispatch outcome of
// CLI(): usage, parse error, unknown subcommand, each subcommand's success
// path, missing-source error, and the "all" pipeline's stop-at-first-failure
// behaviour at both the Run and Shadow steps. Install and uninstall paths
// are exercised with explicit -mods to avoid touching the host's Factorio
// directory.
func TestCLI(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		setup      cliSetup
		args       func(gfx string) []string
		wantExit   int
		wantStderr string
		wantAbsent string
	}{
		{
			name:       "missing command prints usage and exits 2",
			setup:      emptyDir,
			args:       func(string) []string { return nil },
			wantExit:   2,
			wantStderr: "usage",
		},
		{
			name:     "flag parse error exits 2",
			setup:    emptyDir,
			args:     func(string) []string { return []string{"-unknown"} },
			wantExit: 2,
		},
		{
			name:     "unknown subcommand exits 1",
			setup:    emptyDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "bogus"} },
			wantExit: 1,
		},
		{
			name:     "running succeeds",
			setup:    filledDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "running"} },
			wantExit: 0,
		},
		{
			name:     "shadow succeeds",
			setup:    filledDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "shadow"} },
			wantExit: 0,
		},
		{
			name:     "sheets succeeds",
			setup:    filledDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "sheets"} },
			wantExit: 0,
		},
		{
			name:     "knight succeeds",
			setup:    filledDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "knight"} },
			wantExit: 0,
		},
		{
			name:     "corpse succeeds",
			setup:    filledDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "corpse"} },
			wantExit: 0,
		},
		{
			name:     "mining succeeds",
			setup:    filledDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "mining"} },
			wantExit: 0,
		},
		{
			name:     "armed idle succeeds",
			setup:    filledDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "armed-idle"} },
			wantExit: 0,
		},
		{
			name:     "all succeeds",
			setup:    filledDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "all"} },
			wantExit: 0,
		},
		{
			name:     "running with missing source exits 1",
			setup:    emptyDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "running"} },
			wantExit: 1,
		},
		{
			name:     "shadow with missing source exits 1",
			setup:    emptyDir,
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "shadow"} },
			wantExit: 1,
		},
		{
			name:       "all stops at first failing step",
			setup:      emptyDir,
			args:       func(gfx string) []string { return []string{"-gfx", gfx, "all"} },
			wantExit:   1,
			wantAbsent: "gopher-shadow-8dir.png",
		},
		{
			name:       "all stops when shadow step fails",
			setup:      filledDirBlocking("gopher-shadow-8dir.png"),
			args:       func(gfx string) []string { return []string{"-gfx", gfx, "all"} },
			wantExit:   1,
			wantAbsent: "gopher-8dir.png",
		},
		{
			name:       "all stops when sheets step fails",
			setup:      filledDirBlocking("gopher-8dir.png"),
			args:       func(gfx string) []string { return []string{"-gfx", gfx, "all"} },
			wantExit:   1,
			wantAbsent: "knight-idle.png",
		},
		{
			name:       "all stops when knight step fails",
			setup:      filledDirBlocking("knight-running-with-gun-1.png"),
			args:       func(gfx string) []string { return []string{"-gfx", gfx, "all"} },
			wantExit:   1,
			wantAbsent: "gopher-corpse.png",
		},
		{
			name:       "all stops when corpse step fails",
			setup:      filledDirBlocking("gopher-corpse.png"),
			args:       func(gfx string) []string { return []string{"-gfx", gfx, "all"} },
			wantExit:   1,
			wantAbsent: "gopher-mining.png",
		},
		{
			name:       "all stops when mining step fails",
			setup:      filledDirBlocking("gopher-mining.png"),
			args:       func(gfx string) []string { return []string{"-gfx", gfx, "all"} },
			wantExit:   1,
			wantAbsent: "gopher-idle-with-gun.png",
		},
		{
			name:     "all stops when armed idle step fails",
			setup:    filledDirBlocking("gopher-idle-with-gun.png"),
			args:     func(gfx string) []string { return []string{"-gfx", gfx, "all"} },
			wantExit: 1,
		},
		{
			name:  "install succeeds with explicit mods dir",
			setup: filledMod,
			args: func(root string) []string {
				return []string{
					"-mod", filepath.Join(root, "mod"),
					"-mods", filepath.Join(root, "mods"),
					"install",
				}
			},
			wantExit: 0,
		},
		{
			name:  "uninstall succeeds with explicit mods dir",
			setup: filledMod,
			args: func(root string) []string {
				return []string{
					"-mod", filepath.Join(root, "mod"),
					"-mods", filepath.Join(root, "mods"),
					"uninstall",
				}
			},
			wantExit: 0,
		},
		{
			name:  "install missing info exits 1",
			setup: emptyDir,
			args: func(root string) []string {
				return []string{
					"-mod", filepath.Join(root, "missing"),
					"-mods", filepath.Join(root, "mods"),
					"install",
				}
			},
			wantExit: 1,
		},
		{
			name:  "uninstall missing info exits 1",
			setup: emptyDir,
			args: func(root string) []string {
				return []string{
					"-mod", filepath.Join(root, "missing"),
					"-mods", filepath.Join(root, "mods"),
					"uninstall",
				}
			},
			wantExit: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gfx := tc.setup(t)
			var stderr bytes.Buffer
			assert.Equal(t, tc.wantExit, CLI(tc.args(gfx), &stderr))
			if tc.wantStderr != "" {
				assert.Contains(t, stderr.String(), tc.wantStderr)
			}
			if tc.wantAbsent != "" {
				assert.NoFileExists(t, filepath.Join(gfx, tc.wantAbsent))
			}
		})
	}
}

func TestCLIUsageWriteError(t *testing.T) {
	assert.Equal(t, 2, CLI(nil, errorWriter{}))
}

// TestCLIInstallDefaultModsDir runs install with no -mods flag, exercising
// the resolveModsDir branch that calls DefaultModsDir. The test redirects
// HOME (and APPDATA on Windows) to a temp dir so it doesn't touch the real
// Factorio install.
func TestCLIInstallDefaultModsDir(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("APPDATA", root)

	modSrc := filepath.Join(root, "mod")
	require.NoError(t, os.MkdirAll(modSrc, 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(modSrc, "info.json"),
		[]byte(`{"name":"gopher","version":"0.0.1"}`),
		0o600,
	))

	var stderr bytes.Buffer
	assert.Equal(t, 0, CLI([]string{"-mod", modSrc, "install"}, &stderr))
}

// TestDefaultGfxDir confirms the resolved default points at an absolute path
// ending in "graphics", regardless of where the test binary is invoked from.
func TestDefaultGfxDir(t *testing.T) {
	got := DefaultGfxDir()
	assert.True(t, filepath.IsAbs(got), "default gfx dir should be absolute")
	assert.Equal(t, "graphics", filepath.Base(got))
}

// TestDefaultModDir confirms the resolved default points at <repo>/mod.
func TestDefaultModDir(t *testing.T) {
	got := DefaultModDir()
	assert.True(t, filepath.IsAbs(got))
	assert.Equal(t, "mod", filepath.Base(got))
}

// TestCLIDefaultModsDirFails forces the default mods-dir lookup to fail
// (e.g. Windows without APPDATA), exercising the resolveModsDir error
// branch in dispatch's install and uninstall paths.
func TestCLIDefaultModsDirFails(t *testing.T) {
	for _, cmd := range []string{"install", "uninstall"} {
		t.Run(cmd, func(t *testing.T) {
			withFSHook(t, &defaultModsDirFn, func() (string, error) {
				return "", errBoom
			})

			var stderr bytes.Buffer
			assert.Equal(t, 1, CLI([]string{cmd}, &stderr))
		})
	}
}

// TestRepoRoot confirms repoRoot resolves to a non-empty absolute-style path.
func TestRepoRoot(t *testing.T) {
	got := repoRoot()
	require.NotEmpty(t, got)
	assert.True(t, filepath.IsAbs(got))
}

// TestRepoRootFallback covers the branch where the embedded source path no
// longer resolves to an existing directory (e.g. binary built with -trimpath
// or relocated). The fallback returns "." so dependent defaults are
// obviously invalid rather than silently wrong.
func TestRepoRootFallback(t *testing.T) {
	withFSHook(t, &callerFile, func() string { return "/nonexistent/path/file.go" })
	assert.Equal(t, ".", repoRoot())
}
