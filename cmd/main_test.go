package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMainSubprocess covers main() via the subprocess pattern: the test
// re-invokes the test binary with BUILD_TEST_MAIN=1 set, which trips the
// guard below to call main() with no args (exits 2 via gopher.CLI). The
// outer test asserts the subprocess exit code.
func TestMainSubprocess(t *testing.T) {
	if os.Getenv("BUILD_TEST_MAIN") == "1" {
		os.Args = []string{"build"}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainSubprocess$") //nolint:gosec // os.Args[0] is the test binary itself, not user input
	cmd.Env = append(os.Environ(), "BUILD_TEST_MAIN=1")
	err := cmd.Run()
	var ee *exec.ExitError
	require.ErrorAs(t, err, &ee, "expected subprocess to exit non-zero")
	assert.Equal(t, 2, ee.ExitCode(), "main should exit with gopher.CLI's return value")
}
