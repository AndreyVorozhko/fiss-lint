package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fiss-lint/internal/model"
)

// NavigationGraph represents the traversed navigation graph of a FISS intellectual space.
type NavigationGraph struct {
	// ReachableFiles contains all relative paths (POSIX forward slash) from projectRoot reached via navigation.
	ReachableFiles map[string]bool
	// VisitedIndexes contains all INDEX.md files that were parsed during graph traversal.
	VisitedIndexes map[string]bool
}

// IsReachable checks whether the given relative path from projectRoot was reached via navigation.
func (g *NavigationGraph) IsReachable(relPath string) bool {
	if g == nil || g.ReachableFiles == nil {
		return false
	}
	return g.ReachableFiles[filepath.ToSlash(filepath.Clean(relPath))]
}

// buildNavigationGraph constructs the directed navigation graph starting from FISS/INDEX.md.
// It traverses all referenced indexes (BFS) while preventing cycles via visited tracking ([RISK-003]).
func buildNavigationGraph(projectRoot string) (*NavigationGraph, error) {
	graph := &NavigationGraph{
		ReachableFiles: make(map[string]bool),
		VisitedIndexes: make(map[string]bool),
	}

	rootIndexRel := "FISS/INDEX.md"
	exists, isDir, err := checkPathExistsCaseSensitive(projectRoot, rootIndexRel)
	if err != nil {
		return nil, fmt.Errorf("checking root index: %w", err)
	}
	if !exists || isDir {
		// Root index does not exist, return empty graph without error
		return graph, nil
	}

	// Root index itself is always reachable and forms the traversal root
	graph.ReachableFiles[rootIndexRel] = true
	graph.VisitedIndexes[rootIndexRel] = true

	queue := []string{rootIndexRel}

	// Track real filesystem paths to avoid infinite symlink loops ([RISK-003])
	visitedRealPaths := make(map[string]bool)
	fullRootPath := filepath.Join(projectRoot, filepath.FromSlash(rootIndexRel))
	if realRoot, evalErr := filepath.EvalSymlinks(fullRootPath); evalErr == nil {
		visitedRealPaths[realRoot] = true
	}

	for len(queue) > 0 {
		currentRel := queue[0]
		queue = queue[1:]

		fullCurrent := filepath.Join(projectRoot, filepath.FromSlash(currentRel))
		f, openErr := os.Open(fullCurrent)
		if openErr != nil {
			continue
		}

		entries, parseErr := parseIndexNavEntries(f)
		f.Close()
		if parseErr != nil {
			continue
		}

		for _, entry := range entries {
			if isExternalURL(entry.Target) {
				continue
			}

			resolved, resErr := resolveRelativeTarget(projectRoot, currentRel, entry.Target)
			if resErr != nil || resolved.IsSelfAnchor {
				continue
			}

			exists, isDir, checkErr := checkPathExistsCaseSensitive(projectRoot, resolved.RelPathFromRoot)
			if checkErr != nil || !exists {
				continue
			}

			if isDir {
				// Mark the directory itself as reached
				graph.ReachableFiles[resolved.RelPathFromRoot] = true

				indexInDirRel := filepath.ToSlash(filepath.Join(resolved.RelPathFromRoot, "INDEX.md"))
				idxExists, idxIsDir, _ := checkPathExistsCaseSensitive(projectRoot, indexInDirRel)
				if idxExists && !idxIsDir {
					graph.ReachableFiles[indexInDirRel] = true

					if !graph.VisitedIndexes[indexInDirRel] {
						graph.VisitedIndexes[indexInDirRel] = true

						fullIdx := filepath.Join(projectRoot, filepath.FromSlash(indexInDirRel))
						realIdx, evalErr := filepath.EvalSymlinks(fullIdx)
						if evalErr == nil {
							if visitedRealPaths[realIdx] {
								continue
							}
							visitedRealPaths[realIdx] = true
						}

						queue = append(queue, indexInDirRel)
					}
				}
				continue
			}

			// Target is a file
			if strings.HasSuffix(strings.ToLower(resolved.CleanTarget), ".md") {
				graph.ReachableFiles[resolved.RelPathFromRoot] = true

				// If target is an INDEX.md file, enqueue it for traversal
				if filepath.Base(filepath.FromSlash(resolved.RelPathFromRoot)) == "INDEX.md" {
					if !graph.VisitedIndexes[resolved.RelPathFromRoot] {
						graph.VisitedIndexes[resolved.RelPathFromRoot] = true

						fullIdx := filepath.Join(projectRoot, filepath.FromSlash(resolved.RelPathFromRoot))
						realIdx, evalErr := filepath.EvalSymlinks(fullIdx)
						if evalErr == nil {
							if visitedRealPaths[realIdx] {
								continue
							}
							visitedRealPaths[realIdx] = true
						}

						queue = append(queue, resolved.RelPathFromRoot)
					}
				}
			}
		}
	}

	return graph, nil
}

// checkReachability walks the FISS/ directory and verifies that every .md file is reachable
// from the navigation graph starting at FISS/INDEX.md (FISS-R008).
// Any unreachable Markdown file is reported as an error.
func checkReachability(projectRoot string, reachableFiles map[string]bool, report *model.Report) error {
	if projectRoot == "" {
		projectRoot = "."
	}

	fissDir := filepath.Join(projectRoot, "FISS")
	if _, err := os.Stat(fissDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(fissDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// Check if it's a markdown file
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// If it's a symlink, verify if target is a directory
		if d.Type()&os.ModeSymlink != 0 {
			info, statErr := os.Stat(path)
			if statErr != nil {
				// Broken symlink; skip
				return nil
			}
			if info.IsDir() {
				return nil
			}
		}

		relPath, relErr := filepath.Rel(projectRoot, path)
		if relErr != nil {
			relPath = path
		}
		relPath = filepath.ToSlash(filepath.Clean(relPath))

		// Root files (INDEX.md and BOOTSTRAP.md) are mandatory root files governed by FISS-R002 and FISS-R003.
		if relPath == "FISS/INDEX.md" || relPath == "FISS/BOOTSTRAP.md" {
			return nil
		}

		if reachableFiles == nil || !reachableFiles[relPath] {
			report.Add(model.Issue{
				RuleID:   "FISS-R008",
				Severity: model.SeverityError,
				FilePath: relPath,
				Line:     0,
				Message:  "unreachable markdown file (orphan)",
			})
		}

		return nil
	})
}

// extractConditionText extracts the condition string after 'Read when:' from a navigation entry.
func extractConditionText(entry NavEntry) string {
	if !entry.HasNextLine {
		return ""
	}
	trimmed := strings.TrimSpace(entry.NextLineRaw)
	const marker = "Read when:"
	idx := strings.Index(trimmed, marker)
	if idx != -1 {
		return strings.TrimSpace(trimmed[idx+len(marker):])
	}
	return ""
}

// checkCompositeAreas validates rule FISS-R007:
// A composite area (a directory representing an area referenced in navigation) MUST contain its own INDEX.md.
func checkCompositeAreas(projectRoot string, report *model.Report) error {
	if projectRoot == "" {
		projectRoot = "."
	}

	fissDir := filepath.Join(projectRoot, "FISS")
	if _, err := os.Stat(fissDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(fissDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Name() != "INDEX.md" {
			return nil
		}

		relIndex, relErr := filepath.Rel(projectRoot, path)
		if relErr != nil {
			relIndex = path
		}
		relIndex = filepath.ToSlash(filepath.Clean(relIndex))

		f, openErr := os.Open(path)
		if openErr != nil {
			return fmt.Errorf("opening %s: %w", relIndex, openErr)
		}
		entries, parseErr := parseIndexNavEntries(f)
		f.Close()
		if parseErr != nil {
			return fmt.Errorf("parsing %s: %w", relIndex, parseErr)
		}

		for _, entry := range entries {
			if isExternalURL(entry.Target) {
				continue
			}

			resolved, resErr := resolveRelativeTarget(projectRoot, relIndex, entry.Target)
			if resErr != nil || resolved.IsSelfAnchor {
				continue
			}

			exists, isDir, checkErr := checkPathExistsCaseSensitive(projectRoot, resolved.RelPathFromRoot)
			if checkErr != nil {
				continue
			}

			// If target is directly a directory, verify it has an INDEX.md
			if exists && isDir {
				indexInDirRel := filepath.ToSlash(filepath.Join(resolved.RelPathFromRoot, "INDEX.md"))
				idxExists, idxIsDir, _ := checkPathExistsCaseSensitive(projectRoot, indexInDirRel)
				if !idxExists || idxIsDir {
					report.Add(model.Issue{
						RuleID:   "FISS-R007",
						Severity: model.SeverityError,
						FilePath: relIndex,
						Line:     entry.Line,
						Message:  fmt.Sprintf("composite area missing INDEX.md: %s", resolved.RelPathFromRoot),
					})
				}
				continue
			}

			// If target was <dir>/INDEX.md, and <dir> exists on disk as a directory but INDEX.md is missing
			if !exists && strings.HasSuffix(resolved.RelPathFromRoot, "/INDEX.md") {
				dirRel := strings.TrimSuffix(resolved.RelPathFromRoot, "/INDEX.md")
				dirExists, dirIsDir, _ := checkPathExistsCaseSensitive(projectRoot, dirRel)
				if dirExists && dirIsDir {
					report.Add(model.Issue{
						RuleID:   "FISS-R007",
						Severity: model.SeverityError,
						FilePath: relIndex,
						Line:     entry.Line,
						Message:  fmt.Sprintf("composite area missing INDEX.md: %s", dirRel),
					})
				}
			}
		}

		return nil
	})
}

// checkOverridesRule validates rule FISS-R009:
// If FISS/overrides/ exists, it MUST contain INDEX.md, and FISS/INDEX.md MUST link
// to FISS/overrides/INDEX.md with a read condition requiring it before skill use.
func checkOverridesRule(projectRoot string, report *model.Report) error {
	if projectRoot == "" {
		projectRoot = "."
	}

	overridesRel := "FISS/overrides"
	exists, isDir, err := checkPathExistsCaseSensitive(projectRoot, overridesRel)
	if err != nil {
		return fmt.Errorf("checking %s: %w", overridesRel, err)
	}
	if !exists || !isDir {
		// Overrides directory does not exist; rule is satisfied (overrides are optional).
		return nil
	}

	// 1. FISS/overrides/ MUST contain INDEX.md
	indexRel := "FISS/overrides/INDEX.md"
	idxExists, idxIsDir, idxErr := checkPathExistsCaseSensitive(projectRoot, indexRel)
	if idxErr != nil {
		return fmt.Errorf("checking %s: %w", indexRel, idxErr)
	}
	if !idxExists || idxIsDir {
		report.Add(model.Issue{
			RuleID:   "FISS-R009",
			Severity: model.SeverityError,
			FilePath: "FISS/overrides",
			Line:     0,
			Message:  "required file INDEX.md not found in overrides directory",
		})
	}

	// 2. FISS/INDEX.md MUST link to overrides/INDEX.md (or overrides/)
	rootIndexRel := "FISS/INDEX.md"
	rootFullPath := filepath.Join(projectRoot, filepath.FromSlash(rootIndexRel))
	f, openErr := os.Open(rootFullPath)
	if openErr != nil {
		// If root index cannot be opened, root checks already handle it
		return nil
	}
	defer f.Close()

	entries, parseErr := parseIndexNavEntries(f)
	if parseErr != nil {
		return fmt.Errorf("parsing %s: %w", rootIndexRel, parseErr)
	}

	var overrideEntry *NavEntry
	for _, entry := range entries {
		if isExternalURL(entry.Target) {
			continue
		}
		resolved, resErr := resolveRelativeTarget(projectRoot, rootIndexRel, entry.Target)
		if resErr != nil || resolved.IsSelfAnchor {
			continue
		}

		if resolved.RelPathFromRoot == "FISS/overrides/INDEX.md" || resolved.RelPathFromRoot == "FISS/overrides" {
			overrideEntry = &entry
			break
		}
	}

	if overrideEntry == nil {
		report.Add(model.Issue{
			RuleID:   "FISS-R009",
			Severity: model.SeverityError,
			FilePath: "FISS/INDEX.md",
			Line:     0,
			Message:  "missing link to overrides/INDEX.md",
		})
		return nil
	}

	// 3. Read condition MUST require reading before skill use (contain "skill" case-insensitively)
	condText := extractConditionText(*overrideEntry)
	if !strings.Contains(strings.ToLower(condText), "skill") {
		line := overrideEntry.Line
		if overrideEntry.NextLineNum > 0 {
			line = overrideEntry.NextLineNum
		}
		report.Add(model.Issue{
			RuleID:   "FISS-R009",
			Severity: model.SeverityError,
			FilePath: "FISS/INDEX.md",
			Line:     line,
			Message:  "read condition for overrides must require reading before skill usage",
		})
	}

	return nil
}

