package gopher

import (
	"crypto/sha256"
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeVisualSources(t *testing.T, dir string) {
	t.Helper()
	writeTestSources(t, dir)
	colour := uint8(40)
	for _, direction := range directions {
		img := image.NewNRGBA(image.Rect(0, 0, frameSize, frameSize))
		for y := frameSize / 4; y < 7*frameSize/8; y++ {
			for x := frameSize / 4; x < 3*frameSize/4; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: colour, G: 100, B: 140, A: 255})
			}
		}
		for x := 3 * frameSize / 4; x < frameSize; x++ {
			img.SetNRGBA(x, frameSize/2, color.NRGBA{R: 180, G: 220, B: 255, A: 255})
		}
		require.NoError(t, savePNG(knightPath(dir, direction), img))
		colour += 20
	}
}

func TestAnimationContactSheetsContainDistinctFrames(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeVisualSources(t, dir)
	require.NoError(t, buildAll(dir))

	assertFrameDiversity(t, dir, "gopher-running.png", runtimeFrameSize(), frames, frames)
	assertFrameDiversity(t, dir, "gopher-running-with-gun-1.png", runtimeFrameSize(), frames, frames)
	assertFrameDiversity(t, dir, "gopher-mining.png", runtimeFrameSize(), frames, 5)
	assertFrameDiversity(t, dir, "gopher-corpse.png", runtimeFrameSize(), corpseFrames, corpseFrames)
	assertFrameDiversity(t, dir, "knight-running.png", runtimeFrameSize(), frames, 5)
	assertFrameDiversity(t, dir, "knight-running-with-gun-1.png", runtimeFrameSize(), frames, 5)
	assertFrameDiversity(t, dir, "knight-mining.png", runtimeFrameSize(), frames, 5)
	for _, name := range []string{
		"gopher-running-with-gun-1-shadow.png",
		"gopher-running-with-gun-2-shadow.png",
		"gopher-running-with-gun-flipped-1-shadow.png",
		"gopher-running-with-gun-flipped-2-shadow.png",
		"knight-running-with-gun-1-shadow.png",
		"knight-running-with-gun-2-shadow.png",
		"knight-running-with-gun-flipped-1-shadow.png",
		"knight-running-with-gun-flipped-2-shadow.png",
	} {
		assertFramesIdentical(t, dir, name, runtimeFrameSize(), frames)
	}
}

func assertFrameDiversity(
	t *testing.T,
	dir, name string,
	size, frameCount, minimumUnique int,
) {
	t.Helper()
	sheet, err := loadPNG(filepath.Join(dir, name))
	require.NoError(t, err)
	require.Equal(t, frameCount*size, sheet.Bounds().Dx(), name)
	require.Zero(t, sheet.Bounds().Dy()%size, name)
	for row := range sheet.Bounds().Dy() / size {
		hashes := make(map[[sha256.Size]byte]int, frameCount)
		duplicates := make([]image.Point, 0)
		for frame := range frameCount {
			rect := image.Rect(frame*size, row*size, (frame+1)*size, (row+1)*size)
			frameImage := image.NewNRGBA(image.Rect(0, 0, size, size))
			pasteAt(frameImage, sheet.SubImage(rect), 0, 0)
			_, visible := alphaBounds(frameImage)
			assert.True(t, visible, "%s row %d frame %d should not be blank", name, row, frame)
			hash := sha256.Sum256(frameImage.Pix)
			if first, exists := hashes[hash]; exists {
				duplicates = append(duplicates, image.Pt(first+1, frame+1))
			} else {
				hashes[hash] = frame
			}
		}
		assert.GreaterOrEqual(
			t,
			len(hashes),
			minimumUnique,
			"%s row %d should contain at least %d distinct frames; duplicate pairs: %v",
			name,
			row,
			minimumUnique,
			duplicates,
		)
	}
}

func assertFramesIdentical(t *testing.T, dir, name string, size, frameCount int) {
	t.Helper()
	sheet, err := loadPNG(filepath.Join(dir, name))
	require.NoError(t, err)
	require.Equal(t, frameCount*size, sheet.Bounds().Dx(), name)
	require.Zero(t, sheet.Bounds().Dy()%size, name)
	for row := range sheet.Bounds().Dy() / size {
		var first [sha256.Size]byte
		for frame := range frameCount {
			rect := image.Rect(frame*size, row*size, (frame+1)*size, (row+1)*size)
			frameImage := image.NewNRGBA(image.Rect(0, 0, size, size))
			pasteAt(frameImage, sheet.SubImage(rect), 0, 0)
			hash := sha256.Sum256(frameImage.Pix)
			if frame == 0 {
				first = hash
				continue
			}
			assert.Equal(t, first, hash, "%s row %d frame %d", name, row, frame)
		}
	}
}
