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
			name:         "self project repository root",
			args:         []string{"../.."},
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
