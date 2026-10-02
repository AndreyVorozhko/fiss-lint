package cli

import (
	"bytes"
	"strings"
	"testing"

	"fiss-lint/internal/model"
)

func TestRun(t *testing.T) {
	testBuildInfo := model.BuildInfo{
		Version:   "1.2.3",
		Commit:    "abcdef0",
		BuildDate: "2026-09-30T12:00:00Z",
	}

	tests := []struct {
		name           string
		args           []string
		info           model.BuildInfo
		expectedCode   int
		stdoutContains []string
		stderrContains []string
	}{
		{
			name:         "help flag short -h",
			args:         []string{"-h"},
			info:         testBuildInfo,
			expectedCode: 0,
			stdoutContains: []string{
				"Usage:",
				"-h, --help",
				"-v, --version",
				"--format",
				"--strict",
				"FISS-R001",
				"FISS-R018",
			},
		},
		{
			name:         "help flag long --help",
			args:         []string{"--help"},
			info:         testBuildInfo,
			expectedCode: 0,
			stdoutContains: []string{
				"Usage:",
				"FISS Rules Summary:",
				"--format",
				"--strict",
				"FISS-R001",
				"FISS-R018",
			},
		},
		{
			name:         "version flag short -v",
			args:         []string{"-v"},
			info:         testBuildInfo,
			expectedCode: 0,
			stdoutContains: []string{
				"fiss-lint version 1.2.3 (commit: abcdef0, built: 2026-09-30T12:00:00Z)",
			},
		},
		{
			name:         "version flag long --version",
			args:         []string{"--version"},
			info:         testBuildInfo,
			expectedCode: 0,
			stdoutContains: []string{
				"fiss-lint version 1.2.3 (commit: abcdef0, built: 2026-09-30T12:00:00Z)",
			},
		},
		{
			name:         "invalid flag",
			args:         []string{"--non-existent-flag"},
			info:         testBuildInfo,
			expectedCode: 2,
			stderrContains: []string{
				"Error: flag provided but not defined: -non-existent-flag",
			},
		},
		{
			name:         "default invocation without args in dir without FISS",
			args:         []string{},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R001] .: directory FISS/ not found in project root",
			},
		},
		{
			name:         "fixture valid_minimal",
			args:         []string{"../../testdata/valid_minimal"},
			info:         testBuildInfo,
			expectedCode: 0,
		},
		{
			name:         "fixture missing_fiss",
			args:         []string{"../../testdata/missing_fiss"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R001] .: directory FISS/ not found in project root",
			},
		},
		{
			name:         "fixture missing_bootstrap",
			args:         []string{"../../testdata/missing_bootstrap"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R002] FISS: required file BOOTSTRAP.md not found",
			},
		},
		{
			name:         "fixture missing_links",
			args:         []string{"../../testdata/missing_links"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R003] FISS/INDEX.md: missing link to BOOTSTRAP.md",
				"[WARNING] [FISS-R011] FISS/INDEX.md: recommended link to official standard https://fiss.vorozhko.ru not found",
			},
		},
		{
			name:         "fixture invalid_nav_single_line",
			args:         []string{"../../testdata/invalid_nav_single_line"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R005] FISS/INDEX.md:7: missing 'Read when:' condition for navigation entry",
			},
		},
		{
			name:         "fixture invalid_nav_indent",
			args:         []string{"../../testdata/invalid_nav_indent"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R005] FISS/INDEX.md:8: invalid indentation for 'Read when:' condition (expected exactly 2 spaces)",
			},
		},
		{
			name:         "fixture invalid_nav_marker",
			args:         []string{"../../testdata/invalid_nav_marker"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R005] FISS/INDEX.md:8: invalid read condition marker (must be exact 'Read when:')",
			},
		},
		{
			name:         "fixture invalid_nav_empty_condition",
			args:         []string{"../../testdata/invalid_nav_empty_condition"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R005] FISS/INDEX.md:8: empty condition text in 'Read when:'",
			},
		},
		{
			name:         "fixture missing_bootstrap_condition",
			args:         []string{"../../testdata/missing_bootstrap_condition"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R004] FISS/INDEX.md:3: link to BOOTSTRAP.md must have an attached read condition",
				"[ERROR] [FISS-R005] FISS/INDEX.md:3: missing 'Read when:' condition for navigation entry",
			},
		},
		{
			name:         "fixture valid_links",
			args:         []string{"../../testdata/valid_links"},
			info:         testBuildInfo,
			expectedCode: 0,
		},
		{
			name:         "fixture invalid_link_not_found",
			args:         []string{"../../testdata/invalid_link_not_found"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R006] FISS/INDEX.md:5: target file does not exist: missing-doc.md",
			},
		},
		{
			name:         "fixture invalid_bare_directory",
			args:         []string{"../../testdata/invalid_bare_directory"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R006] FISS/INDEX.md:5: link targets a bare directory without INDEX.md: bare_dir/",
			},
		},
		{
			name:         "fixture invalid_non_markdown",
			args:         []string{"../../testdata/invalid_non_markdown"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R006] FISS/INDEX.md:5: link target must be a Markdown file (.md): document.txt",
			},
		},
		{
			name:         "fixture invalid_link_case",
			args:         []string{"../../testdata/invalid_link_case"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R006] FISS/INDEX.md:5: target file does not exist: guide.md",
			},
		},
		{
			name:         "self project repository root",
			args:         []string{"../.."},
			info:         testBuildInfo,
			expectedCode: 0,
		},
		{
			name:         "fixture valid_topology",
			args:         []string{"../../testdata/valid_topology"},
			info:         testBuildInfo,
			expectedCode: 0,
		},
		{
			name:         "fixture invalid_orphan_file",
			args:         []string{"../../testdata/invalid_orphan_file"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R008] FISS/orphan.md: unreachable markdown file (orphan)",
			},
		},
		{
			name:         "fixture invalid_composite_area_missing_index",
			args:         []string{"../../testdata/invalid_composite_area_missing_index"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R007] FISS/INDEX.md:5: composite area missing INDEX.md: FISS/empty_area",
			},
		},
		{
			name:         "fixture invalid_overrides_missing_index",
			args:         []string{"../../testdata/invalid_overrides_missing_index"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R009] FISS/overrides: required file INDEX.md not found in overrides directory",
			},
		},
		{
			name:         "fixture invalid_overrides_unlinked",
			args:         []string{"../../testdata/invalid_overrides_unlinked"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R009] FISS/INDEX.md: missing link to overrides/INDEX.md",
			},
		},
		{
			name:         "fixture invalid_overrides_condition",
			args:         []string{"../../testdata/invalid_overrides_condition"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R009] FISS/INDEX.md:6: read condition for overrides must require reading before skill usage",
			},
		},
		{
			name:         "fixture valid_agents",
			args:         []string{"../../testdata/valid_agents"},
			info:         testBuildInfo,
			expectedCode: 0,
		},
		{
			name:         "fixture valid_no_agents",
			args:         []string{"../../testdata/valid_no_agents"},
			info:         testBuildInfo,
			expectedCode: 0,
		},
		{
			name:         "fixture invalid_agents_no_link",
			args:         []string{"../../testdata/invalid_agents_no_link"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R010] AGENTS.md: does not direct to FISS/INDEX.md",
			},
		},
		{
			name:         "fixture invalid_claude_no_link",
			args:         []string{"../../testdata/invalid_claude_no_link"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R010] CLAUDE.md: does not direct to FISS/INDEX.md",
			},
		},
		{
			name:         "fixture invalid_cursorrules_no_link",
			args:         []string{"../../testdata/invalid_cursorrules_no_link"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				"[ERROR] [FISS-R010] .cursorrules: does not direct to FISS/INDEX.md",
			},
		},
		{
			name:         "invalid format flag value",
			args:         []string{"--format", "invalid"},
			info:         testBuildInfo,
			expectedCode: 2,
			stderrContains: []string{
				`Error: invalid format "invalid", must be "text" or "json"`,
			},
		},
		{
			name:         "format json on valid project",
			args:         []string{"--format", "json", "../../testdata/valid_minimal"},
			info:         testBuildInfo,
			expectedCode: 0,
			stdoutContains: []string{
				`"issues": []`,
				`"summary": {`,
				`"errors": 0`,
				`"warnings": 0`,
				`"total": 0`,
			},
		},
		{
			name:         "format json on invalid project",
			args:         []string{"--format", "json", "../../testdata/invalid_agents_no_link"},
			info:         testBuildInfo,
			expectedCode: 1,
			stdoutContains: []string{
				`"rule_id": "FISS-R010"`,
				`"severity": "ERROR"`,
				`"file": "AGENTS.md"`,
				`"errors": 1`,
			},
		},
		{
			name:         "strict flag on valid project",
			args:         []string{"--strict", "../../testdata/valid_minimal"},
			info:         testBuildInfo,
			expectedCode: 0,
		},
		{
			name:         "non-existent directory target",
			args:         []string{"../../testdata/does_not_exist"},
			info:         testBuildInfo,
			expectedCode: 1,
			stderrContains: []string{
				"Error: reading project root:",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := Run(tc.args, &stdout, &stderr, tc.info)

			if code != tc.expectedCode {
				t.Errorf("Run() code = %d, expected %d", code, tc.expectedCode)
			}

			stdoutStr := stdout.String()
			for _, expected := range tc.stdoutContains {
				if !strings.Contains(stdoutStr, expected) {
					t.Errorf("stdout does not contain %q, got:\n%s", expected, stdoutStr)
				}
			}

			stderrStr := stderr.String()
			for _, expected := range tc.stderrContains {
				if !strings.Contains(stderrStr, expected) {
					t.Errorf("stderr does not contain %q, got:\n%s", expected, stderrStr)
				}
			}
		})
	}
}
