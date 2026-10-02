package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
