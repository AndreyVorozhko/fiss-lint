package linter

import (
	"strings"
	"testing"
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
