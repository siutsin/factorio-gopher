// Package gopher contains the procedural sprite-sheet generators and
// low-level NRGBA PNG helpers for the gopher Factorio mod. The Run/Shadow/
// Sheets entry points are invoked from cmd to regenerate the per-mod
// sheets under mod/graphics.
package gopher

import "path/filepath"

// frameSize is the canonical per-direction source edge in pixels. Defined as
// a var so tests can shrink it via SetFrameSize; production callers never
// mutate it. Runtime sheets use quarter-size frames and compensate in Lua.
var frameSize = 1024

func runtimeFrameSize() int {
	return max(1, frameSize/4)
}

// directions in Factorio's clockwise enum order, starting from N.
var directions = []string{"n", "ne", "e", "se", "s", "sw", "w", "nw"}

// gunMapping follows Factorio's 18-row armed half-sweep from north through
// east to south. The engine mirrors these rows for west-facing combinations.
var gunMapping = []string{
	"n", "n", "n", "ne", "ne", "ne",
	"ne", "e", "e", "e", "e", "se",
	"se", "se", "se", "s", "s", "s",
}

// spritePath returns the path to the per-direction PNG for code d under
// gfxDir (typically <repo>/mod/graphics).
func spritePath(gfxDir, d string) string {
	return filepath.Join(gfxDir, "gopher-"+d+".png")
}
