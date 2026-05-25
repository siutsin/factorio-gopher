package gopher

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSheets covers the success path, the missing-source error, and the
// per-output save-error branch for each of the two stitched sheets.
func TestSheets(t *testing.T) {
	outputs := []string{
		"gopher-8dir.png",
		"gopher-running-with-gun.png",
	}

	cases := []struct {
		name      string
		blockSave string
		wantErr   string
	}{
		{name: "success"},
		{name: "missing source", wantErr: "load"},
		{name: "save error - 8dir", blockSave: "gopher-8dir.png"},
		{name: "save error - gun", blockSave: "gopher-running-with-gun.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runGenerator(t, Sheets, tc.wantErr, tc.blockSave, func(dir string) {
				for _, name := range outputs {
					_, err := loadPNG(filepath.Join(dir, name))
					assert.NoError(t, err, "missing %s", name)
				}
			})
		})
	}
}
