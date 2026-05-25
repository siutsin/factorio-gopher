package gopher

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestShadow covers the success path and the per-output save-error branch
// for each of the three shadow sheets, plus the missing-source error.
func TestShadow(t *testing.T) {
	outputs := []string{
		"gopher-shadow-8dir.png",
		"gopher-shadow-running.png",
		"gopher-shadow-running-with-gun.png",
	}

	cases := []struct {
		name      string
		blockSave string
		wantErr   string
	}{
		{name: "success"},
		{name: "missing source", wantErr: "load"},
		{name: "save error - 8dir", blockSave: "gopher-shadow-8dir.png"},
		{name: "save error - running", blockSave: "gopher-shadow-running.png"},
		{name: "save error - gun", blockSave: "gopher-shadow-running-with-gun.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runGenerator(t, Shadow, tc.wantErr, tc.blockSave, func(dir string) {
				for _, name := range outputs {
					_, err := loadPNG(filepath.Join(dir, name))
					assert.NoError(t, err, "missing %s", name)
				}
			})
		})
	}
}
