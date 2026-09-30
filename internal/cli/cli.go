package cli

import (
	"io"

	"fiss-lint/internal/model"
)

// Run parses command-line arguments and executes the requested CLI action.
// It returns the process exit code.
func Run(args []string, stdout, stderr io.Writer, info model.BuildInfo) int {
	return 0
}
