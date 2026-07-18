// Directional mining animations generated from the canonical body sprites.

package gopher

import (
	"fmt"
	"image"
	"log/slog"
	"math"
	"path/filepath"
)

var miningSwingDegrees = [...]float64{-55, -35, -10, 20, 45, 20, -10, -35}

type miningSource struct {
	name string
	path func(gfxDir, direction string) string
	size int
}

// Mining writes gopher and knight eight-direction mining cycles and shadows.
func Mining(gfxDir string) error {
	sources := []miningSource{
		{name: "gopher", path: spritePath, size: runtimeFrameSize()},
		{name: "knight", path: knightPath, size: runtimeFrameSize()},
	}
	for _, source := range sources {
		srcs := make(map[string]*image.NRGBA, len(directions))
		for _, direction := range directions {
			img, err := loadPNG(source.path(gfxDir, direction))
			if err != nil {
				return fmt.Errorf("load %s-%s mining source: %w", source.name, direction, err)
			}
			if img.Bounds().Dx() != frameSize || img.Bounds().Dy() != frameSize {
				return fmt.Errorf(
					"%s-%s mining source dimensions are %dx%d; want %dx%d",
					source.name,
					direction,
					img.Bounds().Dx(),
					img.Bounds().Dy(),
					frameSize,
					frameSize,
				)
			}
			srcs[direction] = resize(img, source.size, source.size)
		}
		if err := writeMiningSheets(gfxDir, source.name, srcs); err != nil {
			return err
		}
	}
	return nil
}

func writeMiningSheets(
	gfxDir string,
	name string,
	srcs map[string]*image.NRGBA,
) error {
	size := srcs[directions[0]].Bounds().Dx()
	bodySheet := newCanvas(size*frames, size*len(directions))
	shadowSheet := newCanvas(size*frames, size*len(directions))
	for row, direction := range directions {
		for frame := range frames {
			body := makeMiningFrame(srcs[direction], frame)
			pasteAt(bodySheet, body, frame*size, row*size)
			pasteAt(shadowSheet, makeFittedShadow(body), frame*size, row*size)
		}
	}

	if err := saveMiningSheet(gfxDir, name, bodySheet, false); err != nil {
		return err
	}
	return saveMiningSheet(gfxDir, name, shadowSheet, true)
}

func makeMiningFrame(src *image.NRGBA, frame int) *image.NRGBA {
	bodyAngle := miningSwingDegrees[frame] * math.Pi / 720
	return rotateImage(src, bodyAngle)
}

func rotateImage(src *image.NRGBA, radians float64) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	sin, cos := math.Sincos(radians)
	expandedW := max(w, int(math.Ceil(math.Abs(float64(w)*cos)+math.Abs(float64(h)*sin))))
	expandedH := max(h, int(math.Ceil(math.Abs(float64(h)*cos)+math.Abs(float64(w)*sin))))
	rotated := newCanvas(expandedW, expandedH)
	srcCX := float64(w-1) / 2
	srcCY := float64(h-1) / 2
	dstCX := float64(expandedW-1) / 2
	dstCY := float64(expandedH-1) / 2
	for y := range expandedH {
		for x := range expandedW {
			dx := float64(x) - dstCX
			dy := float64(y) - dstCY
			sx := int(math.Round(cos*dx + sin*dy + srcCX))
			sy := int(math.Round(-sin*dx + cos*dy + srcCY))
			if sx >= 0 && sx < w && sy >= 0 && sy < h {
				rotated.SetNRGBA(x, y, src.NRGBAAt(sx, sy))
			}
		}
	}

	bounds, ok := alphaBounds(rotated)
	if !ok {
		return newCanvas(w, h)
	}
	trimmed := newCanvas(bounds.Dx(), bounds.Dy())
	pasteAt(trimmed, rotated.SubImage(bounds), 0, 0)
	margin := max(1, min(w, h)/128)
	maxWidth := max(1, w-2*margin)
	maxHeight := max(1, h-2*margin)
	targetW := bounds.Dx()
	targetH := bounds.Dy()
	if targetW > maxWidth || targetH > maxHeight {
		scale := min(
			float64(maxWidth)/float64(targetW),
			float64(maxHeight)/float64(targetH),
		)
		targetW = max(1, int(math.Round(float64(targetW)*scale)))
		targetH = max(1, int(math.Round(float64(targetH)*scale)))
		trimmed = resize(trimmed, targetW, targetH)
	}

	centerX := float64(bounds.Min.X+bounds.Max.X)/2 - float64(expandedW-w)/2
	centerY := float64(bounds.Min.Y+bounds.Max.Y)/2 - float64(expandedH-h)/2
	x := int(math.Round(centerX - float64(targetW)/2))
	y := int(math.Round(centerY - float64(targetH)/2))
	x = max(margin, min(x, w-margin-targetW))
	y = max(margin, min(y, h-margin-targetH))
	out := newCanvas(w, h)
	pasteAt(out, trimmed, x, y)
	return out
}

func saveMiningSheet(gfxDir, name string, sheet *image.NRGBA, shadow bool) error {
	suffix := ""
	kind := "sheet"
	if shadow {
		suffix = "-shadow"
		kind = "shadow sheet"
	}
	out := filepath.Join(gfxDir, name+"-mining"+suffix+".png")
	if err := savePNG(out, sheet); err != nil {
		return fmt.Errorf("save %s mining %s: %w", name, kind, err)
	}
	slog.Info("wrote sheet", "path", out, "width", sheet.Bounds().Dx(), "height", sheet.Bounds().Dy())
	return nil
}
