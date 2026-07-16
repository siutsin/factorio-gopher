package main

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		slog.SetDefault(slog.New(slog.DiscardHandler))
		t.Cleanup(func() { slog.SetDefault(previous) })

		exitCode := run(nil, func([]string) (string, int, error) {
			return "", 2, errors.New("boom")
		})
		assert.Equal(t, 2, exitCode)
	})
}

func TestMain(t *testing.T) {
	previousArgs := os.Args
	previousExit := exitProcess
	previousLogger := slog.Default()
	t.Cleanup(func() {
		os.Args = previousArgs
		exitProcess = previousExit
		slog.SetDefault(previousLogger)
	})

	os.Args = []string{"publish"}
	t.Setenv("FACTORIO_MOD_PORTAL_API_KEY", "")
	slog.SetDefault(slog.New(slog.DiscardHandler))
	exitCode := -1
	exitProcess = func(code int) { exitCode = code }

	main()

	assert.Equal(t, 1, exitCode)
}
