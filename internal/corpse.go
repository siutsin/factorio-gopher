// Character corpse sheets generated as two fallen poses from the canonical
// south-facing sprites.

package gopher

import (
	"fmt"
	"image"
	"log/slog"
	"path/filepath"
)

const corpseFrames = 2

type corpseSource struct {
	name string
	path string
	size int
}

// Corpse writes the normal and mech-armour death animations and shadows.
func Corpse(gfxDir string) error {
	sources := []corpseSource{
		{name: "gopher", path: spritePath(gfxDir, "s"), size: runtimeFrameSize()},
		{name: "knight", path: knightPath(gfxDir, "s"), size: runtimeFrameSize()},
	}
	for _, source := range sources {
		img, err := loadPNG(source.path)
		if err != nil {
			return fmt.Errorf("load %s corpse source: %w", source.name, err)
		}
		if img.Bounds().Dx() != frameSize || img.Bounds().Dy() != frameSize {
			return fmt.Errorf(
				"%s corpse source dimensions are %dx%d; want %dx%d",
				source.name,
				img.Bounds().Dx(),
				img.Bounds().Dy(),
				frameSize,
				frameSize,
			)
		}
		if err := writeCorpseSheets(gfxDir, source.name, resize(img, source.size, source.size)); err != nil {
			return err
		}
	}
	return nil
}

func writeCorpseSheets(gfxDir, name string, src *image.NRGBA) error {
	size := src.Bounds().Dx()
	bodySheet := newCanvas(size*corpseFrames, size)
	shadowSheet := newCanvas(size*corpseFrames, size)
	for frame := range corpseFrames {
		body := makeCorpseFrame(src, frame == corpseFrames-1)
		pasteAt(bodySheet, body, frame*size, 0)
		pasteAt(shadowSheet, makeFittedShadow(body), frame*size, 0)
	}

	if err := saveCorpseSheet(gfxDir, name, bodySheet, false); err != nil {
		return err
	}
	return saveCorpseSheet(gfxDir, name, shadowSheet, true)
}

func makeCorpseFrame(src *image.NRGBA, clockwise bool) *image.NRGBA {
	size := src.Bounds().Dx()
	out := newCanvas(size, size)
	body := resize(rotateQuarterTurn(src, clockwise), max(1, size*9/10), max(1, size*9/10))

	bounds, ok := alphaBounds(body)
	if !ok {
		return out
	}
	pasteAt(out, body.SubImage(bounds), (size-bounds.Dx())/2, size-bounds.Dy())
	return out
}

func rotateQuarterTurn(src *image.NRGBA, clockwise bool) *image.NRGBA {
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	out := newCanvas(h, w)
	for y := range h {
		for x := range w {
			if clockwise {
				out.SetNRGBA(h-1-y, x, src.NRGBAAt(x, y))
			} else {
				out.SetNRGBA(y, w-1-x, src.NRGBAAt(x, y))
			}
		}
	}
	return out
}

func saveCorpseSheet(gfxDir, name string, sheet *image.NRGBA, shadow bool) error {
	suffix := ""
	kind := "sheet"
	if shadow {
		suffix = "-shadow"
		kind = "shadow sheet"
	}
	out := filepath.Join(gfxDir, name+"-corpse"+suffix+".png")
	if err := savePNG(out, sheet); err != nil {
		return fmt.Errorf("save %s corpse %s: %w", name, kind, err)
	}
	slog.Info("wrote sheet", "path", out, "width", sheet.Bounds().Dx(), "height", sheet.Bounds().Dy())
	return nil
}
