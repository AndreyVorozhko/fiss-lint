package linter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fiss-lint/internal/model"
)

// navItemRegex matches a Markdown list item containing a link: - [Title](target)
var navItemRegex = regexp.MustCompile(`^\s*-\s+\[([^\]]+)\]\(([^)]+)\)`)

// NavEntry represents a single navigation entry in an INDEX.md file.
type NavEntry struct {
	Line        int    // 1-indexed line number of the "- [Title](target)" link line
	Title       string // Link title inside [...]
	Target      string // Target URL or path inside (...)
	RawLine     string // Raw text of the link line

	HasNextLine bool   // Whether there was a subsequent line in the file
	NextLineNum int    // Line number of the subsequent line (Line + 1)
	NextLineRaw string // Raw text of the subsequent line
}

// parseIndexNavEntries reads markdown content and extracts all list item navigation entries
// along with their immediately subsequent lines.
func parseIndexNavEntries(reader io.Reader) ([]NavEntry, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var entries []NavEntry
	for i, line := range lines {
		matches := navItemRegex.FindStringSubmatch(line)
		if matches != nil {
			title := matches[1]
			target := strings.TrimSpace(matches[2])
			if parts := strings.Fields(target); len(parts) > 0 {
				target = parts[0]
			}

			entry := NavEntry{
				Line:    i + 1,
				Title:   title,
				Target:  target,
				RawLine: line,
			}

			if i+1 < len(lines) {
				entry.HasNextLine = true
				entry.NextLineNum = i + 2
				entry.NextLineRaw = lines[i+1]
			}

			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// validateNavEntries checks navigation entries against rules FISS-R005 and FISS-R004.
func validateNavEntries(filePath string, entries []NavEntry, isRootIndex bool, report *model.Report) {
	for _, entry := range entries {
		hasValidCondition := false

		if !entry.HasNextLine {
			report.Add(model.Issue{
				RuleID:   "FISS-R005",
				Severity: model.SeverityError,
				FilePath: filePath,
				Line:     entry.Line,
				Message:  "missing 'Read when:' condition for navigation entry",
			})
		} else {
			trimmedNext := strings.TrimSpace(entry.NextLineRaw)
			if trimmedNext == "" || strings.HasPrefix(trimmedNext, "- [") {
				report.Add(model.Issue{
					RuleID:   "FISS-R005",
					Severity: model.SeverityError,
					FilePath: filePath,
					Line:     entry.Line,
					Message:  "missing 'Read when:' condition for navigation entry",
				})
			} else if !strings.HasPrefix(entry.NextLineRaw, "  ") || strings.HasPrefix(entry.NextLineRaw, "   ") {
				report.Add(model.Issue{
					RuleID:   "FISS-R005",
					Severity: model.SeverityError,
					FilePath: filePath,
					Line:     entry.NextLineNum,
					Message:  "invalid indentation for 'Read when:' condition (expected exactly 2 spaces)",
				})
			} else {
				contentAfterIndent := entry.NextLineRaw[2:]
				const marker = "Read when:"
				if !strings.HasPrefix(contentAfterIndent, marker) {
					report.Add(model.Issue{
						RuleID:   "FISS-R005",
						Severity: model.SeverityError,
						FilePath: filePath,
						Line:     entry.NextLineNum,
						Message:  "invalid read condition marker (must be exact 'Read when:')",
					})
				} else {
					condition := strings.TrimSpace(contentAfterIndent[len(marker):])
					if condition == "" {
						report.Add(model.Issue{
							RuleID:   "FISS-R005",
							Severity: model.SeverityError,
							FilePath: filePath,
							Line:     entry.NextLineNum,
							Message:  "empty condition text in 'Read when:'",
						})
					} else {
						hasValidCondition = true
					}
				}
			}
		}

		if isRootIndex && isBootstrapTarget(entry.Target) && !hasValidCondition {
			report.Add(model.Issue{
				RuleID:   "FISS-R004",
				Severity: model.SeverityError,
				FilePath: filePath,
				Line:     entry.Line,
				Message:  "link to BOOTSTRAP.md must have an attached read condition",
			})
		}
	}
}

// validateAllIndexes walks the FISS/ directory and validates all INDEX.md files.
func validateAllIndexes(projectRoot string, report *model.Report) error {
	fissDir := filepath.Join(projectRoot, "FISS")
	return filepath.WalkDir(fissDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "INDEX.md" {
			relPath, relErr := filepath.Rel(projectRoot, path)
			if relErr != nil {
				relPath = path
			}
			relPath = filepath.ToSlash(relPath)
			isRootIndex := (relPath == "FISS/INDEX.md")

			file, openErr := os.Open(path)
			if openErr != nil {
				return fmt.Errorf("opening %s: %w", relPath, openErr)
			}

			entries, parseErr := parseIndexNavEntries(file)
			file.Close()
			if parseErr != nil {
				return fmt.Errorf("parsing %s: %w", relPath, parseErr)
			}

			validateNavEntries(relPath, entries, isRootIndex, report)
		}
		return nil
	})
}
