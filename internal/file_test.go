package gopher

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileHelpersFollowSymlink(t *testing.T) {
	targetPath := filepath.Join(t.TempDir(), "target.png")
	linkPath := filepath.Join(t.TempDir(), "link.png")

	initial := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	require.NoError(t, savePNG(targetPath, initial))
	require.NoError(t, os.Symlink(targetPath, linkPath))

	data, err := readRootedFile(linkPath)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	replacement := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	replacement.SetNRGBA(0, 0, color.NRGBA{R: 42, A: 255})
	require.NoError(t, savePNG(linkPath, replacement))

	got, err := loadPNG(targetPath)
	require.NoError(t, err)
	assert.Equal(t, replacement.Pix, got.Pix)
}

func TestSavePNGFollowsDanglingSymlink(t *testing.T) {
	cases := []struct {
		name       string
		linkTarget func(linkDir, targetDir string) string
		targetPath func(linkDir, targetDir string) string
	}{
		{
			name:       "absolute target",
			linkTarget: func(_, targetDir string) string { return filepath.Join(targetDir, "target.png") },
			targetPath: func(_, targetDir string) string { return filepath.Join(targetDir, "target.png") },
		},
		{
			name:       "relative target",
			linkTarget: func(_, _ string) string { return "target.png" },
			targetPath: func(linkDir, _ string) string { return filepath.Join(linkDir, "target.png") },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			linkDir := t.TempDir()
			targetDir := t.TempDir()
			linkPath := filepath.Join(linkDir, "link.png")
			targetPath := tc.targetPath(linkDir, targetDir)
			require.NoError(t, os.Symlink(tc.linkTarget(linkDir, targetDir), linkPath))

			img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
			img.SetNRGBA(0, 0, color.NRGBA{G: 42, A: 255})
			require.NoError(t, savePNG(linkPath, img))

			got, err := loadPNG(targetPath)
			require.NoError(t, err)
			assert.Equal(t, img.Pix, got.Pix)
		})
	}
}

func TestWritablePathRejectsSymlinkLoop(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first")
	second := filepath.Join(dir, "second")
	require.NoError(t, os.Symlink(second, first))
	require.NoError(t, os.Symlink(first, second))

	_, err := writablePath(first)
	assert.Error(t, err)
}
