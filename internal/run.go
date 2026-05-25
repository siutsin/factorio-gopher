// Run-cycle: 8-direction × 8-frame animation built procedurally from each
// direction's static sprite via a sine-wave bob and alternating arm/leg
// lifts. Limbs are colour-keyed by beige pixels.

package gopher

import (
	"fmt"
	"image"
	"log/slog"
	"math"
	"path/filepath"
)

// Animation tunables (source-pixel space). On-screen movement is these
// values × scale (0.0375 in Lua). Defined as vars so tests can shrink them
// alongside frameSize via SetFrameSize.
var (
	bobAmp   = 50
	footLift = 150
	armLift  = 120
)

const (
	frames = 8

	// Beige limb colour and tolerance per channel.
	beigeR, beigeG, beigeB = 0xB8, 0x93, 0x7F
	beigeTol               = 35
)

// Foot/arm Y-bands inside a single 1024-px sprite. Defined as vars (not
// const) so tests can rescale them when shrinking frameSize via SetFrameSize.
var (
	footBandTop, footBandBot = 820, 1024
	armBandTop, armBandBot   = 450, 750
)

// isBeige reports whether (r, g, b) lies within beigeTol of the limb colour.
// Used to colour-key arms and feet for the procedural run cycle.
func isBeige(r, g, b uint8) bool {
	return abs(int(r)-beigeR) < beigeTol &&
		abs(int(g)-beigeG) < beigeTol &&
		abs(int(b)-beigeB) < beigeTol
}

// abs returns the absolute value of x. Go's stdlib has no integer abs.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Run writes mod/graphics/gopher-running.png from the per-direction sources.
func Run(gfxDir string) error {
	sheet := newCanvas(frameSize*frames, frameSize*len(directions))

	for ri, d := range directions {
		src, err := loadPNG(spritePath(gfxDir, d))
		if err != nil {
			return fmt.Errorf("load %s: %w", d, err)
		}
		for fi := 0; fi < frames; fi++ {
			bob := int(math.Round(float64(bobAmp) * math.Sin(math.Pi*float64(fi)/2)))
			frame := makeRunFrame(src, bob, fi)
			pasteAt(sheet, frame, frameSize*fi, frameSize*ri)
		}
	}

	out := filepath.Join(gfxDir, "gopher-running.png")
	if err := savePNG(out, sheet); err != nil {
		return err
	}
	slog.Info("wrote sheet", "path", out, "width", frameSize*frames, "height", frameSize*len(directions))
	return nil
}

// limbBuffers holds the four per-limb working images plus a body image
// extracted from a single source frame.
type limbBuffers struct {
	body                *image.NRGBA
	leftFoot, rightFoot *image.NRGBA
	leftArm, rightArm   *image.NRGBA
}

// newLimbBuffers allocates a fresh w×h transparent image for each of the
// five outputs so splitFrom can write each pixel exactly once.
func newLimbBuffers(w, h int) limbBuffers {
	rect := image.Rect(0, 0, w, h)
	return limbBuffers{
		body:      image.NewNRGBA(rect),
		leftFoot:  image.NewNRGBA(rect),
		rightFoot: image.NewNRGBA(rect),
		leftArm:   image.NewNRGBA(rect),
		rightArm:  image.NewNRGBA(rect),
	}
}

// makeRunFrame builds a single run-cycle frame: image bobbed up, then beige
// limbs in the foot/arm bands lifted in opposite-pair gait.
func makeRunFrame(src *image.NRGBA, bob, frameIdx int) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()

	bobbed := image.NewNRGBA(image.Rect(0, 0, w, h))
	if bob >= 0 {
		copy(bobbed.Pix, src.Pix[bob*src.Stride:])
	} else {
		copy(bobbed.Pix[(-bob)*bobbed.Stride:], src.Pix)
	}

	limbs := newLimbBuffers(w, h)
	limbs.splitFrom(bobbed)

	// Opposite-pair gait.
	if frameIdx%2 == 0 {
		limbs.leftFoot = shiftUp(limbs.leftFoot, footLift)
		limbs.rightArm = shiftUp(limbs.rightArm, armLift)
	} else {
		limbs.rightFoot = shiftUp(limbs.rightFoot, footLift)
		limbs.leftArm = shiftUp(limbs.leftArm, armLift)
	}

	out := clone(limbs.body)
	overlay(out, limbs.leftFoot)
	overlay(out, limbs.rightFoot)
	overlay(out, limbs.leftArm)
	overlay(out, limbs.rightArm)
	return out
}

// splitFrom walks every opaque pixel of src and routes it to the body buffer
// or one of the four limb buffers based on Y-band and image half. Beige
// pixels in the foot/arm bands go to the matching limb; everything else goes
// to body.
func (l limbBuffers) splitFrom(src *image.NRGBA) {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	half := w / 2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*src.Stride + x*4
			r, g, b, a := src.Pix[i], src.Pix[i+1], src.Pix[i+2], src.Pix[i+3]
			if a == 0 {
				continue
			}
			target := l.classify(x, y, half, r, g, b)
			j := y*target.Stride + x*4
			target.Pix[j], target.Pix[j+1], target.Pix[j+2], target.Pix[j+3] = r, g, b, a
		}
	}
}

// classify returns which buffer a pixel at (x, y) with colour (r, g, b)
// belongs to: a limb buffer if the pixel is beige and inside the foot or arm
// band, body otherwise.
func (l limbBuffers) classify(x, y, half int, r, g, b uint8) *image.NRGBA {
	inFoot := y >= footBandTop && y < footBandBot
	inArm := y >= armBandTop && y < armBandBot
	if !isBeige(r, g, b) || (!inFoot && !inArm) {
		return l.body
	}
	switch {
	case inFoot && x < half:
		return l.leftFoot
	case inFoot:
		return l.rightFoot
	case x < half:
		return l.leftArm
	default:
		return l.rightArm
	}
}
