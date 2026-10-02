package linter

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// ErrTargetEscapesRoot is returned when a relative link attempts to navigate outside the project root.
	ErrTargetEscapesRoot = errors.New("target path escapes project root")
	// ErrTargetAbsolute is returned when a link uses an absolute filesystem path.
	ErrTargetAbsolute = errors.New("target path must not be absolute")
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
	cleanTarget := stripAnchorAndQuery(target)
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
