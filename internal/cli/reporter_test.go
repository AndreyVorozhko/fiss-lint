package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"fiss-lint/internal/model"
)

type failWriter struct{}

func (f *failWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestJSONReporter_EmptyReport(t *testing.T) {
	rep := model.NewReport()
	reporter := NewJSONReporter()

	var buf bytes.Buffer
	err := reporter.Report(rep, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed JSONReport
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v, raw output:\n%s", err, buf.String())
	}

	if parsed.Issues == nil {
		t.Fatal("expected issues to be non-nil empty slice")
	}
	if len(parsed.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(parsed.Issues))
	}
	if parsed.Summary.Errors != 0 || parsed.Summary.Warnings != 0 || parsed.Summary.Total != 0 {
		t.Errorf("unexpected summary values: %+v", parsed.Summary)
	}
}

func TestJSONReporter_WithIssues(t *testing.T) {
	rep := model.NewReport()
	rep.Add(model.Issue{
		RuleID:   "FISS-R001",
		Severity: model.SeverityError,
		FilePath: "FISS",
		Line:     0,
		Message:  "directory FISS/ not found",
	})
	rep.Add(model.Issue{
		RuleID:   "FISS-R011",
		Severity: model.SeverityWarning,
		FilePath: "FISS/INDEX.md",
		Line:     5,
		Message:  "missing reference to standard website",
	})

	reporter := NewJSONReporter()

	var buf bytes.Buffer
	err := reporter.Report(rep, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed JSONReport
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v, raw output:\n%s", err, buf.String())
	}

	if len(parsed.Issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(parsed.Issues))
	}

	// Check issue 0
	iss0 := parsed.Issues[0]
	if iss0.RuleID != "FISS-R001" || iss0.Severity != "ERROR" || iss0.File != "FISS" || iss0.Line != 0 || iss0.Message != "directory FISS/ not found" {
		t.Errorf("unexpected issue 0: %+v", iss0)
	}

	// Check issue 1
	iss1 := parsed.Issues[1]
	if iss1.RuleID != "FISS-R011" || iss1.Severity != "WARNING" || iss1.File != "FISS/INDEX.md" || iss1.Line != 5 || iss1.Message != "missing reference to standard website" {
		t.Errorf("unexpected issue 1: %+v", iss1)
	}

	// Check summary
	if parsed.Summary.Errors != 1 {
		t.Errorf("expected 1 error, got %d", parsed.Summary.Errors)
	}
	if parsed.Summary.Warnings != 1 {
		t.Errorf("expected 1 warning, got %d", parsed.Summary.Warnings)
	}
	if parsed.Summary.Total != 2 {
		t.Errorf("expected 2 total, got %d", parsed.Summary.Total)
	}
}

func TestJSONReporter_WriterError(t *testing.T) {
	rep := model.NewReport()
	reporter := NewJSONReporter()

	err := reporter.Report(rep, &failWriter{})
	if err == nil {
		t.Fatal("expected error from failWriter, got nil")
	}
}

func TestTextReporter_EmptyReport(t *testing.T) {
	rep := model.NewReport()
	reporter := NewTextReporter()

	var buf bytes.Buffer
	err := reporter.Report(rep, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestTextReporter_WithIssues(t *testing.T) {
	rep := model.NewReport()
	rep.Add(model.Issue{
		RuleID:   "FISS-R001",
		Severity: model.SeverityError,
		FilePath: "FISS",
		Line:     0,
		Message:  "directory FISS/ not found",
	})

	reporter := NewTextReporter()

	var buf bytes.Buffer
	err := reporter.Report(rep, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "[ERROR] [FISS-R001] FISS: directory FISS/ not found\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestTextReporter_WriterError(t *testing.T) {
	rep := model.NewReport()
	rep.Add(model.Issue{
		RuleID:   "FISS-R001",
		Severity: model.SeverityError,
		FilePath: "FISS",
		Line:     0,
		Message:  "directory FISS/ not found",
	})
	reporter := NewTextReporter()

	err := reporter.Report(rep, &failWriter{})
	if err == nil {
		t.Fatal("expected error from failWriter, got nil")
	}
}

