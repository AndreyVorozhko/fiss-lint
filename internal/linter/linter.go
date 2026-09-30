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
	_, _, err = checkMandatoryFiles(projectRoot, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}
