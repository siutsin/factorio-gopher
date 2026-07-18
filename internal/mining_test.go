package gopher

import (
	"image"
	"image/color"
	"math"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMining(t *testing.T) {
	t.Parallel()

	t.Run("writes directional cycles and shadows", func(t *testing.T) {
		t.Parallel()
		dir := filledDir(t)
		require.NoError(t, Mining(dir))
		outputs := []struct {
			name string
			size int
		}{
			{name: "gopher-mining.png", size: runtimeFrameSize()},
			{name: "gopher-mining-shadow.png", size: runtimeFrameSize()},
			{name: "knight-mining.png", size: runtimeFrameSize()},
			{name: "knight-mining-shadow.png", size: runtimeFrameSize()},
		}
		for _, output := range outputs {
			img, err := loadPNG(filepath.Join(dir, output.name))
			require.NoError(t, err)
			assert.Equal(t, image.Rect(0, 0, output.size*frames, output.size*len(directions)), img.Bounds())
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
			wantError: "load gopher-n mining source",
		},
		{
			name: "invalid gopher dimensions",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				writeStubPNG(t, spritePath(dir, "e"), 1, 1)
			},
			wantError: "gopher-e mining source dimensions",
		},
		{
			name: "gopher body save error",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				blockOutput(t, filepath.Join(dir, "gopher-mining.png"))
			},
			wantError: "save gopher mining sheet",
		},
		{
			name: "gopher shadow save error",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				blockOutput(t, filepath.Join(dir, "gopher-mining-shadow.png"))
			},
			wantError: "save gopher mining shadow sheet",
		},
		{
			name: "missing knight source",
			setup: func(t *testing.T, dir string) {
				writeTestSources(t, dir)
			},
			wantError: "load knight-n mining source",
		},
		{
			name: "invalid knight dimensions",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				writeStubPNG(t, knightPath(dir, "e"), 1, 1)
			},
			wantError: "knight-e mining source dimensions",
		},
		{
			name: "knight body save error",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				blockOutput(t, filepath.Join(dir, "knight-mining.png"))
			},
			wantError: "save knight mining sheet",
		},
		{
			name: "knight shadow save error",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				blockOutput(t, filepath.Join(dir, "knight-mining-shadow.png"))
			},
			wantError: "save knight mining shadow sheet",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			tc.setup(t, dir)
			err := Mining(dir)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantError)
		})
	}
}

func TestMiningFrameTransformsBodyWithoutAddingArtwork(t *testing.T) {
	t.Parallel()
	src := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	bodyColour := color.NRGBA{R: 255, A: 255}
	for y := 8; y < 16; y++ {
		for x := 4; x < 12; x++ {
			src.SetNRGBA(x, y, bodyColour)
		}
	}

	out := makeMiningFrame(src, 0)
	assert.NotEqual(t, src.Pix, out.Pix)
	opaque := 0
	for y := range out.Bounds().Dy() {
		for x := range out.Bounds().Dx() {
			pixel := out.NRGBAAt(x, y)
			if pixel.A == 0 {
				continue
			}
			opaque++
			assert.Equal(t, bodyColour, pixel)
		}
	}
	assert.Positive(t, opaque)
}

func TestRotateImageFitsOpaqueArtworkInsideFrame(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := range 32 {
		for x := range 32 {
			src.SetNRGBA(x, y, color.NRGBA{A: 255})
		}
	}

	out := rotateImage(src, math.Pi/8)
	_, visible := alphaBounds(out)
	require.True(t, visible)
	for offset := range 32 {
		assert.Zero(t, out.NRGBAAt(offset, 0).A, "top edge at x=%d", offset)
		assert.Zero(t, out.NRGBAAt(offset, 31).A, "bottom edge at x=%d", offset)
		assert.Zero(t, out.NRGBAAt(0, offset).A, "left edge at y=%d", offset)
		assert.Zero(t, out.NRGBAAt(31, offset).A, "right edge at y=%d", offset)
	}

	empty := rotateImage(image.NewNRGBA(image.Rect(0, 0, 32, 32)), math.Pi/8)
	_, emptyVisible := alphaBounds(empty)
	assert.False(t, emptyVisible)
}
