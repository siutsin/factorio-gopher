package gopher

import (
	"image"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsBeige covers the centre, the in-tolerance edge, the just-outside
// edge on each channel, and far-away black/white.
func TestIsBeige(t *testing.T) {
	cases := []struct {
		name    string
		r, g, b uint8
		want    bool
	}{
		{"exact beige centre", beigeR, beigeG, beigeB, true},
		{"within tolerance on each channel", beigeR + 10, beigeG - 10, beigeB + 5, true},
		{"R just outside tolerance", beigeR + beigeTol, beigeG, beigeB, false},
		{"G just outside tolerance", beigeR, beigeG + beigeTol, beigeB, false},
		{"B just outside tolerance", beigeR, beigeG, beigeB + beigeTol, false},
		{"black far outside", 0, 0, 0, false},
		{"white far outside", 255, 255, 255, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isBeige(tc.r, tc.g, tc.b))
		})
	}
}

// TestAbs spot-checks positive, negative, and zero.
func TestAbs(t *testing.T) {
	cases := []struct{ in, want int }{
		{5, 5},
		{-5, 5},
		{0, 0},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, abs(tc.in))
	}
}

// TestRun exercises Run's success path plus its two error branches:
// missing source PNGs (load fails) and a planted directory at the output
// path (save fails).
func TestRun(t *testing.T) {
	cases := []struct {
		name      string
		blockSave string
		wantErr   string
	}{
		{name: "success"},
		{name: "missing source", wantErr: "load"},
		{name: "save error", blockSave: "gopher-running.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runGenerator(t, Run, tc.wantErr, tc.blockSave, func(dir string) {
				out, err := loadPNG(filepath.Join(dir, "gopher-running.png"))
				require.NoError(t, err)
				assert.Equal(t, image.Rect(0, 0, frameSize*frames, frameSize*len(directions)), out.Bounds())
			})
		})
	}
}
