// Idle and armed-running sheets built from the per-direction body sprites.

package gopher

import (
	"fmt"
	"image"
	"log/slog"
	"path/filepath"
)

// Sheets writes the idle and armed-running body/shadow sheets to gfxDir.
func Sheets(gfxDir string) error {
	srcs := make(map[string]*image.NRGBA, len(directions))
	for _, d := range directions {
		img, err := loadPNG(spritePath(gfxDir, d))
		if err != nil {
			return fmt.Errorf("load %s: %w", d, err)
		}
		srcs[d] = img
	}

	if err := writeSheetIdle(gfxDir, srcs); err != nil {
		return err
	}
	return writeSheetGun(gfxDir, srcs)
}

// writeSheetIdle stitches the eight per-direction body sprites into a
// single-column 8-frame sheet for the idle animation.
func writeSheetIdle(gfxDir string, srcs map[string]*image.NRGBA) error {
	size := runtimeFrameSize()
	sheet := newCanvas(size, size*len(directions))
	for i, d := range directions {
		pasteAt(sheet, resize(srcs[d], size, size), 0, size*i)
	}

	out := filepath.Join(gfxDir, "gopher-8dir.png")
	if err := savePNG(out, sheet); err != nil {
		return err
	}
	slog.Info("wrote sheet", "path", out, "width", size, "height", size*len(directions))
	return nil
}

// writeSheetGun emits eight-frame run cycles across two direction stripes.
func writeSheetGun(gfxDir string, srcs map[string]*image.NRGBA) error {
	size := runtimeFrameSize()
	return writeArmedRunning(
		gfxDir,
		"gopher",
		size,
		func(direction string, aim, frame int) *image.NRGBA {
			return makeGopherArmedRunFrame(srcs[direction], aim, frame, size)
		},
		func(direction string, aim int) *image.NRGBA {
			return resize(makeGopherArmedSource(srcs[direction], aim), size, size)
		},
	)
}

func makeGopherArmedRunFrame(src *image.NRGBA, aim, frame, size int) *image.NRGBA {
	body := makeGopherArmedSource(src, aim)
	return resize(makeRunFrame(body, runBob(frame), frame), size, size)
}

func makeGopherArmedSource(src *image.NRGBA, aim int) *image.NRGBA {
	body := clone(src)
	drawArmedGun(body, aim)
	return body
}
