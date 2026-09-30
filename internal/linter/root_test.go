package linter

import (
	"os"
	"path/filepath"
	"testing"

	"fiss-lint/internal/model"
)

func TestCheckRootDir(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, root string)
		wantFound     bool
		wantErr       bool
		wantIssues    int
		checkIssue    func(t *testing.T, issues []model.Issue)
	}{
		{
			name: "valid FISS directory exists",
			setup: func(t *testing.T, root string) {
				err := os.Mkdir(filepath.Join(root, "FISS"), 0o755)
				if err != nil {
					t.Fatalf("failed to create FISS dir: %v", err)
				}
			},
			wantFound:  true,
			wantErr:    false,
			wantIssues: 0,
		},
		{
			name: "missing FISS directory",
			setup: func(t *testing.T, root string) {
				// empty root dir
			},
			wantFound:  false,
			wantErr:    false,
			wantIssues: 1,
			checkIssue: func(t *testing.T, issues []model.Issue) {
				issue := issues[0]
				if issue.RuleID != "FISS-R001" {
					t.Errorf("expected RuleID 'FISS-R001', got '%s'", issue.RuleID)
				}
				if issue.Severity != model.SeverityError {
					t.Errorf("expected SeverityError, got %v", issue.Severity)
				}
				if issue.FilePath != "." {
					t.Errorf("expected FilePath '.', got '%s'", issue.FilePath)
				}
				if issue.Line != 0 {
					t.Errorf("expected Line 0, got %d", issue.Line)
				}
				wantMsg := "directory FISS/ not found in project root"
				if issue.Message != wantMsg {
					t.Errorf("expected message '%s', got '%s'", wantMsg, issue.Message)
				}
				wantFormat := "[ERROR] [FISS-R001] .: directory FISS/ not found in project root"
				if issue.Format() != wantFormat {
					t.Errorf("expected Format '%s', got '%s'", wantFormat, issue.Format())
				}
			},
		},
		{
			name: "wrong case directory (fiss)",
			setup: func(t *testing.T, root string) {
				err := os.Mkdir(filepath.Join(root, "fiss"), 0o755)
				if err != nil {
					t.Fatalf("failed to create fiss dir: %v", err)
				}
			},
			wantFound:  false,
			wantErr:    false,
			wantIssues: 1,
			checkIssue: func(t *testing.T, issues []model.Issue) {
				if issues[0].RuleID != "FISS-R001" {
					t.Errorf("expected RuleID 'FISS-R001', got '%s'", issues[0].RuleID)
				}
			},
		},
		{
			name: "FISS is a file not a directory",
			setup: func(t *testing.T, root string) {
				err := os.WriteFile(filepath.Join(root, "FISS"), []byte("not a directory"), 0o644)
				if err != nil {
					t.Fatalf("failed to create FISS file: %v", err)
				}
			},
			wantFound:  false,
			wantErr:    false,
			wantIssues: 1,
			checkIssue: func(t *testing.T, issues []model.Issue) {
				if issues[0].RuleID != "FISS-R001" {
					t.Errorf("expected RuleID 'FISS-R001', got '%s'", issues[0].RuleID)
				}
			},
		},
		{
			name: "FISS is a valid symlink to directory",
			setup: func(t *testing.T, root string) {
				realDir := filepath.Join(root, "actual_fiss")
				if err := os.Mkdir(realDir, 0o755); err != nil {
					t.Fatalf("failed to create actual_fiss dir: %v", err)
				}
				if err := os.Symlink(realDir, filepath.Join(root, "FISS")); err != nil {
					t.Fatalf("failed to create symlink: %v", err)
				}
			},
			wantFound:  true,
			wantErr:    false,
			wantIssues: 0,
		},
		{
			name: "non-existent root directory",
			setup: func(t *testing.T, root string) {
				// root will be non-existent path
			},
			wantFound: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			targetPath := tempDir
			if tt.wantErr {
				targetPath = filepath.Join(tempDir, "does-not-exist")
			} else {
				tt.setup(t, targetPath)
			}

			report := model.NewReport()
			found, err := checkRootDir(targetPath, report)

			if (err != nil) != tt.wantErr {
				t.Fatalf("checkRootDir() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if found != tt.wantFound {
				t.Errorf("checkRootDir() found = %v, wantFound = %v", found, tt.wantFound)
			}
			if len(report.Issues) != tt.wantIssues {
				t.Fatalf("expected %d issues, got %d: %v", tt.wantIssues, len(report.Issues), report.Issues)
			}
			if tt.checkIssue != nil {
				tt.checkIssue(t, report.Issues)
			}
		})
	}
}

func TestLinter_Lint_FISS_R001(t *testing.T) {
	t.Run("missing FISS stops cascade", func(t *testing.T) {
		tempDir := t.TempDir()
		l := New()
		report, err := l.Lint(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !report.HasErrors() {
			t.Errorf("expected report to have errors")
		}
		if report.ErrorsCount() != 1 {
			t.Errorf("expected exactly 1 error, got %d", report.ErrorsCount())
		}
		if report.Issues[0].RuleID != "FISS-R001" {
			t.Errorf("expected RuleID 'FISS-R001', got '%s'", report.Issues[0].RuleID)
		}
	})

	t.Run("valid FISS produces no FISS-R001 error", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.Mkdir(filepath.Join(tempDir, "FISS"), 0o755); err != nil {
			t.Fatalf("failed to create FISS dir: %v", err)
		}
		l := New()
		report, err := l.Lint(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// In Step 2, no other checks exist yet or FISS is valid
		for _, issue := range report.Issues {
			if issue.RuleID == "FISS-R001" {
				t.Errorf("unexpected FISS-R001 issue: %v", issue)
			}
		}
	})

	t.Run("empty string defaults to current directory", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.Mkdir(filepath.Join(tempDir, "FISS"), 0o755); err != nil {
			t.Fatalf("failed to create FISS dir: %v", err)
		}
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get wd: %v", err)
		}
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(wd)
		})

		l := New()
		report, err := l.Lint("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, issue := range report.Issues {
			if issue.RuleID == "FISS-R001" {
				t.Errorf("unexpected FISS-R001 issue: %v", issue)
			}
		}
	})
}
