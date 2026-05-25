package gopher

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeModInfo writes a minimal mod/info.json into dir.
func writeModInfo(t *testing.T, dir, version string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "info.json"),
		[]byte(`{"name":"gopher","version":"`+version+`"}`),
		0o600,
	))
}

// withFSHook temporarily replaces a package-level fs hook with the given
// stub and restores it via t.Cleanup. Tests use this to force specific
// os-call failures that would otherwise need privileged filesystem tricks.
func withFSHook[T any](t *testing.T, slot *T, stub T) {
	t.Helper()
	prev := *slot
	*slot = stub
	t.Cleanup(func() { *slot = prev })
}

// errBoom is a sentinel returned by injected hooks so tests can
// require.ErrorIs on the exact error rather than message matching.
var errBoom = errors.New("boom")

// symlinkInfo is a stub os.FileInfo whose Mode reports the symlink bit set.
// Used to fake "this path is a symlink" without actually creating one.
type symlinkInfo struct{}

func (symlinkInfo) Name() string       { return "stub" }
func (symlinkInfo) Size() int64        { return 0 }
func (symlinkInfo) Mode() os.FileMode  { return os.ModeSymlink }
func (symlinkInfo) ModTime() time.Time { return time.Time{} }
func (symlinkInfo) IsDir() bool        { return false }
func (symlinkInfo) Sys() any           { return nil }

func TestModsDirFor(t *testing.T) {
	homeOK := func() (string, error) { return "/home/u", nil }
	envEmpty := func(string) string { return "" }
	envAppdata := func(k string) string {
		if k == "APPDATA" {
			return `C:\Users\u\AppData\Roaming`
		}
		return ""
	}
	homeErr := func() (string, error) { return "", errors.New("boom") }

	cases := []struct {
		name    string
		goos    string
		home    func() (string, error)
		env     func(string) string
		want    string
		wantErr string
	}{
		{name: "darwin", goos: "darwin", home: homeOK, env: envEmpty, want: filepath.Join("/home/u", "Library", "Application Support", "factorio", "mods")},
		{name: "linux", goos: "linux", home: homeOK, env: envEmpty, want: filepath.Join("/home/u", ".factorio", "mods")},
		{name: "windows", goos: "windows", home: homeOK, env: envAppdata, want: filepath.Join(`C:\Users\u\AppData\Roaming`, "Factorio", "mods")},
		{name: "windows missing APPDATA", goos: "windows", home: homeOK, env: envEmpty, wantErr: "APPDATA"},
		{name: "darwin home error", goos: "darwin", home: homeErr, env: envEmpty, wantErr: "boom"},
		{name: "linux home error", goos: "linux", home: homeErr, env: envEmpty, wantErr: "boom"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := modsDirFor(tc.goos, tc.home, tc.env)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDefaultModsDir(t *testing.T) {
	got, err := DefaultModsDir()
	require.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestInstallSymlinks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	modsDir := filepath.Join(root, "mods")
	writeModInfo(t, modSrc, "0.0.1")

	require.NoError(t, Install(modSrc, modsDir))

	link := filepath.Join(modsDir, "gopher_0.0.1")
	stat, err := os.Lstat(link)
	require.NoError(t, err)
	assert.NotZero(t, stat.Mode()&os.ModeSymlink, "expected a symlink")

	target, err := os.Readlink(link)
	require.NoError(t, err)
	assert.Equal(t, modSrc, target)
}

func TestInstallReplacesStaleVersionLinks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	modsDir := filepath.Join(root, "mods")
	require.NoError(t, os.MkdirAll(modsDir, 0o750))
	stale := filepath.Join(modsDir, "gopher_0.3.0")
	require.NoError(t, os.Symlink(modSrc, stale))

	writeModInfo(t, modSrc, "0.4.0")
	require.NoError(t, Install(modSrc, modsDir))

	_, err := os.Lstat(stale)
	assert.True(t, os.IsNotExist(err), "stale symlink should be removed")
	_, err = os.Lstat(filepath.Join(modsDir, "gopher_0.4.0"))
	assert.NoError(t, err, "new versioned symlink should exist")
}

func TestInstallLeavesUnrelatedLinks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	other := filepath.Join(root, "other-mod")
	modsDir := filepath.Join(root, "mods")
	require.NoError(t, os.MkdirAll(other, 0o750))
	require.NoError(t, os.MkdirAll(modsDir, 0o750))
	keepLink := filepath.Join(modsDir, "gopher_0.3.0")
	require.NoError(t, os.Symlink(other, keepLink)) // points elsewhere

	writeModInfo(t, modSrc, "0.0.1")
	require.NoError(t, Install(modSrc, modsDir))

	_, err := os.Lstat(keepLink)
	assert.NoError(t, err, "symlink pointing elsewhere must be preserved")
}

func TestInstallReplacesExistingLink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	modsDir := filepath.Join(root, "mods")
	require.NoError(t, os.MkdirAll(modsDir, 0o750))
	link := filepath.Join(modsDir, "gopher_0.0.1")
	require.NoError(t, os.WriteFile(link, []byte("stub"), 0o600))

	writeModInfo(t, modSrc, "0.0.1")
	require.NoError(t, Install(modSrc, modsDir))

	stat, err := os.Lstat(link)
	require.NoError(t, err)
	assert.NotZero(t, stat.Mode()&os.ModeSymlink, "expected a symlink to replace the file")
}

// TestInfoJSONErrors covers the three failure modes of the info.json
// parser: file missing, file present but malformed, and file present but
// missing required fields. Both Install and Uninstall route through the
// same readModInfo helper, so we exercise each fixture under both fns to
// pin the contract.
func TestInfoJSONErrors(t *testing.T) {
	cases := []struct {
		name      string
		writeInfo func(t *testing.T, dir string) // nil means leave dir absent
		wantErr   string
	}{
		{
			name:    "missing info.json",
			wantErr: "info.json",
		},
		{
			name: "malformed JSON",
			writeInfo: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(dir, 0o750))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "info.json"), []byte("not json"), 0o600))
			},
			wantErr: "parse",
		},
		{
			name: "missing required fields",
			writeInfo: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(dir, 0o750))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "info.json"), []byte(`{"name":""}`), 0o600))
			},
			wantErr: "missing name or version",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			modSrc := filepath.Join(root, "mod")
			if tc.writeInfo != nil {
				tc.writeInfo(t, modSrc)
			}

			err := Install(modSrc, filepath.Join(root, "mods"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)

			err = Uninstall(modSrc, filepath.Join(root, "mods"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestInstallFSErrors covers Install's defensive error paths by injecting
// failures at each os call. Each row sets up the hook (and any prerequisite
// state on disk), then asserts Install returns the wrapped sentinel error.
func TestInstallFSErrors(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T, modSrc, modsDir string) // optional disk prep
		hook    func(t *testing.T)                         // mandatory hook injection
		wantMsg string                                     // optional substring assertion
	}{
		{
			name: "MkdirAll fails",
			hook: func(t *testing.T) {
				withFSHook(t, &fsMkdirAll, func(string, os.FileMode) error { return errBoom })
			},
			wantMsg: "create mods dir",
		},
		{
			name: "RemoveAll fails",
			hook: func(t *testing.T) {
				withFSHook(t, &fsRemoveAll, func(string) error { return errBoom })
			},
			wantMsg: "remove existing",
		},
		{
			name: "Symlink fails",
			hook: func(t *testing.T) {
				withFSHook(t, &fsSymlink, func(string, string) error { return errBoom })
			},
			wantMsg: "symlink",
		},
		{
			name: "stale-link scan fails",
			hook: func(t *testing.T) {
				withFSHook(t, &fsReadDir, func(string) ([]os.DirEntry, error) { return nil, errBoom })
			},
		},
		{
			name: "stale-link remove fails",
			setup: func(t *testing.T, modSrc, modsDir string) {
				require.NoError(t, os.MkdirAll(modsDir, 0o750))
				require.NoError(t, os.Symlink(modSrc, filepath.Join(modsDir, "gopher_0.3.0")))
			},
			hook: func(t *testing.T) {
				withFSHook(t, &fsRemove, func(string) error { return errBoom })
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			modSrc := filepath.Join(root, "mod")
			modsDir := filepath.Join(root, "mods")
			if tc.setup != nil {
				tc.setup(t, modSrc, modsDir)
			}
			writeModInfo(t, modSrc, "0.4.0")
			tc.hook(t)

			err := Install(modSrc, modsDir)
			require.Error(t, err)
			require.ErrorIs(t, err, errBoom)
			if tc.wantMsg != "" {
				assert.Contains(t, err.Error(), tc.wantMsg)
			}
		})
	}
}

// TestUninstallFSErrors covers Uninstall's defensive error paths via hook
// injection. The Lstat-error case fakes the Lstat result; the Remove-error
// case adds a Lstat stub that reports a symlink so the function reaches the
// fsRemove call.
func TestUninstallFSErrors(t *testing.T) {
	cases := []struct {
		name string
		hook func(t *testing.T)
	}{
		{
			name: "Lstat fails",
			hook: func(t *testing.T) {
				withFSHook(t, &fsLstat, func(string) (os.FileInfo, error) { return nil, errBoom })
			},
		},
		{
			name: "Remove fails",
			hook: func(t *testing.T) {
				withFSHook(t, &fsLstat, func(string) (os.FileInfo, error) { return symlinkInfo{}, nil })
				withFSHook(t, &fsRemove, func(string) error { return errBoom })
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			modSrc := filepath.Join(root, "mod")
			writeModInfo(t, modSrc, "0.0.1")
			tc.hook(t)

			err := Uninstall(modSrc, filepath.Join(root, "mods"))
			require.Error(t, err)
			require.ErrorIs(t, err, errBoom)
		})
	}
}

// TestInstallStaleLinkScanMissingDirIsOK forces the read-dir to return
// fs.ErrNotExist; removeStaleLinks must short-circuit cleanly.
func TestInstallStaleLinkScanMissingDirIsOK(t *testing.T) {
	withFSHook(t, &fsReadDir, func(string) ([]os.DirEntry, error) { return nil, fs.ErrNotExist })

	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	writeModInfo(t, modSrc, "0.0.1")

	require.NoError(t, Install(modSrc, filepath.Join(root, "mods")))
}

// TestInstallStaleLinkLstatFails forces fsLstat to fail inside the
// removeStaleLinks loop. The matching name-prefixed entry exists, so the
// loop reaches the Lstat call; the failure must be skipped silently
// (Install still succeeds because the stale-link removal is best-effort).
func TestInstallStaleLinkLstatFails(t *testing.T) {
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	modsDir := filepath.Join(root, "mods")
	require.NoError(t, os.MkdirAll(modsDir, 0o750))
	require.NoError(t, os.Symlink(modSrc, filepath.Join(modsDir, "gopher_0.3.0")))
	writeModInfo(t, modSrc, "0.4.0")

	withFSHook(t, &fsLstat, func(string) (os.FileInfo, error) { return nil, errBoom })

	require.NoError(t, Install(modSrc, modsDir))
}

// TestInstallStaleLinkReadlinkMismatch covers the Readlink branch where the
// returned target differs from modSrc, so the entry must NOT be removed.
func TestInstallStaleLinkReadlinkMismatch(t *testing.T) {
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	other := filepath.Join(root, "other")
	modsDir := filepath.Join(root, "mods")
	require.NoError(t, os.MkdirAll(other, 0o750))
	require.NoError(t, os.MkdirAll(modsDir, 0o750))
	keep := filepath.Join(modsDir, "gopher_0.3.0")
	require.NoError(t, os.Symlink(other, keep))
	writeModInfo(t, modSrc, "0.4.0")

	// Force Readlink to return an unexpected error so the != modSrc branch
	// (and thus the "leave it alone" continue) is hit. The success path is
	// already covered by TestInstallLeavesUnrelatedLinks.
	withFSHook(t, &fsReadlink, func(string) (string, error) { return "", errBoom })

	require.NoError(t, Install(modSrc, modsDir))

	_, err := os.Lstat(keep)
	assert.NoError(t, err, "unrelated symlink must be preserved")
}

func TestUninstallRemovesSymlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	modsDir := filepath.Join(root, "mods")
	writeModInfo(t, modSrc, "0.0.1")
	require.NoError(t, Install(modSrc, modsDir))

	require.NoError(t, Uninstall(modSrc, modsDir))
	_, err := os.Lstat(filepath.Join(modsDir, "gopher_0.0.1"))
	assert.True(t, os.IsNotExist(err))
}

func TestUninstallNoOpWhenMissing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	writeModInfo(t, modSrc, "0.0.1")
	assert.NoError(t, Uninstall(modSrc, filepath.Join(root, "mods")))
}

func TestUninstallLeavesNonSymlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	modSrc := filepath.Join(root, "mod")
	modsDir := filepath.Join(root, "mods")
	writeModInfo(t, modSrc, "0.0.1")
	require.NoError(t, os.MkdirAll(modsDir, 0o750))
	link := filepath.Join(modsDir, "gopher_0.0.1")
	require.NoError(t, os.WriteFile(link, []byte("real file"), 0o600))

	require.NoError(t, Uninstall(modSrc, modsDir))
	_, err := os.Lstat(link)
	assert.NoError(t, err, "real file must not be removed")
}
