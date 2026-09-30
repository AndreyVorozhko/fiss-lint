package linter

import (
	"bufio"
	"io"
	"regexp"
	"strings"
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
