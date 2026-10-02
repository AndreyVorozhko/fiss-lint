package linter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildNavigationGraph_Linear(t *testing.T) {
	tempDir := t.TempDir()

	fissDir := filepath.Join(tempDir, "FISS")
	areaDir := filepath.Join(fissDir, "area")
	if err := os.MkdirAll(areaDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootIndex := `# Root Index
- [Bootstrap](BOOTSTRAP.md)
  Read when: always
- [Area](area/INDEX.md)
  Read when: working in area
`
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(rootIndex), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "BOOTSTRAP.md"), []byte("# Bootstrap\n"), 0644); err != nil {
		t.Fatal(err)
	}

	areaIndex := `# Area Index
- [Doc](doc.md)
  Read when: reading doc
`
	if err := os.WriteFile(filepath.Join(areaDir, "INDEX.md"), []byte(areaIndex), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(areaDir, "doc.md"), []byte("# Doc\n"), 0644); err != nil {
		t.Fatal(err)
	}

	graph, err := buildNavigationGraph(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedReachable := []string{
		"FISS/INDEX.md",
		"FISS/BOOTSTRAP.md",
		"FISS/area/INDEX.md",
		"FISS/area/doc.md",
	}

	for _, p := range expectedReachable {
		if !graph.IsReachable(p) {
			t.Errorf("expected path %q to be reachable, but it was not", p)
		}
	}

	if len(graph.VisitedIndexes) != 2 {
		t.Errorf("expected 2 visited indexes, got %d: %v", len(graph.VisitedIndexes), graph.VisitedIndexes)
	}
}

func TestBuildNavigationGraph_Cycles(t *testing.T) {
	tempDir := t.TempDir()

	fissDir := filepath.Join(tempDir, "FISS")
	sub1Dir := filepath.Join(fissDir, "sub1")
	sub2Dir := filepath.Join(fissDir, "sub2")
	if err := os.MkdirAll(sub1Dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sub2Dir, 0755); err != nil {
		t.Fatal(err)
	}

	// Root -> sub1/INDEX.md
	rootIndex := `# Root
- [Sub1](sub1/INDEX.md)
  Read when: sub1
`
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(rootIndex), 0644); err != nil {
		t.Fatal(err)
	}

	// sub1 -> sub2/INDEX.md and back to root ../INDEX.md
	sub1Index := `# Sub1
- [Sub2](../sub2/INDEX.md)
  Read when: sub2
- [Root](../INDEX.md)
  Read when: root
`
	if err := os.WriteFile(filepath.Join(sub1Dir, "INDEX.md"), []byte(sub1Index), 0644); err != nil {
		t.Fatal(err)
	}

	// sub2 -> sub1/INDEX.md (cycle between sub1 and sub2)
	sub2Index := `# Sub2
- [Sub1](../sub1/INDEX.md)
  Read when: sub1
`
	if err := os.WriteFile(filepath.Join(sub2Dir, "INDEX.md"), []byte(sub2Index), 0644); err != nil {
		t.Fatal(err)
	}

	// Should terminate cleanly despite cycles
	graph, err := buildNavigationGraph(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !graph.IsReachable("FISS/INDEX.md") || !graph.IsReachable("FISS/sub1/INDEX.md") || !graph.IsReachable("FISS/sub2/INDEX.md") {
		t.Errorf("expected all 3 indexes to be reachable, got %v", graph.ReachableFiles)
	}
	if len(graph.VisitedIndexes) != 3 {
		t.Errorf("expected exactly 3 visited indexes, got %d", len(graph.VisitedIndexes))
	}
}

func TestBuildNavigationGraph_DirectoryTarget(t *testing.T) {
	tempDir := t.TempDir()

	fissDir := filepath.Join(tempDir, "FISS")
	areaDir := filepath.Join(fissDir, "area")
	if err := os.MkdirAll(areaDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Root links to directory 'area/' without specifying INDEX.md
	rootIndex := `# Root
- [Area](area/)
  Read when: area
`
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(rootIndex), 0644); err != nil {
		t.Fatal(err)
	}

	areaIndex := `# Area
- [Doc](doc.md)
  Read when: doc
`
	if err := os.WriteFile(filepath.Join(areaDir, "INDEX.md"), []byte(areaIndex), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(areaDir, "doc.md"), []byte("# Doc\n"), 0644); err != nil {
		t.Fatal(err)
	}

	graph, err := buildNavigationGraph(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !graph.IsReachable("FISS/area/INDEX.md") {
		t.Errorf("expected FISS/area/INDEX.md to be reached via directory target")
	}
	if !graph.IsReachable("FISS/area/doc.md") {
		t.Errorf("expected FISS/area/doc.md to be reached via area index")
	}
}

func TestBuildNavigationGraph_SymlinkLoop(t *testing.T) {
	tempDir := t.TempDir()

	fissDir := filepath.Join(tempDir, "FISS")
	if err := os.MkdirAll(fissDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create symlink FISS/loop -> FISS
	loopDir := filepath.Join(fissDir, "loop")
	if err := os.Symlink(fissDir, loopDir); err != nil {
		t.Skip("symlinks not supported in environment, skipping")
	}

	// Root links to loop/INDEX.md
	rootIndex := `# Root
- [Loop](loop/INDEX.md)
  Read when: loop
`
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(rootIndex), 0644); err != nil {
		t.Fatal(err)
	}

	// Must terminate cleanly without infinite loop
	graph, err := buildNavigationGraph(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !graph.IsReachable("FISS/INDEX.md") {
		t.Errorf("expected FISS/INDEX.md to be reachable")
	}
}

func TestBuildNavigationGraph_MissingRootIndex(t *testing.T) {
	tempDir := t.TempDir()
	graph, err := buildNavigationGraph(tempDir)
	if err != nil {
		t.Fatalf("unexpected error when FISS/INDEX.md missing: %v", err)
	}
	if len(graph.ReachableFiles) != 0 {
		t.Errorf("expected empty reachable files, got: %v", graph.ReachableFiles)
	}
}
