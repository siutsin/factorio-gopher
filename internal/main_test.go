package gopher

import (
	"image"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain silences slog so build progress messages don't pollute test logs.
// It also shrinks frameSize and the run-animation layout constants so
// integration tests process 64x64 sources and 16x16 runtime frames instead
// of the production pipeline's 1024x1024 sources and 256x256 runtime frames.
func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.DiscardHandler))
	SetFrameSize(64)
	os.Exit(m.Run())
}

// writeTestSources creates a frameSize×frameSize PNG per direction in dir.
// Each PNG is painted to exercise every branch of makeRunFrame: opaque
// non-beige body pixels, opaque beige pixels in both the foot band and the
// arm band, on both halves of the image, plus transparent pixels.
func writeTestSources(t *testing.T, dir string) {
	t.Helper()
	armY := (armBandTop + armBandBot) / 2
	footY := (footBandTop + footBandBot) / 2
	leftX := frameSize / 4
	rightX := 3 * frameSize / 4
	beige := color.NRGBA{R: beigeR, G: beigeG, B: beigeB, A: 255}
	for _, d := range directions {
		img := image.NewNRGBA(image.Rect(0, 0, frameSize, frameSize))
		img.SetNRGBA(0, 0, color.NRGBA{R: 50, G: 50, B: 50, A: 255}) // non-beige body
		img.SetNRGBA(leftX, armY, beige)
		img.SetNRGBA(rightX, armY, beige)
		img.SetNRGBA(leftX, footY, beige)
		img.SetNRGBA(rightX, footY, beige)
		img.SetNRGBA(1, 0, color.NRGBA{}) // transparent path
		require.NoError(t, savePNG(spritePath(dir, d), img))
	}
}

// blockOutput plants a directory at path so a future os.Create on it returns
// EISDIR, simulating a save failure for that specific output file.
func blockOutput(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, savePNG(path, image.NewNRGBA(image.Rect(0, 0, 1, 1))))
	require.NoError(t, os.Remove(path))
	require.NoError(t, os.Mkdir(path, 0o700))
}

// runGenerator drives the success/failure flow shared by Run/Shadow/Sheets
// table tests.
//   - wantErr non-empty: fn must error and the message must contain wantErr.
//   - blockSave non-empty: fn must error (any message), after planting a
//     directory at that output path.
//   - both empty: fn must succeed and assertSuccess is invoked with dir.
func runGenerator(t *testing.T, fn func(string) error, wantErr, blockSave string, assertSuccess func(dir string)) {
	t.Helper()
	t.Parallel()
	dir := t.TempDir()
	if wantErr == "" {
		writeTestSources(t, dir)
	}
	if blockSave != "" {
		blockOutput(t, filepath.Join(dir, blockSave))
	}

	err := fn(dir)
	switch {
	case wantErr != "":
		require.Error(t, err)
		assert.Contains(t, err.Error(), wantErr)
	case blockSave != "":
		require.Error(t, err)
	default:
		require.NoError(t, err)
		assertSuccess(dir)
	}
}
