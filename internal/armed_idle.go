// Dedicated armed-idle sheets generated from the canonical direction sprites.

package gopher

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"math"
	"path/filepath"
)

type armedIdleSource struct {
	name    string
	path    func(gfxDir, direction string) string
	makeRow func(src *image.NRGBA, direction int) *image.NRGBA
}

// ArmedIdle writes dedicated normal and mech-armour armed-idle sheets.
func ArmedIdle(gfxDir string) error {
	size := runtimeFrameSize()
	sources := []armedIdleSource{
		{
			name: "gopher",
			path: spritePath,
			makeRow: func(src *image.NRGBA, direction int) *image.NRGBA {
				body := resize(src, size, size)
				drawGun(body, direction, len(directions))
				return body
			},
		},
		{
			name: "knight",
			path: knightPath,
			makeRow: func(src *image.NRGBA, _ int) *image.NRGBA {
				body := resize(src, size, size)
				return makeFlightBody(body, flightPose{scaleX: 1.01, scaleY: 0.99})
			},
		},
	}
	for _, source := range sources {
		srcs := make(map[string]*image.NRGBA, len(directions))
		for _, direction := range directions {
			img, err := loadPNG(source.path(gfxDir, direction))
			if err != nil {
				return fmt.Errorf("load %s-%s armed-idle source: %w", source.name, direction, err)
			}
			if img.Bounds().Dx() != frameSize || img.Bounds().Dy() != frameSize {
				return fmt.Errorf(
					"%s-%s armed-idle source dimensions are %dx%d; want %dx%d",
					source.name,
					direction,
					img.Bounds().Dx(),
					img.Bounds().Dy(),
					frameSize,
					frameSize,
				)
			}
			srcs[direction] = img
		}
		if err := writeArmedIdleSheets(gfxDir, source.name, srcs, source.makeRow); err != nil {
			return err
		}
	}
	return nil
}

func writeArmedIdleSheets(
	gfxDir string,
	name string,
	srcs map[string]*image.NRGBA,
	makeRow func(src *image.NRGBA, direction int) *image.NRGBA,
) error {
	size := runtimeFrameSize()
	bodySheet := newCanvas(size, size*len(directions))
	shadowSheet := newCanvas(size, size*len(directions))
	for row, direction := range directions {
		body := makeRow(srcs[direction], row)
		pasteAt(bodySheet, body, 0, row*size)
		pasteAt(shadowSheet, makeFittedShadow(body), 0, row*size)
	}
	if err := saveArmedIdleSheet(gfxDir, name, bodySheet, false); err != nil {
		return err
	}
	return saveArmedIdleSheet(gfxDir, name, shadowSheet, true)
}

func drawGun(img *image.NRGBA, direction, directionCount int) {
	angle := -math.Pi/2 + float64(direction)*2*math.Pi/float64(directionCount)
	drawGunAtAngle(img, angle)
}

func drawArmedGun(img *image.NRGBA, direction int) {
	drawGunAtAngle(img, armedGunAngle(direction))
}

func armedGunAngle(direction int) float64 {
	sweep := math.Pi * float64(direction) / float64(len(gunMapping)-1)
	return -math.Pi/2 + sweep
}

func drawGunAtAngle(img *image.NRGBA, angle float64) {
	size := img.Bounds().Dx()
	x0 := size / 2
	y0 := size * 3 / 5
	barrelLength := size / 4
	x1 := x0 + int(math.Round(math.Cos(angle)*float64(barrelLength)))
	y1 := y0 + int(math.Round(math.Sin(angle)*float64(barrelLength)))
	metal := color.NRGBA{R: 55, G: 71, B: 79, A: 255}
	stock := color.NRGBA{R: 111, G: 78, B: 55, A: 255}
	drawThickLine(img, x0, y0, x1, y1, max(1, size/48), metal)
	stockLength := size / 10
	x2 := x0 - int(math.Round(math.Cos(angle)*float64(stockLength)))
	y2 := y0 - int(math.Round(math.Sin(angle)*float64(stockLength)))
	drawThickLine(img, x0, y0, x2, y2, max(1, size/40), stock)
	gripAngle := angle + math.Pi/2
	x3 := x0 + int(math.Round(math.Cos(gripAngle)*float64(size/12)))
	y3 := y0 + int(math.Round(math.Sin(gripAngle)*float64(size/12)))
	drawThickLine(img, x0, y0, x3, y3, max(1, size/64), stock)
}

func saveArmedIdleSheet(gfxDir, name string, sheet *image.NRGBA, shadow bool) error {
	suffix := ""
	kind := "sheet"
	if shadow {
		suffix = "-shadow"
		kind = "shadow sheet"
	}
	out := filepath.Join(gfxDir, name+"-idle-with-gun"+suffix+".png")
	if err := savePNG(out, sheet); err != nil {
		return fmt.Errorf("save %s armed-idle %s: %w", name, kind, err)
	}
	slog.Info("wrote sheet", "path", out, "width", sheet.Bounds().Dx(), "height", sheet.Bounds().Dy())
	return nil
}
