package linter

import (
	"errors"
	"path/filepath"
	"testing"
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
