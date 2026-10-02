package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fiss-lint/internal/model"
)

func TestParseIndexNavEntries(t *testing.T) {
	t.Run("standard two-line navigation entries", func(t *testing.T) {
		input := `# Project Intellectual Space

Entry point to the intellectual space.

- [Baseline Context](BOOTSTRAP.md)
  Read when: read always before beginning work on the project.
- [Project Overrides](overrides/INDEX.md)
  Read when: before using any skill or executing skill-governed workflows.
`
		entries, err := parseIndexNavEntries(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}

		// Entry 1
		e1 := entries[0]
		if e1.Line != 5 {
			t.Errorf("e1.Line = %d, want 5", e1.Line)
		}
		if e1.Title != "Baseline Context" {
			t.Errorf("e1.Title = %q, want 'Baseline Context'", e1.Title)
		}
		if e1.Target != "BOOTSTRAP.md" {
			t.Errorf("e1.Target = %q, want 'BOOTSTRAP.md'", e1.Target)
		}
		if !e1.HasNextLine {
			t.Errorf("e1.HasNextLine = false, want true")
		}
		if e1.NextLineNum != 6 {
			t.Errorf("e1.NextLineNum = %d, want 6", e1.NextLineNum)
		}
		if e1.NextLineRaw != "  Read when: read always before beginning work on the project." {
			t.Errorf("e1.NextLineRaw = %q", e1.NextLineRaw)
		}

		// Entry 2
		e2 := entries[1]
		if e2.Line != 7 {
			t.Errorf("e2.Line = %d, want 7", e2.Line)
		}
		if e2.Title != "Project Overrides" {
			t.Errorf("e2.Title = %q, want 'Project Overrides'", e2.Title)
		}
		if e2.Target != "overrides/INDEX.md" {
			t.Errorf("e2.Target = %q, want 'overrides/INDEX.md'", e2.Target)
		}
		if !e2.HasNextLine {
			t.Errorf("e2.HasNextLine = false, want true")
		}
		if e2.NextLineNum != 8 {
			t.Errorf("e2.NextLineNum = %d, want 8", e2.NextLineNum)
		}
	})

	t.Run("consecutive single line navigation entries", func(t *testing.T) {
		input := `- [First](first.md)
- [Second](second.md)
  Read when: condition for second
`
		entries, err := parseIndexNavEntries(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}

		if entries[0].Line != 1 {
			t.Errorf("entries[0].Line = %d, want 1", entries[0].Line)
		}
		if entries[0].NextLineRaw != "- [Second](second.md)" {
			t.Errorf("entries[0].NextLineRaw = %q, want '- [Second](second.md)'", entries[0].NextLineRaw)
		}

		if entries[1].Line != 2 {
			t.Errorf("entries[1].Line = %d, want 2", entries[1].Line)
		}
		if entries[1].NextLineRaw != "  Read when: condition for second" {
			t.Errorf("entries[1].NextLineRaw = %q", entries[1].NextLineRaw)
		}
	})

	t.Run("navigation entry at EOF", func(t *testing.T) {
		input := `- [Trailing](trailing.md)`
		entries, err := parseIndexNavEntries(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].HasNextLine {
			t.Errorf("expected HasNextLine = false at EOF")
		}
	})

	t.Run("windows line endings", func(t *testing.T) {
		input := "- [Win](win.md)\r\n  Read when: on windows\r\n"
		entries, err := parseIndexNavEntries(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].NextLineRaw != "  Read when: on windows" {
			t.Errorf("expected clean NextLineRaw, got %q", entries[0].NextLineRaw)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		entries, err := parseIndexNavEntries(strings.NewReader(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})
}

func TestValidateNavEntries(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		isRootIndex bool
		input       string
		wantIssues  int
		checkIssues func(t *testing.T, issues []model.Issue)
	}{
		{
			name:        "valid navigation entries with bootstrap",
			filePath:    "FISS/INDEX.md",
			isRootIndex: true,
			input: `- [Baseline Context](BOOTSTRAP.md)
  Read when: read always before beginning work.
- [Project Overrides](overrides/INDEX.md)
  Read when: before using any skill.
`,
			wantIssues: 0,
		},
		{
			name:        "missing condition line (consecutive links)",
			filePath:    "FISS/INDEX.md",
			isRootIndex: true,
			input: `- [Item 1](item1.md)
- [Item 2](item2.md)
  Read when: condition for item 2
`,
			wantIssues: 1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				iss := issues[0]
				if iss.RuleID != "FISS-R005" {
					t.Errorf("expected RuleID 'FISS-R005', got %q", iss.RuleID)
				}
				if iss.Line != 1 {
					t.Errorf("expected Line 1, got %d", iss.Line)
				}
				if iss.Message != "missing 'Read when:' condition for navigation entry" {
					t.Errorf("expected missing message, got %q", iss.Message)
				}
			},
		},
		{
			name:        "missing condition line at EOF",
			filePath:    "FISS/INDEX.md",
			isRootIndex: true,
			input:       `- [Item at EOF](item.md)`,
			wantIssues:  1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				iss := issues[0]
				if iss.RuleID != "FISS-R005" || iss.Line != 1 {
					t.Errorf("unexpected issue: %+v", iss)
				}
			},
		},
		{
			name:        "invalid indentation (4 spaces instead of 2)",
			filePath:    "FISS/INDEX.md",
			isRootIndex: true,
			input: `- [Item](item.md)
    Read when: four spaces indent
`,
			wantIssues: 1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				iss := issues[0]
				if iss.RuleID != "FISS-R005" {
					t.Errorf("expected RuleID 'FISS-R005', got %q", iss.RuleID)
				}
				if iss.Line != 2 {
					t.Errorf("expected Line 2, got %d", iss.Line)
				}
				if iss.Message != "invalid indentation for 'Read when:' condition (expected exactly 2 spaces)" {
					t.Errorf("expected invalid indentation message, got %q", iss.Message)
				}
			},
		},
		{
			name:        "invalid indentation (0 spaces)",
			filePath:    "FISS/INDEX.md",
			isRootIndex: true,
			input: `- [Item](item.md)
Read when: zero spaces indent
`,
			wantIssues: 1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				iss := issues[0]
				if iss.RuleID != "FISS-R005" || iss.Line != 2 {
					t.Errorf("unexpected issue: %+v", iss)
				}
			},
		},
		{
			name:        "localized marker (Читать, когда:) is prohibited by FISS v1.0.0",
			filePath:    "FISS/INDEX.md",
			isRootIndex: true,
			input: `- [Документ](doc.md)
  Читать, когда: всегда перед началом работы
`,
			wantIssues: 1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				iss := issues[0]
				if iss.RuleID != "FISS-R005" {
					t.Errorf("expected RuleID 'FISS-R005', got %q", iss.RuleID)
				}
				if iss.Line != 2 {
					t.Errorf("expected Line 2, got %d", iss.Line)
				}
				if iss.Message != "invalid read condition marker (must be exact 'Read when:')" {
					t.Errorf("expected invalid marker message, got %q", iss.Message)
				}
			},
		},
		{
			name:        "empty condition text",
			filePath:    "FISS/INDEX.md",
			isRootIndex: true,
			input: `- [Item](item.md)
  Read when:    
`,
			wantIssues: 1,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				iss := issues[0]
				if iss.RuleID != "FISS-R005" {
					t.Errorf("expected RuleID 'FISS-R005', got %q", iss.RuleID)
				}
				if iss.Line != 2 {
					t.Errorf("expected Line 2, got %d", iss.Line)
				}
				if iss.Message != "empty condition text in 'Read when:'" {
					t.Errorf("expected empty condition message, got %q", iss.Message)
				}
			},
		},
		{
			name:        "BOOTSTRAP.md without condition triggers both FISS-R005 and FISS-R004 in root index",
			filePath:    "FISS/INDEX.md",
			isRootIndex: true,
			input: `- [Baseline Context](BOOTSTRAP.md)
- [Next](next.md)
  Read when: condition for next
`,
			wantIssues: 2,
			checkIssues: func(t *testing.T, issues []model.Issue) {
				var hasR005, hasR004 bool
				for _, iss := range issues {
					if iss.RuleID == "FISS-R005" && iss.Line == 1 {
						hasR005 = true
					}
					if iss.RuleID == "FISS-R004" && iss.Line == 1 {
						hasR004 = true
						if iss.Message != "link to BOOTSTRAP.md must have an attached read condition" {
							t.Errorf("unexpected FISS-R004 message: %q", iss.Message)
						}
					}
				}
				if !hasR005 || !hasR004 {
					t.Errorf("expected both FISS-R005 and FISS-R004, got issues: %+v", issues)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := parseIndexNavEntries(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}
			report := model.NewReport()
			validateNavEntries(tt.filePath, entries, tt.isRootIndex, report)

			if len(report.Issues) != tt.wantIssues {
				t.Fatalf("expected %d issues, got %d: %+v", tt.wantIssues, len(report.Issues), report.Issues)
			}
			if tt.checkIssues != nil {
				tt.checkIssues(t, report.Issues)
			}
		})
	}
}

func TestValidateAllIndexes(t *testing.T) {
	tempDir := t.TempDir()
	fissDir := filepath.Join(tempDir, "FISS")
	overridesDir := filepath.Join(fissDir, "overrides")
	if err := os.MkdirAll(overridesDir, 0o755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	rootIndex := `- [Baseline Context](BOOTSTRAP.md)
  Read when: always
- [Overrides](overrides/INDEX.md)
  Read when: before skills
`
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(rootIndex), 0o644); err != nil {
		t.Fatalf("failed to write root INDEX.md: %v", err)
	}

	subIndex := `- [Rule](rule.md)
- [Invalid Single Line](invalid.md)
`
	if err := os.WriteFile(filepath.Join(overridesDir, "INDEX.md"), []byte(subIndex), 0o644); err != nil {
		t.Fatalf("failed to write sub INDEX.md: %v", err)
	}

	report := model.NewReport()
	if err := validateAllIndexes(tempDir, report); err != nil {
		t.Fatalf("unexpected validateAllIndexes error: %v", err)
	}

	// subIndex has 2 invalid entries (missing conditions)
	if report.ErrorsCount() != 2 {
		t.Fatalf("expected 2 errors in subIndex, got %d: %+v", report.ErrorsCount(), report.Issues)
	}
	for _, iss := range report.Issues {
		if iss.FilePath != "FISS/overrides/INDEX.md" {
			t.Errorf("expected FilePath 'FISS/overrides/INDEX.md', got %q", iss.FilePath)
		}
	}
}


