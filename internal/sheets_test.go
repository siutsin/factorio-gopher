package gopher

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSheets covers the success path, the missing-source error, and the
// per-output save-error branch for each of the two stitched sheets.
func TestSheets(t *testing.T) {
	outputs := []string{
		"gopher-8dir.png",
		"gopher-running-with-gun-1.png",
		"gopher-running-with-gun-1-shadow.png",
		"gopher-running-with-gun-2.png",
		"gopher-running-with-gun-2-shadow.png",
		"gopher-running-with-gun-flipped-1-shadow.png",
		"gopher-running-with-gun-flipped-2-shadow.png",
	}

	cases := []struct {
		name      string
		blockSave string
		wantErr   string
	}{
		{name: "success"},
		{name: "missing source", wantErr: "load"},
		{name: "save error - 8dir", blockSave: "gopher-8dir.png"},
		{name: "save error - gun 1", blockSave: "gopher-running-with-gun-1.png"},
		{name: "save error - gun shadow 1", blockSave: "gopher-running-with-gun-1-shadow.png"},
		{name: "save error - gun 2", blockSave: "gopher-running-with-gun-2.png"},
		{name: "save error - gun shadow 2", blockSave: "gopher-running-with-gun-2-shadow.png"},
		{
			name:      "save error - flipped gun shadow 1",
			blockSave: "gopher-running-with-gun-flipped-1-shadow.png",
		},
		{
			name:      "save error - flipped gun shadow 2",
			blockSave: "gopher-running-with-gun-flipped-2-shadow.png",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runGenerator(t, Sheets, tc.wantErr, tc.blockSave, func(dir string) {
				for _, name := range outputs {
					_, err := loadPNG(filepath.Join(dir, name))
					assert.NoError(t, err, "missing %s", name)
				}
			})
		})
	}
}

func TestArmedGunTracksBodyBob(t *testing.T) {
	t.Parallel()
	bodyColour := color.NRGBA{R: 10, G: 120, B: 220, A: 255}
	src := image.NewNRGBA(image.Rect(0, 0, frameSize, frameSize))
	for y := frameSize / 6; y < 5*frameSize/6; y++ {
		for x := frameSize / 4; x < 3*frameSize/4; x++ {
			src.SetNRGBA(x, y, bodyColour)
		}
	}

	wantOffset := 0
	for frame := range frames {
		out := makeGopherArmedRunFrame(src, 0, frame, frameSize)
		bodyBounds, bodyOK := matchingPixelBounds(out, func(pixel color.NRGBA) bool {
			return pixel == bodyColour
		})
		gunBounds, gunOK := matchingPixelBounds(out, func(pixel color.NRGBA) bool {
			return pixel.A != 0 && pixel != bodyColour
		})
		require.True(t, bodyOK, "frame %d body", frame+1)
		require.True(t, gunOK, "frame %d gun", frame+1)
		offset := gunBounds.Min.Y - bodyBounds.Min.Y
		if frame == 0 {
			wantOffset = offset
		}
		assert.Equal(t, wantOffset, offset, "frame %d", frame+1)
	}
}

func matchingPixelBounds(
	img *image.NRGBA,
	matches func(pixel color.NRGBA) bool,
) (image.Rectangle, bool) {
	bounds := img.Bounds()
	result := image.Rectangle{Min: bounds.Max, Max: bounds.Min}
	found := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if !matches(img.NRGBAAt(x, y)) {
				continue
			}
			result.Min.X = min(result.Min.X, x)
			result.Min.Y = min(result.Min.Y, y)
			result.Max.X = max(result.Max.X, x+1)
			result.Max.Y = max(result.Max.Y, y+1)
			found = true
		}
	}
	return result, found
}
