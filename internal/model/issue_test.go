package model

import (
	"testing"
)

func TestSeverityString(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{
			name:     "severity error",
			severity: SeverityError,
			want:     "ERROR",
		},
		{
			name:     "severity warning",
			severity: SeverityWarning,
			want:     "WARNING",
		},
		{
			name:     "severity unknown",
			severity: Severity(999),
			want:     "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIssueFormat(t *testing.T) {
	tests := []struct {
		name  string
		issue Issue
		want  string
	}{
		{
			name: "issue with file and line",
			issue: Issue{
				RuleID:   "FISS-R003",
				Severity: SeverityError,
				FilePath: "FISS/INDEX.md",
				Line:     5,
				Message:  "missing link to BOOTSTRAP.md",
			},
			want: "[ERROR] [FISS-R003] FISS/INDEX.md:5: missing link to BOOTSTRAP.md",
		},
		{
			name: "issue with file and line 0",
			issue: Issue{
				RuleID:   "FISS-R001",
				Severity: SeverityError,
				FilePath: ".",
				Line:     0,
				Message:  "directory FISS/ not found in project root",
			},
			want: "[ERROR] [FISS-R001] .: directory FISS/ not found in project root",
		},
		{
			name: "issue without file path",
			issue: Issue{
				RuleID:   "FISS-R001",
				Severity: SeverityError,
				FilePath: "",
				Line:     0,
				Message:  "project root is inaccessible",
			},
			want: "[ERROR] [FISS-R001] project root is inaccessible",
		},
		{
			name: "warning severity issue",
			issue: Issue{
				RuleID:   "FISS-R011",
				Severity: SeverityWarning,
				FilePath: "FISS/INDEX.md",
				Line:     0,
				Message:  "recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found",
			},
			want: "[WARNING] [FISS-R011] FISS/INDEX.md: recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.issue.Format(); got != tt.want {
				t.Errorf("Issue.Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReportCollector(t *testing.T) {
	report := NewReport()

	if report.HasErrors() {
		t.Errorf("expected empty report to have no errors")
	}
	if got := report.ErrorsCount(); got != 0 {
		t.Errorf("expected 0 errors, got %d", got)
	}
	if got := report.WarningsCount(); got != 0 {
		t.Errorf("expected 0 warnings, got %d", got)
	}

	report.Add(Issue{
		RuleID:   "FISS-R011",
		Severity: SeverityWarning,
		FilePath: "FISS/INDEX.md",
		Line:     0,
		Message:  "recommended link warning",
	})

	if report.HasErrors() {
		t.Errorf("report with only warning should not report HasErrors() = true")
	}
	if got := report.ErrorsCount(); got != 0 {
		t.Errorf("expected 0 errors, got %d", got)
	}
	if got := report.WarningsCount(); got != 1 {
		t.Errorf("expected 1 warning, got %d", got)
	}

	report.Add(Issue{
		RuleID:   "FISS-R001",
		Severity: SeverityError,
		FilePath: ".",
		Line:     0,
		Message:  "root directory error",
	})

	if !report.HasErrors() {
		t.Errorf("report with error should report HasErrors() = true")
	}
	if got := report.ErrorsCount(); got != 1 {
		t.Errorf("expected 1 error, got %d", got)
	}
	if got := report.WarningsCount(); got != 1 {
		t.Errorf("expected 1 warning, got %d", got)
	}
}
