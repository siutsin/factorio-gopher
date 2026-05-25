package gopher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetFrameSizeClamp covers the size-clamping branch (size < 1024 / scale
// rounds to 0 → clamped to 1). Resets to 64 (the package's TestMain default)
// afterwards. Not parallel because it mutates package globals.
func TestSetFrameSizeClamp(t *testing.T) {
	t.Cleanup(func() { SetFrameSize(64) })
	SetFrameSize(1)
	assert.Equal(t, 1, frameSize)
	assert.GreaterOrEqual(t, bobAmp, 1, "bobAmp must clamp to >= 1")
	assert.GreaterOrEqual(t, footLift, 1)
	assert.GreaterOrEqual(t, armLift, 1)
}
