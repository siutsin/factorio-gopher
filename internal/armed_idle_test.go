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

func TestArmedIdle(t *testing.T) {
	t.Parallel()

	t.Run("writes body and shadow sheets", func(t *testing.T) {
		t.Parallel()
		dir := filledDir(t)
		require.NoError(t, ArmedIdle(dir))
		for _, name := range []string{
			"gopher-idle-with-gun.png",
			"gopher-idle-with-gun-shadow.png",
			"knight-idle-with-gun.png",
			"knight-idle-with-gun-shadow.png",
		} {
			img, err := loadPNG(filepath.Join(dir, name))
			require.NoError(t, err)
			assert.Equal(t, image.Rect(0, 0, runtimeFrameSize(), runtimeFrameSize()*len(directions)), img.Bounds())
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
			wantError: "load gopher-n armed-idle source",
		},
		{
			name: "invalid gopher dimensions",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				writeStubPNG(t, spritePath(dir, "e"), 1, 1)
			},
			wantError: "gopher-e armed-idle source dimensions",
		},
		{
			name: "gopher body save error",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				blockOutput(t, filepath.Join(dir, "gopher-idle-with-gun.png"))
			},
			wantError: "save gopher armed-idle sheet",
		},
		{
			name: "gopher shadow save error",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				blockOutput(t, filepath.Join(dir, "gopher-idle-with-gun-shadow.png"))
			},
			wantError: "save gopher armed-idle shadow sheet",
		},
		{
			name: "missing knight source",
			setup: func(t *testing.T, dir string) {
				writeTestSources(t, dir)
			},
			wantError: "load knight-n armed-idle source",
		},
		{
			name: "invalid knight dimensions",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				writeStubPNG(t, knightPath(dir, "e"), 1, 1)
			},
			wantError: "knight-e armed-idle source dimensions",
		},
		{
			name: "knight body save error",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				blockOutput(t, filepath.Join(dir, "knight-idle-with-gun.png"))
			},
			wantError: "save knight armed-idle sheet",
		},
		{
			name: "knight shadow save error",
			setup: func(t *testing.T, dir string) {
				writeAllSources(t, dir)
				blockOutput(t, filepath.Join(dir, "knight-idle-with-gun-shadow.png"))
			},
			wantError: "save knight armed-idle shadow sheet",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			tc.setup(t, dir)
			err := ArmedIdle(dir)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantError)
		})
	}
}

func TestDrawGunChangesFrame(t *testing.T) {
	t.Parallel()
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	before := clone(img)
	drawGun(img, 0, len(directions))
	assert.NotEqual(t, before.Pix, img.Pix)
	assert.NotEqual(t, color.NRGBA{}, img.NRGBAAt(16, 19))
}

func TestArmedGunUsesFactorioHalfSweep(t *testing.T) {
	t.Parallel()
	assert.InDelta(t, -math.Pi/2, armedGunAngle(0), 0.0001)
	assert.InDelta(t, math.Pi/2, armedGunAngle(len(gunMapping)-1), 0.0001)
}
