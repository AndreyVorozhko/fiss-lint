package linter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
