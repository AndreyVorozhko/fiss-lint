package linter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fiss-lint/internal/model"
)

var (
	// ErrTargetEscapesRoot is returned when a relative link attempts to navigate outside the project root.
	ErrTargetEscapesRoot = errors.New("target path escapes project root")
	// ErrTargetAbsolute is returned when a link uses an absolute filesystem path.
	ErrTargetAbsolute = errors.New("target path must not be absolute")
	// ErrTargetEmpty is returned when a link target is empty.
	ErrTargetEmpty = errors.New("target path is empty")
)

// isExternalURL checks whether the given link target has an external scheme (http, https, mailto, etc.).
func isExternalURL(target string) bool {
	lower := strings.ToLower(strings.TrimSpace(target))
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:") ||
		strings.HasPrefix(lower, "ftp://") ||
		strings.Contains(lower, "://")
}

// stripAnchorAndQuery removes anchor (#...) and query (?...) fragments from a link target.
func stripAnchorAndQuery(target string) string {
	if idx := strings.IndexAny(target, "#?"); idx != -1 {
		target = target[:idx]
	}
	return strings.TrimSpace(target)
}

// ResolvedTarget holds the resolved filesystem paths for a navigation link target.
type ResolvedTarget struct {
	OriginalTarget  string // Original target string from markdown link
	CleanTarget     string // Target without anchor/query fragments
	RelPathFromRoot string // Normalized path relative to projectRoot (POSIX forward slash)
	FullPath        string // Absolute or OS-specific path on disk
	IsSelfAnchor    bool   // True if target was only an anchor within the same file (#...)
}

// resolveRelativeTarget resolves a link target relative to the directory containing indexRelPath.
// indexRelPath is the path to the INDEX.md file relative to projectRoot (e.g., "FISS/INDEX.md").
func resolveRelativeTarget(projectRoot, indexRelPath, target string) (*ResolvedTarget, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return nil, ErrTargetEmpty
	}

	cleanTarget := stripAnchorAndQuery(trimmed)
	if cleanTarget == "" {
		// Pure anchor link within the current file
		fullPath := filepath.Join(projectRoot, filepath.FromSlash(indexRelPath))
		return &ResolvedTarget{
			OriginalTarget:  target,
			CleanTarget:     cleanTarget,
			RelPathFromRoot: filepath.ToSlash(indexRelPath),
			FullPath:        fullPath,
			IsSelfAnchor:    true,
		}, nil
	}

	targetFsPath := filepath.FromSlash(cleanTarget)
	if filepath.IsAbs(targetFsPath) {
		return nil, fmt.Errorf("%w: %s", ErrTargetAbsolute, target)
	}

	indexDir := filepath.Dir(filepath.FromSlash(indexRelPath))
	resolvedRel := filepath.Clean(filepath.Join(indexDir, targetFsPath))

	if resolvedRel == ".." || strings.HasPrefix(resolvedRel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("%w: %s", ErrTargetEscapesRoot, target)
	}

	fullPath := filepath.Join(projectRoot, resolvedRel)

	return &ResolvedTarget{
		OriginalTarget:  target,
		CleanTarget:     cleanTarget,
		RelPathFromRoot: filepath.ToSlash(resolvedRel),
		FullPath:        fullPath,
		IsSelfAnchor:    false,
	}, nil
}

// checkPathExistsCaseSensitive traverses relPath from projectRoot component by component,
// ensuring strict case-sensitive matching against directory entries via os.ReadDir.
// Returns exists, isDir, and any unexpected filesystem error.
func checkPathExistsCaseSensitive(projectRoot, relPath string) (exists bool, isDir bool, err error) {
	relPath = filepath.Clean(relPath)
	if relPath == "." || relPath == "" {
		info, statErr := os.Stat(projectRoot)
		if statErr != nil {
			return false, false, statErr
		}
		return true, info.IsDir(), nil
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	currentDir := projectRoot

	for i, part := range parts {
		if part == "." || part == "" {
			continue
		}
		if part == ".." {
			currentDir = filepath.Dir(currentDir)
			continue
		}

		entries, readErr := os.ReadDir(currentDir)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				return false, false, nil
			}
			return false, false, readErr
		}

		var found os.DirEntry
		for _, entry := range entries {
			if entry.Name() == part {
				found = entry
				break
			}
		}

		if found == nil {
			return false, false, nil
		}

		currentPath := filepath.Join(currentDir, part)
		isLast := (i == len(parts)-1)

		if isLast {
			if found.Type()&os.ModeSymlink != 0 {
				info, statErr := os.Stat(currentPath)
				if statErr != nil {
					return false, false, nil
				}
				return true, info.IsDir(), nil
			}
			return true, found.IsDir(), nil
		}

		// Intermediate component must be a directory
		if found.Type()&os.ModeSymlink != 0 {
			info, statErr := os.Stat(currentPath)
			if statErr != nil || !info.IsDir() {
				return false, false, nil
			}
		} else if !found.IsDir() {
			return false, false, nil
		}

		currentDir = currentPath
	}

	return true, true, nil
}

// validateIndexLinksIntegrity checks rule FISS-R006 for all navigation entries in an INDEX.md file.
func validateIndexLinksIntegrity(projectRoot, indexRelPath string, entries []NavEntry, report *model.Report) {
	for _, entry := range entries {
		if isExternalURL(entry.Target) {
			continue
		}

		resolved, err := resolveRelativeTarget(projectRoot, indexRelPath, entry.Target)
		if err != nil {
			if errors.Is(err, ErrTargetEscapesRoot) {
				report.Add(model.Issue{
					RuleID:   "FISS-R006",
					Severity: model.SeverityError,
					FilePath: indexRelPath,
					Line:     entry.Line,
					Message:  fmt.Sprintf("link target escapes project root: %s", entry.Target),
				})
				continue
			}
			if errors.Is(err, ErrTargetAbsolute) {
				report.Add(model.Issue{
					RuleID:   "FISS-R006",
					Severity: model.SeverityError,
					FilePath: indexRelPath,
					Line:     entry.Line,
					Message:  fmt.Sprintf("link target must not be absolute: %s", entry.Target),
				})
				continue
			}
			if errors.Is(err, ErrTargetEmpty) {
				report.Add(model.Issue{
					RuleID:   "FISS-R006",
					Severity: model.SeverityError,
					FilePath: indexRelPath,
					Line:     entry.Line,
					Message:  "link target is empty",
				})
				continue
			}
			report.Add(model.Issue{
				RuleID:   "FISS-R006",
				Severity: model.SeverityError,
				FilePath: indexRelPath,
				Line:     entry.Line,
				Message:  fmt.Sprintf("invalid link target %q: %v", entry.Target, err),
			})
			continue
		}

		if resolved.IsSelfAnchor {
			continue
		}

		exists, isDir, checkErr := checkPathExistsCaseSensitive(projectRoot, resolved.RelPathFromRoot)
		if checkErr != nil {
			report.Add(model.Issue{
				RuleID:   "FISS-R006",
				Severity: model.SeverityError,
				FilePath: indexRelPath,
				Line:     entry.Line,
				Message:  fmt.Sprintf("checking target %q: %v", entry.Target, checkErr),
			})
			continue
		}

		if !exists {
			report.Add(model.Issue{
				RuleID:   "FISS-R006",
				Severity: model.SeverityError,
				FilePath: indexRelPath,
				Line:     entry.Line,
				Message:  fmt.Sprintf("target file does not exist: %s", entry.Target),
			})
			continue
		}

		if !isDir {
			if !strings.HasSuffix(strings.ToLower(resolved.CleanTarget), ".md") {
				report.Add(model.Issue{
					RuleID:   "FISS-R006",
					Severity: model.SeverityError,
					FilePath: indexRelPath,
					Line:     entry.Line,
					Message:  fmt.Sprintf("link target must be a Markdown file (.md): %s", entry.Target),
				})
			}
			continue
		}

		// isDir == true: verify that the target directory contains an INDEX.md file (composite area).
		indexInDirRel := filepath.ToSlash(filepath.Join(resolved.RelPathFromRoot, "INDEX.md"))
		indexExists, indexIsDir, indexErr := checkPathExistsCaseSensitive(projectRoot, indexInDirRel)
		if indexErr != nil {
			report.Add(model.Issue{
				RuleID:   "FISS-R006",
				Severity: model.SeverityError,
				FilePath: indexRelPath,
				Line:     entry.Line,
				Message:  fmt.Sprintf("checking target directory %q: %v", entry.Target, indexErr),
			})
			continue
		}

		if !indexExists || indexIsDir {
			report.Add(model.Issue{
				RuleID:   "FISS-R006",
				Severity: model.SeverityError,
				FilePath: indexRelPath,
				Line:     entry.Line,
				Message:  fmt.Sprintf("link targets a bare directory without INDEX.md: %s", entry.Target),
			})
			continue
		}
	}
}
