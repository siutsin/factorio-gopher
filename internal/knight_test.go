package gopher

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeKnightSources creates the eight authored direction frames used by
// Knight.
func writeKnightSources(t *testing.T, dir string) {
	t.Helper()
	for i, d := range directions {
		img := image.NewNRGBA(image.Rect(0, 0, frameSize, frameSize))
		img.SetNRGBA(i, i, color.NRGBA{R: uint8(i + 1), A: 255})
		require.NoError(t, savePNG(knightPath(dir, d), img))
	}
}

func TestKnight(t *testing.T) {
	outputs := []struct {
		name       string
		width      int
		directions int
	}{
		{name: "knight-idle.png", width: 1, directions: 8},
		{name: "knight-idle-shadow.png", width: 1, directions: 8},
		{name: "knight-running.png", width: frames, directions: 8},
		{name: "knight-running-shadow.png", width: frames, directions: 8},
		{name: "knight-running-with-gun.png", width: 6, directions: 3},
		{name: "knight-running-with-gun-shadow.png", width: 6, directions: 3},
		{name: "knight-take-off.png", width: takeOffFrames, directions: 8},
		{name: "knight-take-off-shadow.png", width: takeOffFrames, directions: 8},
		{name: "knight-hover.png", width: hoverFrames, directions: 8},
		{name: "knight-hover-shadow.png", width: hoverFrames, directions: 8},
	}

	cases := []struct {
		name      string
		blockSave string
		wantErr   string
	}{
		{name: "success"},
		{name: "missing source", wantErr: "load knight-n"},
		{name: "save error - idle", blockSave: "knight-idle.png"},
		{name: "save error - idle shadow", blockSave: "knight-idle-shadow.png"},
		{name: "save error - running", blockSave: "knight-running.png"},
		{name: "save error - running shadow", blockSave: "knight-running-shadow.png"},
		{name: "save error - gun", blockSave: "knight-running-with-gun.png"},
		{name: "save error - gun shadow", blockSave: "knight-running-with-gun-shadow.png"},
		{name: "save error - takeoff", blockSave: "knight-take-off.png"},
		{name: "save error - takeoff shadow", blockSave: "knight-take-off-shadow.png"},
		{name: "save error - hover", blockSave: "knight-hover.png"},
		{name: "save error - hover shadow", blockSave: "knight-hover-shadow.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if tc.wantErr == "" {
				writeKnightSources(t, dir)
			}
			if tc.blockSave != "" {
				writeKnightSources(t, dir)
				blockOutput(t, filepath.Join(dir, tc.blockSave))
			}

			err := Knight(dir)
			switch {
			case tc.wantErr != "":
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			case tc.blockSave != "":
				require.Error(t, err)
			default:
				require.NoError(t, err)
				for _, output := range outputs {
					img, lerr := loadPNG(filepath.Join(dir, output.name))
					require.NoError(t, lerr, "missing %s", output.name)
					assert.Equal(
						t,
						image.Rect(
							0,
							0,
							flightFrameSize()*output.width,
							flightFrameSize()*output.directions,
						),
						img.Bounds(),
					)
					_, populated := alphaBounds(img)
					assert.True(t, populated, "%s should contain visible pixels", output.name)
				}
			}
		})
	}
}

func TestKnightRejectsWrongDimensions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeKnightSources(t, dir)
	require.NoError(t, savePNG(knightPath(dir, "e"), image.NewNRGBA(image.Rect(0, 0, 1, 1))))

	err := Knight(dir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "knight-e dimensions are 1x1")
}

func TestWriteKnightGunUsesFactorioDirectionOrder(t *testing.T) {
	t.Parallel()
	size := flightFrameSize()
	srcs := make(map[string]*image.NRGBA, len(directions))
	for i, direction := range directions {
		img := image.NewNRGBA(image.Rect(0, 0, size, size))
		for y := range size {
			for x := range size {
				img.SetNRGBA(x, y, color.NRGBA{R: uint8(i + 1), A: 255})
			}
		}
		srcs[direction] = img
	}

	dir := t.TempDir()
	require.NoError(t, writeKnightGun(dir, srcs))
	sheet, err := loadPNG(filepath.Join(dir, "knight-running-with-gun.png"))
	require.NoError(t, err)

	for i, direction := range gunMapping {
		column := i % 6
		row := i / 6
		want := srcs[direction].NRGBAAt(size/2, size/2)
		got := sheet.NRGBAAt(column*size+size/2, row*size+size/2)
		assert.Equal(t, want, got, "cell %d should use %s", i, direction)
	}
}

func TestMakeFittedShadowKeepsWideSpriteInsideFrame(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, frameSize, frameSize))
	for y := range frameSize {
		for x := range frameSize {
			src.SetNRGBA(x, y, color.NRGBA{A: 255})
		}
	}

	out := makeFittedShadow(src)

	assert.Equal(t, src.Bounds(), out.Bounds())
	for y := range frameSize {
		assert.Zero(t, out.NRGBAAt(0, y).A, "left edge at y=%d", y)
		assert.Zero(t, out.NRGBAAt(frameSize-1, y).A, "right edge at y=%d", y)
	}
	_, ok := alphaBounds(out)
	assert.True(t, ok, "fitted shadow should retain opaque pixels")
}

func TestMakeFlightBodyTransformsInsideFrame(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 4; y < 12; y++ {
		for x := 4; x < 12; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 100, A: 255})
		}
	}

	pose := flightPose{scaleX: 1.25, scaleY: 0.5, lift: 0.125}
	out := makeFlightBody(src, pose)
	ground := makeFlightBody(src, flightPose{scaleX: pose.scaleX, scaleY: pose.scaleY})

	assert.Equal(t, src.Bounds(), out.Bounds())
	bounds, ok := alphaBounds(out)
	require.True(t, ok)
	groundBounds, groundOK := alphaBounds(ground)
	require.True(t, groundOK)
	assert.Greater(t, bounds.Dx(), 8, "horizontal stretch should widen the body")
	assert.Less(t, bounds.Dy(), 8, "vertical compression should shorten the body")
	assert.Less(t, bounds.Min.Y, groundBounds.Min.Y, "lift should move the transformed body upward")
}

func TestMakeFlightShadowTracksHeight(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 12; y < 24; y++ {
		for x := 4; x < 28; x++ {
			src.SetNRGBA(x, y, color.NRGBA{A: 200})
		}
	}

	ground := makeFlightShadow(src, 0)
	air := makeFlightShadow(src, 1)
	groundBounds, groundOK := alphaBounds(ground)
	airBounds, airOK := alphaBounds(air)
	require.True(t, groundOK)
	require.True(t, airOK)
	assert.Less(t, airBounds.Dx(), groundBounds.Dx())
	assert.Less(t, airBounds.Dy(), groundBounds.Dy())
	assert.Less(t, maxAlpha(air), maxAlpha(ground))

	empty := makeFlightShadow(image.NewNRGBA(image.Rect(0, 0, 32, 32)), 1)
	_, emptyOK := alphaBounds(empty)
	assert.False(t, emptyOK)
}

func TestFlightPoseSequencesAreContinuous(t *testing.T) {
	assert.Len(t, groundRunPoses, frames)
	assert.Len(t, takeOffPoses, takeOffFrames)
	assert.Len(t, hoverPoses, hoverFrames)
	assert.Equal(t, takeOffPoses[len(takeOffPoses)-1], hoverPoses[0])
	assert.Equal(t, hoverPoses[0], hoverPoses[len(hoverPoses)-1])
}

func maxAlpha(img *image.NRGBA) uint8 {
	var result uint8
	for i := 3; i < len(img.Pix); i += 4 {
		result = max(result, img.Pix[i])
	}
	return result
}
