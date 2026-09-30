package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"fiss-lint/internal/model"
)

const helpText = `fiss-lint - File-based Intellectual Space Standard (FISS) linter

Usage:
  fiss-lint [flags] [path]

Flags:
  -h, --help      Show help and rules summary
  -v, --version   Show version information

FISS Rules Summary:
  FISS-R001  [Error]    Root Structure: The project root MUST contain the FISS/ directory.
  FISS-R002  [Error]    Root Structure: The FISS/ directory MUST contain INDEX.md and BOOTSTRAP.md.
  FISS-R003  [Error]    Navigation: FISS/INDEX.md MUST contain a link targeting BOOTSTRAP.md.
  FISS-R004  [Error]    Navigation: The link to BOOTSTRAP.md MUST have an attached read condition requiring reading before project work.
  FISS-R005  [Error]    Syntax: Every navigation entry in any INDEX.md MUST follow the two-line format (- [Title](path) + Read when:).
  FISS-R006  [Error]    Link Integrity: Every relative link in an index MUST resolve to an existing physical .md file or INDEX.md of a composite area.
  FISS-R007  [Error]    Topology: A composite area MUST contain its own INDEX.md.
  FISS-R008  [Error]    Topology: Every used area MUST be reachable from FISS/INDEX.md through indexes.
  FISS-R009  [Error]    Overrides Entry: If FISS/overrides/ exists, it MUST contain INDEX.md, and FISS/INDEX.md MUST link to FISS/overrides/INDEX.md.
  FISS-R010  [Error]    Agent Entry: If AGENTS.md exists, it MUST direct agents to FISS/INDEX.md.
  FISS-R011  [Warning]  Standard Reference: FISS/INDEX.md SHOULD contain a reference to the official standard website (https://fiss.vorozhko.ru).
  FISS-R012  [Error]    Knowledge Partitioning: FISS/knowledge/ MUST contain strictly subject/ and/or project/ areas.
  FISS-R013  [Error]    Human Partitioning: FISS/human/ MUST contain strictly knowledge/ and/or hmm/ areas.
  FISS-R014  [Error]    State Registry: Registries in FISS/state/ MUST be composite areas with INDEX.md or consolidated single files.
  FISS-R015  [Error]    Derivation Syntax: Derived knowledge MUST use Derived from: with resolving Markdown source links.
  FISS-R016  [Error]    HMM Derivation: Every content material in FISS/human/hmm/ MUST be explicitly marked as derived.
  FISS-R017  [Error]    Override Routing: Override navigation in FISS/overrides/ MUST be organized by the subject of a rule.
  FISS-R018  [Error]    Handoff Validity: The resolved handoff record MUST declare a valid synchronization state and work item.
`

// Run parses command-line arguments and executes the requested CLI action.
// It returns the process exit code.
func Run(args []string, stdout, stderr io.Writer, info model.BuildInfo) int {
	fs := flag.NewFlagSet("fiss-lint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var showHelp bool
	var showVersion bool

	fs.BoolVar(&showHelp, "h", false, "Show help and rules summary")
	fs.BoolVar(&showHelp, "help", false, "Show help and rules summary")
	fs.BoolVar(&showVersion, "v", false, "Show version information")
	fs.BoolVar(&showVersion, "version", false, "Show version information")

	fs.Usage = func() {
		// Suppress default FlagSet output so we can format help to stdout.
	}

	err := fs.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, helpText)
			return 0
		}
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 2
	}

	if showHelp {
		fmt.Fprint(stdout, helpText)
		return 0
	}

	if showVersion {
		fmt.Fprintf(stdout, "fiss-lint version %s (commit: %s, built: %s)\n", info.Version, info.Commit, info.BuildDate)
		return 0
	}

	targetPath := "."
	if fs.NArg() > 0 {
		targetPath = fs.Arg(0)
	}
	_ = targetPath

	return 0
}
