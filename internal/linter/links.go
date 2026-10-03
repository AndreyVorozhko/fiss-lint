package linter

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"fiss-lint/internal/model"
)

// linkRegex matches Markdown links of the form [text](target) or [text](target "title").
var linkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)(?:\s+[^)]*)?\)`)

// isBootstrapTarget returns true if the link target points to BOOTSTRAP.md.
func isBootstrapTarget(target string) bool {
	if idx := strings.IndexAny(target, "#?"); idx != -1 {
		target = target[:idx]
	}
	target = strings.TrimSpace(target)
	cleaned := path.Clean(target)
	return cleaned == "BOOTSTRAP.md" || cleaned == "FISS/BOOTSTRAP.md"
}

// isStandardTarget returns true if the link target references the official FISS standard
// (official website or official GitHub mirror, including version branch links).
func isStandardTarget(target string) bool {
	t := strings.TrimSpace(target)
	return strings.Contains(t, "https://fiss.vorozhko.ru") ||
		strings.Contains(t, "http://fiss.vorozhko.ru") ||
		strings.Contains(t, "https://github.com/AndreyVorozhko/fiss") ||
		strings.Contains(t, "http://github.com/AndreyVorozhko/fiss")
}

// checkIndexLinks verifies rule FISS-R003 (link to BOOTSTRAP.md) and FISS-R011 (link to official standard).
func checkIndexLinks(projectRoot string, report *model.Report) error {
	indexPath := filepath.Join(projectRoot, "FISS", "INDEX.md")
	file, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("opening FISS/INDEX.md: %w", err)
	}
	defer file.Close()

	var hasBootstrapLink bool
	var hasStandardLink bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := linkRegex.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 2 {
				target := match[2]
				if isBootstrapTarget(target) {
					hasBootstrapLink = true
				}
				if isStandardTarget(target) {
					hasStandardLink = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading FISS/INDEX.md: %w", err)
	}

	if !hasBootstrapLink {
		report.Add(model.Issue{
			RuleID:   "FISS-R003",
			Severity: model.SeverityError,
			FilePath: "FISS/INDEX.md",
			Line:     0,
			Message:  "missing link to BOOTSTRAP.md",
		})
	}

	if !hasStandardLink {
		report.Add(model.Issue{
			RuleID:   "FISS-R011",
			Severity: model.SeverityWarning,
			FilePath: "FISS/INDEX.md",
			Line:     0,
			Message:  "recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found",
		})
	}

	return nil
}
