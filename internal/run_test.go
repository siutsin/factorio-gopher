package gopher

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunGaitEventsMatchLua(t *testing.T) {
	var leftLandings, rightLandings []int
	for frame, pose := range runGait {
		previous := runGait[(frame+len(runGait)-1)%len(runGait)]
		if previous.leftFoot > 0 && pose.leftFoot == 0 {
			leftLandings = append(leftLandings, frame+1)
		}
		if previous.rightFoot > 0 && pose.rightFoot == 0 {
			rightLandings = append(rightLandings, frame+1)
		}
	}

	require.Len(t, leftLandings, 1)
	require.Len(t, rightLandings, 1)
	assert.Zero(t, groundRunPoses[leftLandings[0]-1].lift, "knight left step must touch ground")
	assert.Zero(t, groundRunPoses[rightLandings[0]-1].lift, "knight right step must touch ground")
	lua, err := os.ReadFile(filepath.Join(repoRoot(), "mod", "data-updates.lua"))
	require.NoError(t, err)
	assert.Contains(t, string(lua), "local LEFT_STEP_FRAME = "+strconv.Itoa(leftLandings[0]))
	assert.Contains(t, string(lua), "local RIGHT_STEP_FRAME = "+strconv.Itoa(rightLandings[0]))
}

func TestRunBobKeepsLandingsAtBaseline(t *testing.T) {
	for frame := range frames {
		assert.GreaterOrEqual(t, runBob(frame), 0)
		assert.LessOrEqual(t, runBob(frame), bobAmp)
	}
	assert.Zero(t, runBob(1))
	assert.Zero(t, runBob(5))
	assert.Equal(t, bobAmp, runBob(3))
	assert.Equal(t, bobAmp, runBob(7))
}

func TestRunFramesPreserveBottomPixels(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, frameSize, frameSize))
	src.SetNRGBA(frameSize/2, frameSize-1, color.NRGBA{R: 1, A: 255})
	for frame := range frames {
		out := makeRunFrame(src, runBob(frame), frame)
		opaque := 0
		for offset := 3; offset < len(out.Pix); offset += 4 {
			if out.Pix[offset] != 0 {
				opaque++
			}
		}
		assert.Equal(t, 1, opaque, "frame %d", frame+1)
	}
}

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
				assert.Equal(
					t,
					image.Rect(0, 0, runtimeFrameSize()*frames, runtimeFrameSize()*len(directions)),
					out.Bounds(),
				)
			})
		})
	}
}
