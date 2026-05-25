// Idle 8-direction sheet and 18-frame running_with_gun sheet stitched from
// the per-direction body sprites.

package gopher

import (
	"fmt"
	"image"
	"log/slog"
	"path/filepath"
)

// Sheets writes the idle 8-direction sheet and the running_with_gun
// composite sheet to gfxDir.
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
	sheet := newCanvas(frameSize, frameSize*len(directions))
	for i, d := range directions {
		pasteAt(sheet, srcs[d], 0, frameSize*i)
	}

	out := filepath.Join(gfxDir, "gopher-8dir.png")
	if err := savePNG(out, sheet); err != nil {
		return err
	}
	slog.Info("wrote sheet", "path", out, "width", frameSize, "height", frameSize*len(directions))
	return nil
}

// writeSheetGun lays the 18 gun-frame directions into a 6×3 grid sheet
// matching Factorio's running_with_gun layout.
func writeSheetGun(gfxDir string, srcs map[string]*image.NRGBA) error {
	const cols, rows = 6, 3
	sheet := newCanvas(frameSize*cols, frameSize*rows)
	for i, d := range gunMapping {
		c := i % cols
		r := i / cols
		pasteAt(sheet, srcs[d], frameSize*c, frameSize*r)
	}

	out := filepath.Join(gfxDir, "gopher-running-with-gun.png")
	if err := savePNG(out, sheet); err != nil {
		return err
	}
	slog.Info("wrote sheet", "path", out, "width", frameSize*cols, "height", frameSize*rows)
	return nil
}
