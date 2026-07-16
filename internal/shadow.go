// Shadow projection: per-direction silhouette flattened to black, vertically
// squashed (anchored at the foot row), and sheared right (sun upper-left).
// Composes the 8dir, running, and running-with-gun shadow sheets.

package gopher

import (
	"fmt"
	"image"
	"log/slog"
	"math"
	"path/filepath"
)

const (
	shadowAlpha    = 110  // 0-255; ~43% opacity
	shadowShearDeg = 25.0 // sun upper-left
	shadowSquashY  = 0.55 // vertical foreshortening
)

// makeShadow turns a sprite into its flat black ground-projected shadow.
func makeShadow(src *image.NRGBA) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()

	flat := clone(src)
	blacken(flat)
	scaleAlpha(flat, shadowAlpha)

	newH := int(float64(h) * shadowSquashY)
	squashed := resize(flat, w, newH)
	canvas := newCanvas(w, h)
	pasteAt(canvas, squashed, 0, h-newH)

	return shearRight(canvas, math.Tan(shadowShearDeg*math.Pi/180.0))
}

// makeFittedShadow preserves wide silhouettes by shearing on an expanded
// canvas, then fitting the result back into one sprite frame. Knight frames
// nearly fill their source canvas, so the fixed-width transform in makeShadow
// would otherwise clip their sword and shield against the right edge.
func makeFittedShadow(src *image.NRGBA) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()

	flat := clone(src)
	blacken(flat)
	scaleAlpha(flat, shadowAlpha)

	newH := int(float64(h) * shadowSquashY)
	squashed := resize(flat, w, newH)
	tan := math.Tan(shadowShearDeg * math.Pi / 180.0)
	maxOffset := int(math.Ceil(float64(h) * tan))
	expanded := newCanvas(w+maxOffset, h)
	pasteAt(expanded, squashed, 0, h-newH)
	projected := shearRight(expanded, tan)

	bounds, ok := alphaBounds(projected)
	if !ok {
		return newCanvas(w, h)
	}
	trimmed := newCanvas(bounds.Dx(), bounds.Dy())
	pasteAt(trimmed, projected.SubImage(bounds), 0, 0)

	margin := max(1, w/64)
	maxWidth := w - 2*margin
	if trimmed.Bounds().Dx() > maxWidth {
		trimmed = resize(trimmed, maxWidth, trimmed.Bounds().Dy())
	}

	out := newCanvas(w, h)
	x := (w - trimmed.Bounds().Dx()) / 2
	pasteAt(out, trimmed, x, bounds.Min.Y)
	return out
}

func alphaBounds(img *image.NRGBA) (image.Rectangle, bool) {
	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	found := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.Pix[y*img.Stride+x*4+3] == 0 {
				continue
			}
			minX = min(minX, x)
			minY = min(minY, y)
			maxX = max(maxX, x)
			maxY = max(maxY, y)
			found = true
		}
	}
	if !found {
		return image.Rectangle{}, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), true
}

// Shadow generates all three shadow sheets in gfxDir.
func Shadow(gfxDir string) error {
	shadows := make(map[string]*image.NRGBA, len(directions))
	for _, d := range directions {
		src, err := loadPNG(spritePath(gfxDir, d))
		if err != nil {
			return fmt.Errorf("load %s: %w", d, err)
		}
		shadows[d] = makeShadow(src)
	}

	if err := writeShadowIdle(gfxDir, shadows); err != nil {
		return err
	}
	if err := writeShadowRunning(gfxDir, shadows); err != nil {
		return err
	}
	return writeShadowGun(gfxDir, shadows)
}

// writeShadowIdle stitches the eight per-direction shadows into the idle
// 8-frame shadow sheet (single column).
func writeShadowIdle(gfxDir string, shadows map[string]*image.NRGBA) error {
	sheet := newCanvas(frameSize, frameSize*len(directions))
	for i, d := range directions {
		pasteAt(sheet, shadows[d], 0, frameSize*i)
	}

	out := filepath.Join(gfxDir, "gopher-shadow-8dir.png")
	if err := savePNG(out, sheet); err != nil {
		return err
	}
	slog.Info("wrote sheet", "path", out, "width", frameSize, "height", frameSize*len(directions))
	return nil
}

// writeShadowRunning replicates each direction's shadow across the eight
// run-cycle frames, since the shadow doesn't animate with the body.
func writeShadowRunning(gfxDir string, shadows map[string]*image.NRGBA) error {
	sheet := newCanvas(frameSize*frames, frameSize*len(directions))
	for ri, d := range directions {
		for fi := 0; fi < frames; fi++ {
			pasteAt(sheet, shadows[d], frameSize*fi, frameSize*ri)
		}
	}

	out := filepath.Join(gfxDir, "gopher-shadow-running.png")
	if err := savePNG(out, sheet); err != nil {
		return err
	}
	slog.Info("wrote sheet", "path", out, "width", frameSize*frames, "height", frameSize*len(directions))
	return nil
}

// writeShadowGun lays the 18 gun-frame shadows into a 6×3 grid that matches
// the running_with_gun sheet so layers align frame-for-frame.
func writeShadowGun(gfxDir string, shadows map[string]*image.NRGBA) error {
	const cols, rows = 6, 3
	sheet := newCanvas(frameSize*cols, frameSize*rows)
	for i, d := range gunMapping {
		c := i % cols
		r := i / cols
		pasteAt(sheet, shadows[d], frameSize*c, frameSize*r)
	}

	out := filepath.Join(gfxDir, "gopher-shadow-running-with-gun.png")
	if err := savePNG(out, sheet); err != nil {
		return err
	}
	slog.Info("wrote sheet", "path", out, "width", frameSize*cols, "height", frameSize*rows)
	return nil
}
