package main

import (
	"os"

	"fiss-lint/internal/cli"
	"fiss-lint/internal/model"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	info := model.BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	}
	code := cli.Run(os.Args[1:], os.Stdout, os.Stderr, info)
	os.Exit(code)
}
