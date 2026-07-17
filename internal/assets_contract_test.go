package gopher

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratedAssetDimensions(t *testing.T) {
	t.Parallel()
	assets := map[string]image.Point{
		"gopher-8dir.png":                              {X: 256, Y: 2048},
		"gopher-shadow-8dir.png":                       {X: 256, Y: 2048},
		"gopher-running.png":                           {X: 2048, Y: 2048},
		"gopher-shadow-running.png":                    {X: 2048, Y: 2048},
		"gopher-running-with-gun-1.png":                {X: 2048, Y: 4096},
		"gopher-running-with-gun-1-shadow.png":         {X: 2048, Y: 4096},
		"gopher-running-with-gun-2.png":                {X: 2048, Y: 512},
		"gopher-running-with-gun-2-shadow.png":         {X: 2048, Y: 512},
		"gopher-running-with-gun-flipped-1-shadow.png": {X: 2048, Y: 4096},
		"gopher-running-with-gun-flipped-2-shadow.png": {X: 2048, Y: 512},
		"gopher-idle-with-gun.png":                     {X: 256, Y: 2048},
		"gopher-idle-with-gun-shadow.png":              {X: 256, Y: 2048},
		"gopher-mining.png":                            {X: 2048, Y: 2048},
		"gopher-mining-shadow.png":                     {X: 2048, Y: 2048},
		"gopher-corpse.png":                            {X: 512, Y: 256},
		"gopher-corpse-shadow.png":                     {X: 512, Y: 256},
		"knight-idle.png":                              {X: 256, Y: 2048},
		"knight-idle-shadow.png":                       {X: 256, Y: 2048},
		"knight-idle-with-gun.png":                     {X: 256, Y: 2048},
		"knight-idle-with-gun-shadow.png":              {X: 256, Y: 2048},
		"knight-running.png":                           {X: 2048, Y: 2048},
		"knight-running-shadow.png":                    {X: 2048, Y: 2048},
		"knight-running-with-gun-1.png":                {X: 2048, Y: 4096},
		"knight-running-with-gun-1-shadow.png":         {X: 2048, Y: 4096},
		"knight-running-with-gun-2.png":                {X: 2048, Y: 512},
		"knight-running-with-gun-2-shadow.png":         {X: 2048, Y: 512},
		"knight-running-with-gun-flipped-1-shadow.png": {X: 2048, Y: 4096},
		"knight-running-with-gun-flipped-2-shadow.png": {X: 2048, Y: 512},
		"knight-flying-with-gun-1.png":                 {X: 1280, Y: 4096},
		"knight-flying-with-gun-1-shadow.png":          {X: 1280, Y: 4096},
		"knight-flying-with-gun-2.png":                 {X: 1280, Y: 512},
		"knight-flying-with-gun-2-shadow.png":          {X: 1280, Y: 512},
		"knight-mining.png":                            {X: 2048, Y: 2048},
		"knight-mining-shadow.png":                     {X: 2048, Y: 2048},
		"knight-take-off.png":                          {X: 4096, Y: 2048},
		"knight-take-off-shadow.png":                   {X: 4096, Y: 2048},
		"knight-hover.png":                             {X: 1280, Y: 2048},
		"knight-hover-shadow.png":                      {X: 1280, Y: 2048},
		"knight-corpse.png":                            {X: 512, Y: 256},
		"knight-corpse-shadow.png":                     {X: 512, Y: 256},
	}

	for name, want := range assets {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(DefaultGfxDir(), name)
			file, err := openRootedFile(path)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, file.Close()) })
			config, err := png.DecodeConfig(file)
			require.NoError(t, err)
			assert.Equal(t, want, image.Pt(config.Width, config.Height))
			assert.LessOrEqual(t, config.Width, 4096)
			assert.LessOrEqual(t, config.Height, 4096)
		})
	}
}

func TestGeneratedAssetDecodedPixelBudget(t *testing.T) {
	t.Parallel()
	inputs := make(map[string]bool, len(directions)*2)
	for _, direction := range directions {
		inputs["gopher-"+direction+".png"] = true
		inputs["knight-"+direction+".png"] = true
	}
	entries, err := os.ReadDir(DefaultGfxDir())
	require.NoError(t, err)

	var decodedBytes int64
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" || inputs[entry.Name()] {
			continue
		}
		file, openErr := openRootedFile(filepath.Join(DefaultGfxDir(), entry.Name()))
		require.NoError(t, openErr)
		config, decodeErr := png.DecodeConfig(file)
		require.NoError(t, decodeErr)
		require.NoError(t, file.Close())
		decodedBytes += int64(config.Width) * int64(config.Height) * 4
	}

	const maxDecodedBytes = int64(512 * 1024 * 1024)
	assert.LessOrEqual(
		t,
		decodedBytes,
		maxDecodedBytes,
		"runtime sprites decode to %.1f MiB",
		float64(decodedBytes)/(1024*1024),
	)
}
