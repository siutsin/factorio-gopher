package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	previousArgs := os.Args
	previousExit := exitProcess
	previousStderr := standardError
	t.Cleanup(func() {
		os.Args = previousArgs
		exitProcess = previousExit
		standardError = previousStderr
	})

	os.Args = []string{"build"}
	standardError = io.Discard
	exitCode := -1
	exitProcess = func(code int) { exitCode = code }

	main()

	assert.Equal(t, 2, exitCode)
}
