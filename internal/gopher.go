// Package gopher contains the procedural sprite-sheet generators and
// low-level NRGBA PNG helpers for the gopher Factorio mod. The Run/Shadow/
// Sheets entry points are invoked from cmd to regenerate the per-mod
// sheets under mod/graphics.
package gopher

import "path/filepath"

// frameSize is the per-direction sprite edge in pixels. Defined as a var (not
// const) so tests can shrink it via SetFrameSize for fast integration runs;
// production callers never mutate it.
var frameSize = 1024

// directions in Factorio's clockwise enum order, starting from N.
var directions = []string{"n", "ne", "e", "se", "s", "sw", "w", "nw"}

// gunMapping maps each of running_with_gun's 18 frames (20° apart starting
// at 0°/N, clockwise) to its nearest 45° direction.
var gunMapping = []string{
	"n", "n", "ne", "ne", "e", "e",
	"se", "se", "s", "s", "s", "sw",
	"sw", "w", "w", "nw", "nw", "n",
}

// spritePath returns the path to the per-direction PNG for code d under
// gfxDir (typically <repo>/mod/graphics).
func spritePath(gfxDir, d string) string {
	return filepath.Join(gfxDir, "gopher-"+d+".png")
}
