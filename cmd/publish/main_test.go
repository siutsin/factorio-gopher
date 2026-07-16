package main

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		exitCode := run(nil, func([]string) (string, int, error) {
			return "published gopher 0.0.3", 0, nil
		})
		assert.Equal(t, 0, exitCode)
	})

	t.Run("error", func(t *testing.T) {
		previous := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		t.Cleanup(func() { slog.SetDefault(previous) })

		exitCode := run(nil, func([]string) (string, int, error) {
			return "", 2, errors.New("boom")
		})
		assert.Equal(t, 2, exitCode)
	})
}

func TestMainSubprocess(t *testing.T) {
	if os.Getenv("PUBLISH_TEST_MAIN") == "1" {
		os.Args = []string{"publish"}
		require.NoError(t, os.Unsetenv("FACTORIO_MOD_PORTAL_API_KEY"))
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainSubprocess$") //nolint:gosec // test binary path
	cmd.Env = append(os.Environ(), "PUBLISH_TEST_MAIN=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
}
