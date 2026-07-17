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

func writeCorpseSources(t *testing.T, dir string) {
	t.Helper()
	for _, name := range []string{"gopher-s.png", "knight-s.png"} {
		img := image.NewNRGBA(image.Rect(0, 0, frameSize, frameSize))
		img.SetNRGBA(frameSize/4, frameSize/3, color.NRGBA{R: 200, A: 255})
		require.NoError(t, savePNG(filepath.Join(dir, name), img))
	}
}

func TestCorpse(t *testing.T) {
	t.Parallel()

	t.Run("writes body and shadow sheets", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeCorpseSources(t, dir)
		require.NoError(t, Corpse(dir))

		outputs := []struct {
			name string
			size int
		}{
			{name: "gopher-corpse.png", size: runtimeFrameSize()},
			{name: "gopher-corpse-shadow.png", size: runtimeFrameSize()},
			{name: "knight-corpse.png", size: runtimeFrameSize()},
			{name: "knight-corpse-shadow.png", size: runtimeFrameSize()},
		}
		for _, output := range outputs {
			img, err := loadPNG(filepath.Join(dir, output.name))
			require.NoError(t, err)
			assert.Equal(t, output.size*corpseFrames, img.Bounds().Dx(), output.name)
			assert.Equal(t, output.size, img.Bounds().Dy(), output.name)
		}
	})

	cases := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		wantError string
	}{
		{
			name:      "missing gopher source",
			setup:     func(*testing.T, string) {},
			wantError: "load gopher corpse source",
		},
		{
			name: "invalid gopher dimensions",
			setup: func(t *testing.T, dir string) {
				writeStubPNG(t, filepath.Join(dir, "gopher-s.png"), 1, 1)
			},
			wantError: "gopher corpse source dimensions",
		},
		{
			name: "gopher body save error",
			setup: func(t *testing.T, dir string) {
				writeCorpseSources(t, dir)
				blockOutput(t, filepath.Join(dir, "gopher-corpse.png"))
			},
			wantError: "save gopher corpse sheet",
		},
		{
			name: "gopher shadow save error",
			setup: func(t *testing.T, dir string) {
				writeCorpseSources(t, dir)
				blockOutput(t, filepath.Join(dir, "gopher-corpse-shadow.png"))
			},
			wantError: "save gopher corpse shadow sheet",
		},
		{
			name: "missing knight source",
			setup: func(t *testing.T, dir string) {
				writeStubPNG(t, filepath.Join(dir, "gopher-s.png"), frameSize, frameSize)
			},
			wantError: "load knight corpse source",
		},
		{
			name: "invalid knight dimensions",
			setup: func(t *testing.T, dir string) {
				writeStubPNG(t, filepath.Join(dir, "gopher-s.png"), frameSize, frameSize)
				writeStubPNG(t, filepath.Join(dir, "knight-s.png"), 1, 1)
			},
			wantError: "knight corpse source dimensions",
		},
		{
			name: "knight body save error",
			setup: func(t *testing.T, dir string) {
				writeCorpseSources(t, dir)
				blockOutput(t, filepath.Join(dir, "knight-corpse.png"))
			},
			wantError: "save knight corpse sheet",
		},
		{
			name: "knight shadow save error",
			setup: func(t *testing.T, dir string) {
				writeCorpseSources(t, dir)
				blockOutput(t, filepath.Join(dir, "knight-corpse-shadow.png"))
			},
			wantError: "save knight corpse shadow sheet",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			tc.setup(t, dir)
			err := Corpse(dir)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantError)
		})
	}
}

func TestCorpseFramesUseOppositeHorizontalPoses(t *testing.T) {
	t.Parallel()
	src := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	src.SetNRGBA(2, 3, color.NRGBA{R: 255, A: 255})

	clockwise := rotateQuarterTurn(src, true)
	counterClockwise := rotateQuarterTurn(src, false)
	assert.Equal(t, color.NRGBA{R: 255, A: 255}, clockwise.NRGBAAt(12, 2))
	assert.Equal(t, color.NRGBA{R: 255, A: 255}, counterClockwise.NRGBAAt(3, 13))

	left := makeCorpseFrame(src, false)
	right := makeCorpseFrame(src, true)
	assert.NotEqual(t, left.Pix, right.Pix)
}

func TestCorpseFramesStayGrounded(t *testing.T) {
	t.Parallel()
	const size = 32
	src := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := size / 3; y < size; y++ {
		for x := size / 4; x < size/2; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 200, A: 255})
		}
	}

	for _, clockwise := range []bool{false, true} {
		body := makeCorpseFrame(src, clockwise)
		bodyBounds, bodyOK := alphaBounds(body)
		require.True(t, bodyOK)
		assert.Greater(t, bodyBounds.Dx(), bodyBounds.Dy(), "clockwise=%v body", clockwise)
		assert.Equal(t, size, bodyBounds.Max.Y, "clockwise=%v body", clockwise)

		shadowBounds, shadowOK := alphaBounds(makeFittedShadow(body))
		require.True(t, shadowOK)
		assert.Equal(t, size, shadowBounds.Max.Y, "clockwise=%v shadow", clockwise)
	}

	empty := makeCorpseFrame(image.NewNRGBA(image.Rect(0, 0, size, size)), true)
	_, ok := alphaBounds(empty)
	assert.False(t, ok)
}

func TestCorpseOutputCollisionIsDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeCorpseSources(t, dir)
	path := filepath.Join(dir, "gopher-corpse.png")
	require.NoError(t, os.Mkdir(path, 0o700))
	require.Error(t, Corpse(dir))
}
