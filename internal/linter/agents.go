package linter

import (
	"fmt"
	"os"
	"path/filepath"
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
