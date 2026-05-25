package gopher

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpritePath checks the simple "<dir>/gopher-<code>.png" join.
func TestSpritePath(t *testing.T) {
	assert.Equal(t, filepath.Join("/tmp/gfx", "gopher-ne.png"), spritePath("/tmp/gfx", "ne"))
}

// TestGunMappingShape guards the 18→8 mapping invariant: 18 entries, each
// pointing at a real direction code.
func TestGunMappingShape(t *testing.T) {
	require.Len(t, gunMapping, 18)
	dirs := make(map[string]bool, len(directions))
	for _, d := range directions {
		dirs[d] = true
	}
	for i, d := range gunMapping {
		assert.True(t, dirs[d], "gunMapping[%d]=%q not in directions", i, d)
	}
}
