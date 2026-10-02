package linter

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fiss-lint/internal/model"
)

func TestIsExternalURL(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected bool
	}{
		{"http_url", "http://example.com", true},
		{"https_url", "https://fiss.vorozhko.ru/v1.0.0/llms.txt", true},
		{"https_uppercase", "HTTPS://VOROZHKO.RU", true},
		{"mailto_scheme", "mailto:dron@example.com", true},
		{"ftp_scheme", "ftp://files.example.com/archive.zip", true},
		{"custom_scheme", "git://github.com/user/repo.git", true},
		{"local_file", "BOOTSTRAP.md", false},
		{"local_subpath", "overrides/INDEX.md", false},
		{"local_relative_dot", "./BOOTSTRAP.md", false},
		{"local_relative_parent", "../BOOTSTRAP.md", false},
		{"anchor_only", "#section", false},
		{"empty_string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExternalURL(tt.target)
			if got != tt.expected {
				t.Errorf("isExternalURL(%q) = %v; want %v", tt.target, got, tt.expected)
			}
		})
	}
}

func TestStripAnchorAndQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain_filename", "BOOTSTRAP.md", "BOOTSTRAP.md"},
		{"with_anchor", "rules.md#fiss-r006", "rules.md"},
		{"with_query", "rules.md?raw=true", "rules.md"},
		{"with_query_and_anchor", "rules.md?raw=true#section", "rules.md"},
		{"anchor_only", "#fiss-r006", ""},
		{"query_only", "?version=1", ""},
		{"with_spaces", "  rules.md#anchor  ", "rules.md"},
		{"nested_with_anchor", "knowledge/subject/rules.md#title", "knowledge/subject/rules.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAnchorAndQuery(tt.input)
			if got != tt.expected {
				t.Errorf("stripAnchorAndQuery(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveRelativeTarget(t *testing.T) {
	fakeRoot := filepath.FromSlash("/home/user/project")

	tests := []struct {
		name            string
		indexRelPath    string
		target          string
		wantRelFromRoot string
		wantFullPath    string
		wantSelfAnchor  bool
		wantErr         error
	}{
		{
			name:            "same_directory_file",
			indexRelPath:    "FISS/INDEX.md",
			target:          "BOOTSTRAP.md",
			wantRelFromRoot: "FISS/BOOTSTRAP.md",
			wantFullPath:    filepath.Join(fakeRoot, "FISS", "BOOTSTRAP.md"),
			wantSelfAnchor:  false,
			wantErr:         nil,
		},
		{
			name:            "explicit_dot_prefix",
			indexRelPath:    "FISS/INDEX.md",
			target:          "./BOOTSTRAP.md",
			wantRelFromRoot: "FISS/BOOTSTRAP.md",
			wantFullPath:    filepath.Join(fakeRoot, "FISS", "BOOTSTRAP.md"),
			wantSelfAnchor:  false,
			wantErr:         nil,
		},
		{
			name:            "subfolder_path",
			indexRelPath:    "FISS/INDEX.md",
			target:          "overrides/INDEX.md",
			wantRelFromRoot: "FISS/overrides/INDEX.md",
			wantFullPath:    filepath.Join(fakeRoot, "FISS", "overrides", "INDEX.md"),
			wantSelfAnchor:  false,
			wantErr:         nil,
		},
		{
			name:            "parent_traversal_within_root",
			indexRelPath:    "FISS/human/knowledge/INDEX.md",
			target:          "../../state/INDEX.md",
			wantRelFromRoot: "FISS/state/INDEX.md",
			wantFullPath:    filepath.Join(fakeRoot, "FISS", "state", "INDEX.md"),
			wantSelfAnchor:  false,
			wantErr:         nil,
		},
		{
			name:            "target_with_anchor",
			indexRelPath:    "FISS/INDEX.md",
			target:          "knowledge/subject/rules.md#fiss-r006",
			wantRelFromRoot: "FISS/knowledge/subject/rules.md",
			wantFullPath:    filepath.Join(fakeRoot, "FISS", "knowledge", "subject", "rules.md"),
			wantSelfAnchor:  false,
			wantErr:         nil,
		},
		{
			name:            "anchor_only_self_reference",
			indexRelPath:    "FISS/INDEX.md",
			target:          "#table-of-contents",
			wantRelFromRoot: "FISS/INDEX.md",
			wantFullPath:    filepath.Join(fakeRoot, "FISS", "INDEX.md"),
			wantSelfAnchor:  true,
			wantErr:         nil,
		},
		{
			name:            "empty_target_error",
			indexRelPath:    "FISS/INDEX.md",
			target:          "   ",
			wantRelFromRoot: "",
			wantFullPath:    "",
			wantSelfAnchor:  false,
			wantErr:         ErrTargetEmpty,
		},
		{
			name:            "target_escaping_root",
			indexRelPath:    "FISS/INDEX.md",
			target:          "../../outside.md",
			wantRelFromRoot: "",
			wantFullPath:    "",
			wantSelfAnchor:  false,
			wantErr:         ErrTargetEscapesRoot,
		},
		{
			name:            "absolute_unix_path_rejected",
			indexRelPath:    "FISS/INDEX.md",
			target:          "/etc/passwd",
			wantRelFromRoot: "",
			wantFullPath:    "",
			wantSelfAnchor:  false,
			wantErr:         ErrTargetAbsolute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := resolveRelativeTarget(fakeRoot, tt.indexRelPath, tt.target)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("resolveRelativeTarget() error = nil; want error wrapping %v", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("resolveRelativeTarget() error = %v; want errors.Is %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolveRelativeTarget() unexpected error: %v", err)
			}

			if resolved.RelPathFromRoot != tt.wantRelFromRoot {
				t.Errorf("RelPathFromRoot = %q; want %q", resolved.RelPathFromRoot, tt.wantRelFromRoot)
			}
			if resolved.FullPath != tt.wantFullPath {
				t.Errorf("FullPath = %q; want %q", resolved.FullPath, tt.wantFullPath)
			}
			if resolved.IsSelfAnchor != tt.wantSelfAnchor {
				t.Errorf("IsSelfAnchor = %v; want %v", resolved.IsSelfAnchor, tt.wantSelfAnchor)
			}
		})
	}
}

func TestCheckPathExistsCaseSensitive(t *testing.T) {
	tempDir := t.TempDir()

	// Setup structure:
	// tempDir/
	// ├── FISS/
	// │   ├── INDEX.md
	// │   ├── BOOTSTRAP.md
	// │   ├── overrides/
	// │   │   └── INDEX.md
	// │   └── notes.txt
	fissDir := filepath.Join(tempDir, "FISS")
	if err := os.MkdirAll(filepath.Join(fissDir, "overrides"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte("# Index\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "BOOTSTRAP.md"), []byte("# Bootstrap\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "overrides", "INDEX.md"), []byte("# Overrides Index\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "notes.txt"), []byte("non-md file\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		relPath    string
		wantExists bool
		wantIsDir  bool
	}{
		{"exact_file_root", "FISS/INDEX.md", true, false},
		{"exact_file_bootstrap", "FISS/BOOTSTRAP.md", true, false},
		{"exact_nested_file", "FISS/overrides/INDEX.md", true, false},
		{"exact_directory", "FISS/overrides", true, true},
		{"non_markdown_file", "FISS/notes.txt", true, false},
		{"wrong_case_file_bootstrap", "FISS/bootstrap.md", false, false},
		{"wrong_case_file_index", "FISS/index.md", false, false},
		{"wrong_case_dir", "fiss/INDEX.md", false, false},
		{"wrong_case_nested_dir", "FISS/Overrides/INDEX.md", false, false},
		{"non_existent_file", "FISS/missing.md", false, false},
		{"non_existent_dir", "FISS/missing_dir", false, false},
		{"project_root_itself", ".", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, isDir, err := checkPathExistsCaseSensitive(tempDir, tt.relPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exists != tt.wantExists {
				t.Errorf("checkPathExistsCaseSensitive(%q) exists = %v; want %v", tt.relPath, exists, tt.wantExists)
			}
			if isDir != tt.wantIsDir {
				t.Errorf("checkPathExistsCaseSensitive(%q) isDir = %v; want %v", tt.relPath, isDir, tt.wantIsDir)
			}
		})
	}
}

func TestValidateIndexLinksIntegrity(t *testing.T) {
	tempDir := t.TempDir()

	fissDir := filepath.Join(tempDir, "FISS")
	if err := os.MkdirAll(fissDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte("# Index\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "BOOTSTRAP.md"), []byte("# Bootstrap\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "notes.txt"), []byte("text file\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		indexRelPath  string
		entry         NavEntry
		wantIssueCode string
		wantInMsg     string
	}{
		{
			name:          "valid_existing_bootstrap_link",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 5, Title: "Bootstrap", Target: "BOOTSTRAP.md"},
			wantIssueCode: "",
		},
		{
			name:          "valid_external_https_link",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 7, Title: "Standard", Target: "https://fiss.vorozhko.ru/v1.0.0/llms.txt"},
			wantIssueCode: "",
		},
		{
			name:          "valid_self_anchor_link",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 9, Title: "Top", Target: "#heading"},
			wantIssueCode: "",
		},
		{
			name:          "missing_file_triggers_fiss_r006",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 12, Title: "Missing", Target: "missing.md"},
			wantIssueCode: "FISS-R006",
			wantInMsg:     "target file does not exist: missing.md",
		},
		{
			name:          "wrong_case_file_triggers_fiss_r006",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 15, Title: "Bootstrap", Target: "bootstrap.md"},
			wantIssueCode: "FISS-R006",
			wantInMsg:     "target file does not exist: bootstrap.md",
		},
		{
			name:          "non_markdown_file_triggers_fiss_r006",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 18, Title: "Notes", Target: "notes.txt"},
			wantIssueCode: "FISS-R006",
			wantInMsg:     "link target must be a Markdown file (.md): notes.txt",
		},
		{
			name:          "escaping_root_triggers_fiss_r006",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 20, Title: "Escape", Target: "../../outside.md"},
			wantIssueCode: "FISS-R006",
			wantInMsg:     "link target escapes project root",
		},
		{
			name:          "absolute_path_triggers_fiss_r006",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 22, Title: "Abs", Target: "/etc/passwd"},
			wantIssueCode: "FISS-R006",
			wantInMsg:     "link target must not be absolute",
		},
		{
			name:          "empty_target_triggers_fiss_r006",
			indexRelPath:  "FISS/INDEX.md",
			entry:         NavEntry{Line: 25, Title: "Empty", Target: ""},
			wantIssueCode: "FISS-R006",
			wantInMsg:     "link target is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := model.NewReport()
			validateIndexLinksIntegrity(tempDir, tt.indexRelPath, []NavEntry{tt.entry}, report)

			if tt.wantIssueCode == "" {
				if report.HasErrors() {
					t.Fatalf("expected no errors, got: %v", report.Issues)
				}
				return
			}

			if !report.HasErrors() {
				t.Fatalf("expected error %s, got none", tt.wantIssueCode)
			}

			var matched bool
			for _, issue := range report.Issues {
				if issue.RuleID == tt.wantIssueCode && issue.Line == tt.entry.Line {
					matched = true
					if tt.wantInMsg != "" && !strings.Contains(issue.Message, tt.wantInMsg) {
						t.Errorf("issue message %q does not contain %q", issue.Message, tt.wantInMsg)
					}
				}
			}

			if !matched {
				t.Fatalf("expected issue with RuleID=%s on line %d, got: %v", tt.wantIssueCode, tt.entry.Line, report.Issues)
			}
		})
	}
}
