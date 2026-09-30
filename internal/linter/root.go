package linter

import (
	"fmt"
	"os"
	"path/filepath"

	"fiss-lint/internal/model"
)

// checkRootDir verifies rule FISS-R001: The project root MUST contain the FISS/ directory.
// Returns true if FISS/ directory is found with exact case sensitivity, false otherwise.
func checkRootDir(projectRoot string, report *model.Report) (bool, error) {
	entries, err := os.ReadDir(projectRoot)
	if err != nil {
		return false, fmt.Errorf("reading project root: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == "FISS" {
			if entry.IsDir() {
				return true, nil
			}
			// If it's a symlink, check whether the target is a directory
			if entry.Type()&os.ModeSymlink != 0 {
				info, statErr := os.Stat(filepath.Join(projectRoot, entry.Name()))
				if statErr == nil && info.IsDir() {
					return true, nil
				}
			}
		}
	}

	report.Add(model.Issue{
		RuleID:   "FISS-R001",
		Severity: model.SeverityError,
		FilePath: ".",
		Line:     0,
		Message:  "directory FISS/ not found in project root",
	})
	return false, nil
}
