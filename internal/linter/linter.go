package linter

// Linter represents the FISS rule evaluation engine.
type Linter struct{}

// New creates a new Linter instance.
func New() *Linter {
	return &Linter{}
}
