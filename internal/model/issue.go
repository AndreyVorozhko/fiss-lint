package model

import "fmt"

// Severity represents the severity level of a FISS validation issue.
type Severity int

const (
	// SeverityError indicates a violation of normative MUST/MUST NOT/REQUIRED rules.
	SeverityError Severity = iota
	// SeverityWarning indicates a deviation from SHOULD/RECOMMENDED guidelines.
	SeverityWarning
)

// String returns the uppercase string representation of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "ERROR"
	case SeverityWarning:
		return "WARNING"
	default:
		return "UNKNOWN"
	}
}

// Issue represents a single rule validation diagnostic finding.
type Issue struct {
	RuleID   string
	Severity Severity
	FilePath string
	Line     int
	Message  string
}

// Format returns the standard human-readable diagnostic string representation.
// Format: [SEVERITY] [RULE_ID] path:line: message or [SEVERITY] [RULE_ID] path: message.
func (i Issue) Format() string {
	if i.FilePath == "" {
		return fmt.Sprintf("[%s] [%s] %s", i.Severity, i.RuleID, i.Message)
	}
	if i.Line > 0 {
		return fmt.Sprintf("[%s] [%s] %s:%d: %s", i.Severity, i.RuleID, i.FilePath, i.Line, i.Message)
	}
	return fmt.Sprintf("[%s] [%s] %s: %s", i.Severity, i.RuleID, i.FilePath, i.Message)
}

// Report collects all diagnostic issues produced during FISS validation.
type Report struct {
	Issues []Issue
}

// NewReport initializes an empty Report collector.
func NewReport() *Report {
	return &Report{
		Issues: make([]Issue, 0),
	}
}

// Add appends a new Issue to the report.
func (r *Report) Add(issue Issue) {
	r.Issues = append(r.Issues, issue)
}

// HasErrors returns true if the report contains one or more issues with SeverityError.
func (r *Report) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

// ErrorsCount returns the total number of error-level issues in the report.
func (r *Report) ErrorsCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			count++
		}
	}
	return count
}

// WarningsCount returns the total number of warning-level issues in the report.
func (r *Report) WarningsCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			count++
		}
	}
	return count
}
