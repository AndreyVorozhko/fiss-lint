package linter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"fiss-lint/internal/model"
)

func TestFindAgentInstructionFiles(t *testing.T) {
	t.Run("empty directory with no agent files", func(t *testing.T) {
		tempDir := t.TempDir()
		found, err := findAgentInstructionFiles(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(found) != 0 {
			t.Errorf("expected 0 files, got %v", found)
		}
	})

	t.Run("all known files present", func(t *testing.T) {
		tempDir := t.TempDir()
		for _, name := range knownAgentInstructionFiles {
			if err := os.WriteFile(filepath.Join(tempDir, name), []byte("content"), 0644); err != nil {
				t.Fatalf("failed to create %s: %v", name, err)
			}
		}

		found, err := findAgentInstructionFiles(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"AGENTS.md", "CLAUDE.md", ".cursorrules"}
		if !reflect.DeepEqual(found, expected) {
			t.Errorf("expected %v, got %v", expected, found)
		}
	})

	t.Run("partial subset present", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "AGENTS.md"), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create AGENTS.md: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, ".cursorrules"), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create .cursorrules: %v", err)
		}

		found, err := findAgentInstructionFiles(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"AGENTS.md", ".cursorrules"}
		if !reflect.DeepEqual(found, expected) {
			t.Errorf("expected %v, got %v", expected, found)
		}
	})

	t.Run("case sensitivity: lowercase agents.md is ignored", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "agents.md"), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create agents.md: %v", err)
		}

		found, err := findAgentInstructionFiles(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(found) != 0 {
			t.Errorf("expected lowercase agents.md to be ignored, got %v", found)
		}
	})

	t.Run("candidate directory is ignored", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.Mkdir(filepath.Join(tempDir, "AGENTS.md"), 0755); err != nil {
			t.Fatalf("failed to create AGENTS.md dir: %v", err)
		}

		found, err := findAgentInstructionFiles(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(found) != 0 {
			t.Errorf("expected directory AGENTS.md to be ignored, got %v", found)
		}
	})

	t.Run("symlink to regular file is accepted", func(t *testing.T) {
		tempDir := t.TempDir()
		target := filepath.Join(tempDir, "target_agent.txt")
		if err := os.WriteFile(target, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create target: %v", err)
		}
		if err := os.Symlink(target, filepath.Join(tempDir, "CLAUDE.md")); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}

		found, err := findAgentInstructionFiles(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"CLAUDE.md"}
		if !reflect.DeepEqual(found, expected) {
			t.Errorf("expected %v, got %v", expected, found)
		}
	})

	t.Run("non-existent project root returns error", func(t *testing.T) {
		_, err := findAgentInstructionFiles(filepath.Join(t.TempDir(), "nonexistent"))
		if err == nil {
			t.Errorf("expected error for non-existent root, got nil")
		}
	})
}

func TestDirectsToFissIndex(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "markdown link to FISS/INDEX.md",
			content:  "- [FISS Index](FISS/INDEX.md)\n  Read before project work.",
			expected: true,
		},
		{
			name:     "plain text mention of FISS/INDEX.md",
			content:  "Agents must navigate to FISS/INDEX.md",
			expected: true,
		},
		{
			name:     "plain text mention of FISS/",
			content:  "Project documentation is located in FISS/ directory",
			expected: true,
		},
		{
			name:     "windows backslash FISS\\INDEX.md",
			content:  "Read FISS\\INDEX.md before continuing",
			expected: true,
		},
		{
			name:     "windows backslash FISS\\",
			content:  "Check FISS\\ folder",
			expected: true,
		},
		{
			name:     "only standard URL without FISS/ entry point",
			content:  "See https://fiss.vorozhko.ru/v1.0.0/llms.txt for standard specification",
			expected: false,
		},
		{
			name:     "only tool name fiss-lint",
			content:  "This project uses fiss-lint for checking rules.",
			expected: false,
		},
		{
			name:     "lowercase fiss/index.md",
			content:  "Check fiss/index.md for info",
			expected: false,
		},
		{
			name:     "generic instructions without FISS reference",
			content:  "# Agent Instructions\nAlways write tests and maintain git commit style.",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := directsToFissIndex(tc.content)
			if got != tc.expected {
				t.Errorf("directsToFissIndex() = %v, expected %v", got, tc.expected)
			}
		})
	}
}

func TestCheckAgentInstructions(t *testing.T) {
	t.Run("no agent instruction files produces zero issues", func(t *testing.T) {
		tempDir := t.TempDir()
		report := model.NewReport()

		err := checkAgentInstructions(tempDir, report)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report.Issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(report.Issues))
		}
	})

	t.Run("valid AGENTS.md directing to FISS/INDEX.md", func(t *testing.T) {
		tempDir := t.TempDir()
		content := "# AI Agent Instructions\nRead [FISS Index](FISS/INDEX.md) first.\n"
		if err := os.WriteFile(filepath.Join(tempDir, "AGENTS.md"), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create AGENTS.md: %v", err)
		}

		report := model.NewReport()
		err := checkAgentInstructions(tempDir, report)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report.Issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(report.Issues))
		}
	})

	t.Run("invalid AGENTS.md not directing to FISS/INDEX.md", func(t *testing.T) {
		tempDir := t.TempDir()
		content := "# AI Agent Instructions\nJust write code and do not read anything.\n"
		if err := os.WriteFile(filepath.Join(tempDir, "AGENTS.md"), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create AGENTS.md: %v", err)
		}

		report := model.NewReport()
		err := checkAgentInstructions(tempDir, report)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report.Issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(report.Issues))
		}

		issue := report.Issues[0]
		if issue.RuleID != "FISS-R010" {
			t.Errorf("expected RuleID FISS-R010, got %s", issue.RuleID)
		}
		if issue.Severity != model.SeverityError {
			t.Errorf("expected SeverityError, got %v", issue.Severity)
		}
		if issue.FilePath != "AGENTS.md" {
			t.Errorf("expected FilePath AGENTS.md, got %s", issue.FilePath)
		}
	})

	t.Run("multiple files: CLAUDE.md valid and .cursorrules invalid", func(t *testing.T) {
		tempDir := t.TempDir()
		claudeContent := "Please navigate to FISS/INDEX.md before taking tasks."
		if err := os.WriteFile(filepath.Join(tempDir, "CLAUDE.md"), []byte(claudeContent), 0644); err != nil {
			t.Fatalf("failed to create CLAUDE.md: %v", err)
		}

		cursorContent := "Always run tests."
		if err := os.WriteFile(filepath.Join(tempDir, ".cursorrules"), []byte(cursorContent), 0644); err != nil {
			t.Fatalf("failed to create .cursorrules: %v", err)
		}

		report := model.NewReport()
		err := checkAgentInstructions(tempDir, report)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report.Issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(report.Issues))
		}

		if report.Issues[0].FilePath != ".cursorrules" {
			t.Errorf("expected issue on .cursorrules, got %s", report.Issues[0].FilePath)
		}
	})

	t.Run("multiple files: all invalid", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "AGENTS.md"), []byte("no link"), 0644); err != nil {
			t.Fatalf("failed to create AGENTS.md: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, "CLAUDE.md"), []byte("no link either"), 0644); err != nil {
			t.Fatalf("failed to create CLAUDE.md: %v", err)
		}

		report := model.NewReport()
		err := checkAgentInstructions(tempDir, report)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report.Issues) != 2 {
			t.Fatalf("expected 2 issues, got %d", len(report.Issues))
		}
		if report.Issues[0].RuleID != "FISS-R010" || report.Issues[1].RuleID != "FISS-R010" {
			t.Errorf("expected both issues to be FISS-R010")
		}
	})
}

func TestLinter_Lint_AgentInstructions(t *testing.T) {
	tempDir := t.TempDir()

	// Setup minimal valid FISS
	fissDir := filepath.Join(tempDir, "FISS")
	if err := os.Mkdir(fissDir, 0755); err != nil {
		t.Fatalf("failed to create FISS dir: %v", err)
	}

	indexContent := "# Index\n\n- [Baseline Context](BOOTSTRAP.md)\n  Read when: read always before beginning work.\n"
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("failed to create INDEX.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "BOOTSTRAP.md"), []byte("# Bootstrap\n"), 0644); err != nil {
		t.Fatalf("failed to create BOOTSTRAP.md: %v", err)
	}

	// 1. Without AGENTS.md -> passes cleanly
	l := New()
	report, err := l.Lint(tempDir)
	if err != nil {
		t.Fatalf("unexpected lint error: %v", err)
	}
	if report.HasErrors() {
		t.Fatalf("expected clean report, got errors: %v", report.Issues)
	}

	// 2. Add invalid AGENTS.md -> reports FISS-R010 error
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("# Rules\nDo something else.\n"), 0644); err != nil {
		t.Fatalf("failed to create AGENTS.md: %v", err)
	}

	report, err = l.Lint(tempDir)
	if err != nil {
		t.Fatalf("unexpected lint error: %v", err)
	}
	if !report.HasErrors() {
		t.Fatalf("expected FISS-R010 error, got no errors")
	}
	var foundR010 bool
	for _, issue := range report.Issues {
		if issue.RuleID == "FISS-R010" && issue.FilePath == "AGENTS.md" {
			foundR010 = true
			break
		}
	}
	if !foundR010 {
		t.Errorf("expected FISS-R010 issue on AGENTS.md, got issues: %v", report.Issues)
	}

	// 3. Fix AGENTS.md with valid reference -> passes cleanly
	validAgents := "# Rules\nRead [FISS Index](FISS/INDEX.md) before starting work.\n"
	if err := os.WriteFile(agentsPath, []byte(validAgents), 0644); err != nil {
		t.Fatalf("failed to update AGENTS.md: %v", err)
	}

	report, err = l.Lint(tempDir)
	if err != nil {
		t.Fatalf("unexpected lint error: %v", err)
	}
	if report.HasErrors() {
		t.Fatalf("expected clean report after fixing AGENTS.md, got errors: %v", report.Issues)
	}
}


