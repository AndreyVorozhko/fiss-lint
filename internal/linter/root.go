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

// isRegularOrSymlinkFile checks if the directory entry is a non-directory file or a symlink pointing to a file.
func isRegularOrSymlinkFile(dirPath string, entry os.DirEntry) bool {
	if !entry.IsDir() {
		if entry.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(filepath.Join(dirPath, entry.Name()))
			return err == nil && !info.IsDir()
		}
		return true
	}
	return false
}

// checkMandatoryFiles verifies rule FISS-R002: The FISS/ directory MUST contain INDEX.md and BOOTSTRAP.md.
// Returns whether INDEX.md and BOOTSTRAP.md exist as files with exact case sensitivity.
func checkMandatoryFiles(projectRoot string, report *model.Report) (hasIndex bool, hasBootstrap bool, err error) {
	fissDir := filepath.Join(projectRoot, "FISS")
	entries, err := os.ReadDir(fissDir)
	if err != nil {
		return false, false, fmt.Errorf("reading FISS directory: %w", err)
	}

	for _, entry := range entries {
		switch entry.Name() {
		case "INDEX.md":
			if isRegularOrSymlinkFile(fissDir, entry) {
				hasIndex = true
			}
		case "BOOTSTRAP.md":
			if isRegularOrSymlinkFile(fissDir, entry) {
				hasBootstrap = true
			}
		}
	}

	if !hasIndex {
		report.Add(model.Issue{
			RuleID:   "FISS-R002",
			Severity: model.SeverityError,
			FilePath: "FISS",
			Line:     0,
			Message:  "required file INDEX.md not found",
		})
	}

	if !hasBootstrap {
		report.Add(model.Issue{
			RuleID:   "FISS-R002",
			Severity: model.SeverityError,
			FilePath: "FISS",
			Line:     0,
			Message:  "required file BOOTSTRAP.md not found",
		})
	}

	return hasIndex, hasBootstrap, nil
}
