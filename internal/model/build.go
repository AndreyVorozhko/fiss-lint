package model

// BuildInfo encapsulates compile-time version and build metadata.
type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
}
