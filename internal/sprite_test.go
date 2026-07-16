// Low-level pixel-helper tests. The previous external sprite package merged
// in here verbatim, plus the close-helper internal test. Public API names
// went down a tier (Load → loadPNG, Save → savePNG, etc) when the package
// merged into gopher.

package gopher

import (
	"errors"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errCloser is a stub io.Closer that returns a fixed error from Close,
// letting closeWithErr's defer path be tested without involving real files.
type errCloser struct{ err error }

// Close returns the canned error so tests can verify the defer wraps it.
func (e errCloser) Close() error { return e.err }

// TestCloseWithErr covers all three branches: clean close, close failure
// when no prior error, and close failure when a prior error must win.
func TestCloseWithErr(t *testing.T) {
	prior := errors.New("decode failed")
	closeErr := errors.New("disk full")

	cases := []struct {
		name      string
		closeErr  error
		startErr  error
		wantErrIs error  // exact identity expected via assert.Same; nil means err must be nil
		wantMsg   string // substring expected in err.Error(); empty means no check
	}{
		{name: "close success leaves err nil"},
		{name: "close error sets err when nil", closeErr: closeErr, wantMsg: "close: disk full"},
		{name: "close error preserves prior err", closeErr: closeErr, startErr: prior, wantErrIs: prior},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.startErr
			closeWithErr(errCloser{err: tc.closeErr}, &err)
			switch {
			case tc.wantErrIs != nil:
				assert.Same(t, tc.wantErrIs, err)
			case tc.wantMsg != "":
				require.ErrorContains(t, err, tc.wantMsg)
			default:
				require.NoError(t, err)
			}
		})
	}
}

// TestLoadSavePNGRoundtrip writes an NRGBA image and reads it back to confirm
// pixel-byte equality, guarding against any silent format conversion.
func TestLoadSavePNGRoundtrip(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	src.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 255})
	src.SetNRGBA(3, 3, color.NRGBA{R: 10, G: 20, B: 30, A: 128})

	path := filepath.Join(t.TempDir(), "round.png")
	require.NoError(t, savePNG(path, src))

	got, err := loadPNG(path)
	require.NoError(t, err)
	assert.Equal(t, src.Bounds(), got.Bounds())
	assert.Equal(t, src.Pix, got.Pix)
}

// TestLoadPNGConvertsNonNRGBA feeds a paletted PNG through loadPNG to exercise
// the non-NRGBA conversion branch.
func TestLoadPNGConvertsNonNRGBA(t *testing.T) {
	pal := color.Palette{color.NRGBA{}, color.NRGBA{R: 200, A: 255}}
	src := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	src.SetColorIndex(1, 1, 1)

	path := filepath.Join(t.TempDir(), "paletted.png")
	require.NoError(t, savePNG(path, src))

	got, err := loadPNG(path)
	require.NoError(t, err)
	assert.Equal(t, src.Bounds(), got.Bounds())
}

// TestLoadPNGErrors covers the two failure modes: missing file (open error)
// and a file that exists but isn't a valid PNG (decode error).
func TestLoadPNGErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		_, err := loadPNG(filepath.Join(t.TempDir(), "nope.png"))
		assert.Error(t, err)
	})
	t.Run("not a PNG", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "garbage.png")
		require.NoError(t, os.WriteFile(path, []byte("not a png"), 0o600))
		_, err := loadPNG(path)
		assert.Error(t, err)
	})
}

// TestSavePNGErrorBadPath confirms savePNG surfaces the os.Create error when
// the parent directory doesn't exist.
func TestSavePNGErrorBadPath(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "no_such_dir", "out.png")
	err := savePNG(bad, image.NewNRGBA(image.Rect(0, 0, 1, 1)))
	assert.Error(t, err)
}

// TestNewCanvas checks that the returned canvas has the requested dimensions
// and is fully transparent (every Pix byte is zero).
func TestNewCanvas(t *testing.T) {
	img := newCanvas(3, 5)
	assert.Equal(t, image.Rect(0, 0, 3, 5), img.Bounds())
	for _, b := range img.Pix {
		assert.EqualValues(t, 0, b, "expected fully transparent canvas")
	}
}

// TestClone verifies the clone shares no underlying storage with the source
// by mutating the clone and re-reading the original.
func TestClone(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 10, A: 255})
	src.SetNRGBA(1, 1, color.NRGBA{R: 20, A: 255})

	dst := clone(src)
	assert.Equal(t, src.Bounds(), dst.Bounds())

	dst.SetNRGBA(0, 0, color.NRGBA{R: 99, A: 255})
	assert.EqualValues(t, 10, src.NRGBAAt(0, 0).R, "source mutated by clone edit")
}

// TestPasteAt asserts pixels land at the target offset and that cells
// outside the paste rect remain transparent.
func TestPasteAt(t *testing.T) {
	dst := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 1, A: 255})
	src.SetNRGBA(1, 0, color.NRGBA{R: 2, A: 255})
	src.SetNRGBA(0, 1, color.NRGBA{R: 3, A: 255})
	src.SetNRGBA(1, 1, color.NRGBA{R: 4, A: 255})

	pasteAt(dst, src, 1, 2)

	wants := []struct {
		x, y int
		r    uint8
	}{
		{1, 2, 1}, {2, 2, 2},
		{1, 3, 3}, {2, 3, 4},
	}
	for _, w := range wants {
		assert.Equal(t, w.r, dst.NRGBAAt(w.x, w.y).R, "dst at %d,%d", w.x, w.y)
	}
	assert.EqualValues(t, 0, dst.NRGBAAt(0, 0).A, "untouched cell stays transparent")
}

// TestResize sanity-checks that the output bounds match the requested size
// and at least one pixel is opaque (i.e. CatmullRom didn't zero everything).
func TestResize(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			src.SetNRGBA(x, y, color.NRGBA{R: 200, A: 255})
		}
	}
	out := resize(src, 2, 2)
	assert.Equal(t, image.Rect(0, 0, 2, 2), out.Bounds())
	assert.NotZero(t, out.NRGBAAt(0, 0).A, "resized output has no opaque pixels")
}

// TestOverlaySkipsTransparent confirms src pixels with alpha 0 leave the
// destination untouched, while opaque src pixels overwrite.
func TestOverlaySkipsTransparent(t *testing.T) {
	dst := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	dst.SetNRGBA(0, 0, color.NRGBA{R: 1, G: 2, B: 3, A: 255})
	dst.SetNRGBA(1, 0, color.NRGBA{R: 4, G: 5, B: 6, A: 255})

	src := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	src.SetNRGBA(0, 0, color.NRGBA{}) // alpha 0, must not overwrite
	src.SetNRGBA(1, 0, color.NRGBA{R: 99, G: 99, B: 99, A: 255})

	overlay(dst, src)

	assert.Equal(t, color.NRGBA{R: 1, G: 2, B: 3, A: 255}, dst.NRGBAAt(0, 0))
	assert.Equal(t, color.NRGBA{R: 99, G: 99, B: 99, A: 255}, dst.NRGBAAt(1, 0))
}

// TestShiftUp covers the three branches of shiftUp (no-op, full clear,
// partial shift) plus the boundary at amount == h.
func TestShiftUp(t *testing.T) {
	build := func() *image.NRGBA {
		img := image.NewNRGBA(image.Rect(0, 0, 1, 4))
		for y := range 4 {
			img.SetNRGBA(0, y, color.NRGBA{R: uint8(y + 1), A: 255})
		}
		return img
	}

	cases := []struct {
		name   string
		amount int
		want   []uint8 // R values per row, top to bottom
	}{
		{"zero amount returns copy", 0, []uint8{1, 2, 3, 4}},
		{"negative returns copy", -5, []uint8{1, 2, 3, 4}},
		{"shift by one", 1, []uint8{2, 3, 4, 0}},
		{"shift by three", 3, []uint8{4, 0, 0, 0}},
		{"shift equals height", 4, []uint8{0, 0, 0, 0}},
		{"shift exceeds height", 10, []uint8{0, 0, 0, 0}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := shiftUp(build(), tc.amount)
			for y, want := range tc.want {
				assert.Equal(t, want, out.NRGBAAt(0, y).R, "row %d", y)
			}
		})
	}
}

// TestScaleAlpha verifies the multiplier maths at the two extremes (0, 255)
// and one mid-value, and confirms RGB is left untouched.
func TestScaleAlpha(t *testing.T) {
	cases := []struct {
		name      string
		factor    uint8
		startA    uint8
		wantAlpha uint8
	}{
		{"factor 255 keeps alpha", 255, 200, 200},
		{"factor 0 zeroes alpha", 0, 200, 0},
		{"factor 128 halves alpha", 128, 200, uint8(uint16(200) * 128 / 255)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
			img.SetNRGBA(0, 0, color.NRGBA{R: 50, G: 60, B: 70, A: tc.startA})

			scaleAlpha(img, tc.factor)

			got := img.NRGBAAt(0, 0)
			assert.Equal(t, tc.wantAlpha, got.A)
			assert.Equal(t, color.NRGBA{R: 50, G: 60, B: 70, A: tc.wantAlpha}, got, "RGB must not change")
		})
	}
}

func TestClampByte(t *testing.T) {
	assert.EqualValues(t, 0, clampByte(0))
	assert.EqualValues(t, 255, clampByte(255))
	assert.EqualValues(t, 255, clampByte(256))
}

// TestBlackenPreservesAlpha asserts RGB is zeroed but the per-pixel alpha
// channel survives unchanged.
func TestBlackenPreservesAlpha(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 128})
	img.SetNRGBA(1, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 0})

	blacken(img)

	assert.Equal(t, color.NRGBA{A: 128}, img.NRGBAAt(0, 0))
	assert.Equal(t, color.NRGBA{}, img.NRGBAAt(1, 0))
}

// TestShearRightAnchorsBottom checks the foot-row anchor: pixels in the
// bottom row shift the least, and pixels at the top can shear off-canvas.
func TestShearRightAnchorsBottom(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		src.SetNRGBA(1, y, color.NRGBA{R: 255, A: 255})
	}

	out := shearRight(src, 1.0)

	assert.EqualValues(t, 255, out.NRGBAAt(2, 3).R, "bottom row shifted right by 1")
	assert.EqualValues(t, 0, out.NRGBAAt(1, 3).A, "source x=1 should be empty after shift")
	for x := range 4 {
		assert.EqualValues(t, 0, out.NRGBAAt(x, 0).A, "y=0 x=%d fully sheared off", x)
	}
}
