// Knight sprite sheets for the mech-armour ground and in-air animations.

package gopher

import (
	"fmt"
	"image"
	"log/slog"
	"math"
	"path/filepath"
)

const (
	takeOffFrames = 16
	hoverFrames   = 5
)

type flightPose struct {
	scaleX float64
	scaleY float64
	lift   float64
	height float64
}

var groundRunPoses = []flightPose{
	{scaleX: 1.000, scaleY: 1.000, lift: 0.000},
	{scaleX: 0.995, scaleY: 1.005, lift: 0.024},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.000},
	{scaleX: 1.005, scaleY: 0.995, lift: -0.016},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.000},
	{scaleX: 0.995, scaleY: 1.005, lift: 0.024},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.000},
	{scaleX: 1.005, scaleY: 0.995, lift: -0.016},
}

// takeOffPoses starts with a grounded pose, compresses for anticipation, then
// settles into the first hover pose. Landing plays these frames in reverse.
var takeOffPoses = []flightPose{
	{scaleX: 1.000, scaleY: 1.000, lift: 0.000, height: 0.00},
	{scaleX: 1.010, scaleY: 0.985, lift: 0.000, height: 0.00},
	{scaleX: 1.025, scaleY: 0.960, lift: 0.000, height: 0.00},
	{scaleX: 1.040, scaleY: 0.930, lift: 0.000, height: 0.00},
	{scaleX: 1.030, scaleY: 0.950, lift: 0.000, height: 0.04},
	{scaleX: 1.015, scaleY: 0.980, lift: 0.002, height: 0.12},
	{scaleX: 0.995, scaleY: 1.020, lift: 0.006, height: 0.25},
	{scaleX: 0.990, scaleY: 1.035, lift: 0.012, height: 0.42},
	{scaleX: 0.995, scaleY: 1.020, lift: 0.016, height: 0.60},
	{scaleX: 1.000, scaleY: 1.010, lift: 0.020, height: 0.76},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.023, height: 0.88},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.027, height: 0.95},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.031, height: 1.00},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.031, height: 1.00},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.031, height: 1.00},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.031, height: 1.00},
}

var hoverPoses = []flightPose{
	{scaleX: 1.000, scaleY: 1.000, lift: 0.031, height: 1.00},
	{scaleX: 0.998, scaleY: 1.006, lift: 0.039, height: 1.04},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.031, height: 1.00},
	{scaleX: 1.002, scaleY: 0.994, lift: 0.023, height: 0.96},
	{scaleX: 1.000, scaleY: 1.000, lift: 0.031, height: 1.00},
}

// Knight writes the mech-armour ground sheets, a 16-frame takeoff transition,
// and a five-frame hover loop. Lua reuses the takeoff sheet in reverse for
// landing.
func Knight(gfxDir string) error {
	srcs := make(map[string]*image.NRGBA, len(directions))
	for _, d := range directions {
		img, err := loadPNG(knightPath(gfxDir, d))
		if err != nil {
			return fmt.Errorf("load knight-%s: %w", d, err)
		}
		if img.Bounds().Dx() != frameSize || img.Bounds().Dy() != frameSize {
			return fmt.Errorf(
				"knight-%s dimensions are %dx%d; want %dx%d",
				d,
				img.Bounds().Dx(),
				img.Bounds().Dy(),
				frameSize,
				frameSize,
			)
		}
		srcs[d] = img
	}

	if err := writeKnightIdle(gfxDir, srcs); err != nil {
		return err
	}
	if err := writeKnightRunning(gfxDir, srcs); err != nil {
		return err
	}
	if err := writeKnightGun(gfxDir, srcs); err != nil {
		return err
	}
	if err := writeFlightAnimation(gfxDir, "take-off", srcs, takeOffPoses); err != nil {
		return err
	}
	return writeFlightAnimation(gfxDir, "hover", srcs, hoverPoses)
}

func writeKnightIdle(gfxDir string, srcs map[string]*image.NRGBA) error {
	size := flightFrameSize()
	bodySheet := newCanvas(size, size*len(directions))
	shadowSheet := newCanvas(size, size*len(directions))
	for row, d := range directions {
		body := resize(srcs[d], size, size)
		pasteAt(bodySheet, body, 0, row*size)
		pasteAt(shadowSheet, makeFittedShadow(body), 0, row*size)
	}
	return saveKnightSheets(gfxDir, "idle", bodySheet, shadowSheet)
}

func writeKnightRunning(gfxDir string, srcs map[string]*image.NRGBA) error {
	size := flightFrameSize()
	bodySheet := newCanvas(size*len(groundRunPoses), size*len(directions))
	shadowSheet := newCanvas(size*len(groundRunPoses), size*len(directions))
	for row, d := range directions {
		body := resize(srcs[d], size, size)
		shadow := makeFittedShadow(body)
		for column, pose := range groundRunPoses {
			pasteAt(bodySheet, makeFlightBody(body, pose), column*size, row*size)
			pasteAt(shadowSheet, shadow, column*size, row*size)
		}
	}
	return saveKnightSheets(gfxDir, "running", bodySheet, shadowSheet)
}

func writeKnightGun(gfxDir string, srcs map[string]*image.NRGBA) error {
	const columns, rows = 6, 3
	size := flightFrameSize()
	bodySheet := newCanvas(size*columns, size*rows)
	shadowSheet := newCanvas(size*columns, size*rows)
	for i, d := range gunMapping {
		column := i % columns
		row := i / columns
		body := resize(srcs[d], size, size)
		pasteAt(bodySheet, body, column*size, row*size)
		pasteAt(shadowSheet, makeFittedShadow(body), column*size, row*size)
	}
	return saveKnightSheets(gfxDir, "running-with-gun", bodySheet, shadowSheet)
}

func writeFlightAnimation(
	gfxDir string,
	name string,
	srcs map[string]*image.NRGBA,
	poses []flightPose,
) error {
	size := flightFrameSize()
	sheet := newCanvas(size*len(poses), size*len(directions))

	for row, d := range directions {
		base := resize(srcs[d], size, size)
		for column, pose := range poses {
			pasteAt(sheet, makeFlightBody(base, pose), column*size, row*size)
		}
	}
	if err := saveKnightSheet(gfxDir, name, sheet, false); err != nil {
		return err
	}

	clear(sheet.Pix)
	for row, d := range directions {
		baseShadow := makeFittedShadow(resize(srcs[d], size, size))
		for column, pose := range poses {
			pasteAt(sheet, makeFlightShadow(baseShadow, pose.height), column*size, row*size)
		}
	}
	return saveKnightSheet(gfxDir, name, sheet, true)
}

func saveKnightSheets(
	gfxDir string,
	name string,
	bodySheet *image.NRGBA,
	shadowSheet *image.NRGBA,
) error {
	if err := saveKnightSheet(gfxDir, name, bodySheet, false); err != nil {
		return err
	}
	return saveKnightSheet(gfxDir, name, shadowSheet, true)
}

func saveKnightSheet(gfxDir, name string, sheet *image.NRGBA, shadow bool) error {
	suffix := ""
	kind := "sheet"
	if shadow {
		suffix = "-shadow"
		kind = "shadow sheet"
	}
	out := filepath.Join(gfxDir, "knight-"+name+suffix+".png")
	if err := savePNG(out, sheet); err != nil {
		return fmt.Errorf("save knight %s %s: %w", name, kind, err)
	}
	slog.Info(
		"wrote sheet",
		"path", out,
		"width", sheet.Bounds().Dx(),
		"height", sheet.Bounds().Dy(),
	)
	return nil
}

func flightFrameSize() int {
	return max(1, frameSize/2)
}

func makeFlightBody(src *image.NRGBA, pose flightPose) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	targetW := max(1, int(math.Round(float64(w)*pose.scaleX)))
	targetH := max(1, int(math.Round(float64(h)*pose.scaleY)))
	transformed := resize(src, targetW, targetH)

	out := newCanvas(w, h)
	x := (w - targetW) / 2
	y := h - targetH - int(math.Round(float64(h)*pose.lift))
	pasteAt(out, transformed, x, y)
	return out
}

func makeFlightShadow(src *image.NRGBA, height float64) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	bounds, ok := alphaBounds(src)
	if !ok {
		return newCanvas(w, h)
	}

	widthScale := 1.0 - 0.25*height
	heightScale := 1.0 - 0.15*height
	targetW := max(1, int(math.Round(float64(bounds.Dx())*widthScale)))
	targetH := max(1, int(math.Round(float64(bounds.Dy())*heightScale)))
	trimmed := newCanvas(bounds.Dx(), bounds.Dy())
	pasteAt(trimmed, src.SubImage(bounds), 0, 0)
	transformed := resize(trimmed, targetW, targetH)
	alpha := uint8(math.Round(255 * (1.0 - 0.45*height)))
	scaleAlpha(transformed, alpha)

	out := newCanvas(w, h)
	x := bounds.Min.X + (bounds.Dx()-targetW)/2
	y := bounds.Min.Y + (bounds.Dy()-targetH)/2
	pasteAt(out, transformed, x, y)
	return out
}

func knightPath(gfxDir, d string) string {
	return filepath.Join(gfxDir, "knight-"+d+".png")
}
