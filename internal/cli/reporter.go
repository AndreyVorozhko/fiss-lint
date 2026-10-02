package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"fiss-lint/internal/model"
)

// Reporter formats and writes validation issues from model.Report.
type Reporter interface {
	Report(report *model.Report, w io.Writer) error
}

// JSONReport represents the top-level structured machine-readable report schema.
type JSONReport struct {
	Issues  []JSONIssue `json:"issues"`
	Summary JSONSummary `json:"summary"`
}

// JSONIssue represents a single rule violation diagnostic finding in JSON format.
type JSONIssue struct {
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
}

// JSONSummary aggregates total counts of issues grouped by severity level.
type JSONSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Total    int `json:"total"`
}

// JSONReporter formats validation diagnostics as a structured JSON payload.
type JSONReporter struct{}

// NewJSONReporter creates a new JSONReporter instance.
func NewJSONReporter() Reporter {
	return &JSONReporter{}
}

// Report formats the report as indented JSON and writes it to w.
func (r *JSONReporter) Report(report *model.Report, w io.Writer) error {
	issues := make([]JSONIssue, 0, len(report.Issues))
	for _, issue := range report.Issues {
		issues = append(issues, JSONIssue{
			RuleID:   issue.RuleID,
			Severity: issue.Severity.String(),
			File:     issue.FilePath,
			Line:     issue.Line,
			Message:  issue.Message,
		})
	}

	jr := JSONReport{
		Issues: issues,
		Summary: JSONSummary{
			Errors:   report.ErrorsCount(),
			Warnings: report.WarningsCount(),
			Total:    len(report.Issues),
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jr)
}

// TextReporter formats validation issues as human-readable diagnostic lines.
type TextReporter struct{}

// NewTextReporter creates a new TextReporter instance.
func NewTextReporter() Reporter {
	return &TextReporter{}
}

// Report formats the report as line-by-line diagnostic text.
func (r *TextReporter) Report(report *model.Report, w io.Writer) error {
	for _, issue := range report.Issues {
		if _, err := fmt.Fprintln(w, issue.Format()); err != nil {
			return err
		}
	}
	return nil
}
