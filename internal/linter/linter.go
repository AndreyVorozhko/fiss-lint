package linter

import (
	"fiss-lint/internal/model"
)

// Linter represents the FISS rule evaluation engine.
type Linter struct{}

// New creates a new Linter instance.
func New() *Linter {
	return &Linter{}
}

// Lint runs all deterministic FISS validation rules against the specified project root.
func (l *Linter) Lint(projectRoot string) (*model.Report, error) {
	if projectRoot == "" {
		projectRoot = "."
	}

	report := model.NewReport()

	// Rule FISS-R001: Root FISS/ directory presence.
	fissFound, err := checkRootDir(projectRoot, report)
	if err != nil {
		return nil, err
	}
	if !fissFound {
		// Cascade stop: without FISS/ directory, root structure checks cannot proceed.
		return report, nil
	}

	// Rule FISS-R002: Mandatory files INDEX.md and BOOTSTRAP.md in FISS/.
	hasIndex, _, err := checkMandatoryFiles(projectRoot, report)
	if err != nil {
		return nil, err
	}

	// Rule FISS-R003 & FISS-R011: Check links in FISS/INDEX.md (only if INDEX.md exists).
	if hasIndex {
		if err := checkIndexLinks(projectRoot, report); err != nil {
			return nil, err
		}

		// Rule FISS-R004 & FISS-R005: Validate navigation format (Read when:) across all INDEX.md files.
		// Rule FISS-R006: Link integrity.
		if err := validateAllIndexes(projectRoot, report); err != nil {
			return nil, err
		}

		// Rule FISS-R007: Composite area validation.
		if err := checkCompositeAreas(projectRoot, report); err != nil {
			return nil, err
		}

		// Rule FISS-R008: Navigation graph reachability and orphan file detection.
		graph, err := buildNavigationGraph(projectRoot)
		if err != nil {
			return nil, err
		}
		if err := checkReachability(projectRoot, graph.ReachableFiles, report); err != nil {
			return nil, err
		}

		// Rule FISS-R009: Overrides directory and entry rules.
		if err := checkOverridesRule(projectRoot, report); err != nil {
			return nil, err
		}
	}

	return report, nil
}
