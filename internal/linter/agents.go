package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fiss-lint/internal/model"
)

// knownAgentInstructionFiles contains root agent instruction files inspected by FISS-R010.
var knownAgentInstructionFiles = []string{"AGENTS.md", "CLAUDE.md", ".cursorrules"}

// findAgentInstructionFiles searches projectRoot for known agent instruction files.
// It verifies exact case matching against directory entries to ensure cross-platform
// consistency and checks that candidate entries are regular files or symlinks to files.
func findAgentInstructionFiles(projectRoot string) ([]string, error) {
	entries, err := os.ReadDir(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("reading project root: %w", err)
	}

	diskNames := make(map[string]bool, len(entries))
	for _, entry := range entries {
		diskNames[entry.Name()] = true
	}

	var found []string
	for _, candidate := range knownAgentInstructionFiles {
		if !diskNames[candidate] {
			continue
		}

		fullPath := filepath.Join(projectRoot, candidate)
		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("stat agent file %s: %w", candidate, err)
		}

		if !info.IsDir() {
			found = append(found, candidate)
		}
	}

	return found, nil
}

// directsToFissIndex checks whether content contains a link or directive targeting FISS/INDEX.md or FISS/.
func directsToFissIndex(content string) bool {
	return strings.Contains(content, "FISS/INDEX.md") ||
		strings.Contains(content, "FISS/index.md") ||
		strings.Contains(content, "FISS/") ||
		strings.Contains(content, "FISS\\INDEX.md") ||
		strings.Contains(content, "FISS\\")
}

// checkAgentInstructions checks that any root agent instruction files direct to FISS/INDEX.md.
// If any such file exists and does not reference FISS/INDEX.md or FISS/, it records a FISS-R010 error.
func checkAgentInstructions(projectRoot string, report *model.Report) error {
	agentFiles, err := findAgentInstructionFiles(projectRoot)
	if err != nil {
		return err
	}

	for _, filename := range agentFiles {
		filePath := filepath.Join(projectRoot, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading agent instruction file %s: %w", filename, err)
		}

		if !directsToFissIndex(string(content)) {
			report.Add(model.Issue{
				RuleID:   "FISS-R010",
				Severity: model.SeverityError,
				FilePath: filename,
				Line:     0,
				Message:  "does not direct to FISS/INDEX.md",
			})
		}
	}

	return nil
}
