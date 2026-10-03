package linter

import (
	"os"
	"path/filepath"
	"testing"

	"fiss-lint/internal/model"
)

func TestCheckIndexLinks(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantIssues  int
		checkIssues func(t *testing.T, issues []model.Issue)
	}{
		{
			name: "valid index with both links",
			content: `# Project Intellectual Space
- [Baseline Context](BOOTSTRAP.md)
  Read when: read always before beginning work on the project.
- [Standard Specification (v1.0.0)](https://fiss.vorozhko.ru/v1.0.0/llms.txt)
  Read when: creating or modifying the intellectual space structure.
`,
			wantIssues: 0,
		},
		{
			name: "missing bootstrap link",
			content: `# Project Intellectual Space
- [Standard Specification](https://fiss.vorozhko.ru)
`,
			wantIssues: 1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				issue := issues[0]
				if issue.RuleID != "FISS-R003" {
					t.Errorf("expected RuleID 'FISS-R003', got '%s'", issue.RuleID)
				}
				if issue.Severity != model.SeverityError {
					t.Errorf("expected SeverityError, got %v", issue.Severity)
				}
				if issue.FilePath != "FISS/INDEX.md" {
					t.Errorf("expected FilePath 'FISS/INDEX.md', got '%s'", issue.FilePath)
				}
				wantMsg := "missing link to BOOTSTRAP.md"
				if issue.Message != wantMsg {
					t.Errorf("expected Message '%s', got '%s'", wantMsg, issue.Message)
				}
				wantFormat := "[ERROR] [FISS-R003] FISS/INDEX.md: missing link to BOOTSTRAP.md"
				if issue.Format() != wantFormat {
					t.Errorf("expected Format '%s', got '%s'", wantFormat, issue.Format())
				}
			},
		},
		{
			name: "missing standard link (warning)",
			content: `# Project Intellectual Space
- [Baseline Context](BOOTSTRAP.md)
`,
			wantIssues: 1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				issue := issues[0]
				if issue.RuleID != "FISS-R011" {
					t.Errorf("expected RuleID 'FISS-R011', got '%s'", issue.RuleID)
				}
				if issue.Severity != model.SeverityWarning {
					t.Errorf("expected SeverityWarning, got %v", issue.Severity)
				}
				if issue.FilePath != "FISS/INDEX.md" {
					t.Errorf("expected FilePath 'FISS/INDEX.md', got '%s'", issue.FilePath)
				}
				wantMsg := "recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found"
				if issue.Message != wantMsg {
					t.Errorf("expected Message '%s', got '%s'", wantMsg, issue.Message)
				}
				wantFormat := "[WARNING] [FISS-R011] FISS/INDEX.md: recommended link to official standard (https://fiss.vorozhko.ru or GitHub mirror) not found"
				if issue.Format() != wantFormat {
					t.Errorf("expected Format '%s', got '%s'", wantFormat, issue.Format())
				}
			},
		},
		{
			name: "valid index with github mirror repo link",
			content: `# Project Intellectual Space
- [Baseline Context](BOOTSTRAP.md)
  Read when: read always before beginning work on the project.
- [Standard Specification](https://github.com/AndreyVorozhko/fiss)
  Read when: creating or modifying the intellectual space structure.
`,
			wantIssues: 0,
		},
		{
			name: "valid index with github mirror branch file link",
			content: `# Project Intellectual Space
- [Baseline Context](BOOTSTRAP.md)
  Read when: read always before beginning work on the project.
- [Standard Specification (v1.0.0)](https://github.com/AndreyVorozhko/fiss/blob/v1.0.0/llms.txt)
  Read when: creating or modifying the intellectual space structure.
`,
			wantIssues: 0,
		},
		{
			name: "valid index with both links (website and github mirror)",
			content: `# Project Intellectual Space
- [Baseline Context](BOOTSTRAP.md)
  Read when: read always before beginning work on the project.
- [Standard Specification (v1.0.0)](https://fiss.vorozhko.ru/v1.0.0/llms.txt)
  Read when: creating or modifying the intellectual space structure.
- [Standard Specification (GitHub Mirror)](https://github.com/AndreyVorozhko/fiss/blob/v1.0.0/llms.txt)
  Read when: недоступен https://fiss.vorozhko.ru или работа ведётся в автономном окружении.
`,
			wantIssues: 0,
		},
		{
			name:       "missing both links",
			content:    `# Empty Index`,
			wantIssues: 2,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				var hasR003, hasR011 bool
				for _, iss := range issues {
					if iss.RuleID == "FISS-R003" {
						hasR003 = true
					}
					if iss.RuleID == "FISS-R011" {
						hasR011 = true
					}
				}
				if !hasR003 || !hasR011 {
					t.Errorf("expected both FISS-R003 and FISS-R011 issues, got %v", issues)
				}
			},
		},
		{
			name: "relative path ./BOOTSTRAP.md with anchor",
			content: `# Index
- [Bootstrap](./BOOTSTRAP.md#quickstart)
- [FISS Standard](https://fiss.vorozhko.ru/spec)
`,
			wantIssues: 0,
		},
		{
			name: "wrong case bootstrap.md is not recognized",
			content: `# Index
- [Bootstrap](bootstrap.md)
- [FISS Standard](https://fiss.vorozhko.ru)
`,
			wantIssues: 1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				if issues[0].RuleID != "FISS-R003" {
					t.Errorf("expected RuleID 'FISS-R003', got '%s'", issues[0].RuleID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			fissDir := filepath.Join(tempDir, "FISS")
			if err := os.Mkdir(fissDir, 0o755); err != nil {
				t.Fatalf("failed to create FISS dir: %v", err)
			}
			indexPath := filepath.Join(fissDir, "INDEX.md")
			if err := os.WriteFile(indexPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write INDEX.md: %v", err)
			}

			report := model.NewReport()
			err := checkIndexLinks(tempDir, report)
			if err != nil {
				t.Fatalf("unexpected checkIndexLinks error: %v", err)
			}
			if len(report.Issues) != tt.wantIssues {
				t.Fatalf("expected %d issues, got %d: %v", tt.wantIssues, len(report.Issues), report.Issues)
			}
			if tt.checkIssues != nil {
				tt.checkIssues(t, report.Issues)
			}
		})
	}
}

func TestCheckIndexLinks_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	report := model.NewReport()
	err := checkIndexLinks(tempDir, report)
	if err == nil {
		t.Fatalf("expected error when INDEX.md not found, got nil")
	}
}

func TestLinter_Lint_Links(t *testing.T) {
	tempDir := t.TempDir()
	fissDir := filepath.Join(tempDir, "FISS")
	if err := os.Mkdir(fissDir, 0o755); err != nil {
		t.Fatalf("failed to create FISS dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte("# Only Title\n"), 0o644); err != nil {
		t.Fatalf("failed to write INDEX.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "BOOTSTRAP.md"), []byte("# Bootstrap\n"), 0o644); err != nil {
		t.Fatalf("failed to write BOOTSTRAP.md: %v", err)
	}

	l := New()
	report, err := l.Lint(tempDir)
	if err != nil {
		t.Fatalf("unexpected lint error: %v", err)
	}
	if !report.HasErrors() {
		t.Errorf("expected report to have errors")
	}
	if report.ErrorsCount() != 1 { // FISS-R003
		t.Errorf("expected 1 error (FISS-R003), got %d", report.ErrorsCount())
	}
	if report.WarningsCount() != 1 { // FISS-R011
		t.Errorf("expected 1 warning (FISS-R011), got %d", report.WarningsCount())
	}
}

