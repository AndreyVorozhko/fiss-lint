package linter

import (
	"os"
	"path/filepath"
	"testing"

	"fiss-lint/internal/model"
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

func TestCheckReachability_AllReachable(t *testing.T) {
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
		t.Fatalf("unexpected graph error: %v", err)
	}

	report := model.NewReport()
	if err := checkReachability(tempDir, graph.ReachableFiles, report); err != nil {
		t.Fatalf("unexpected checkReachability error: %v", err)
	}

	if report.ErrorsCount() != 0 {
		t.Errorf("expected 0 errors, got %d: %v", report.ErrorsCount(), report.Issues)
	}
}

func TestCheckReachability_OrphanFiles(t *testing.T) {
	tempDir := t.TempDir()
	fissDir := filepath.Join(tempDir, "FISS")
	subDir := filepath.Join(fissDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootIndex := `# Root Index
- [Bootstrap](BOOTSTRAP.md)
  Read when: always
`
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(rootIndex), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "BOOTSTRAP.md"), []byte("# Bootstrap\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Orphan files
	if err := os.WriteFile(filepath.Join(fissDir, "orphan.md"), []byte("# Orphan\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "unlinked.md"), []byte("# Unlinked\n"), 0644); err != nil {
		t.Fatal(err)
	}

	graph, err := buildNavigationGraph(tempDir)
	if err != nil {
		t.Fatalf("unexpected graph error: %v", err)
	}

	report := model.NewReport()
	if err := checkReachability(tempDir, graph.ReachableFiles, report); err != nil {
		t.Fatalf("unexpected checkReachability error: %v", err)
	}

	if report.ErrorsCount() != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", report.ErrorsCount(), report.Issues)
	}

	foundOrphan := false
	foundUnlinked := false
	for _, issue := range report.Issues {
		if issue.RuleID != "FISS-R008" {
			t.Errorf("expected rule FISS-R008, got %s", issue.RuleID)
		}
		if issue.Severity != model.SeverityError {
			t.Errorf("expected error severity, got %v", issue.Severity)
		}
		if issue.FilePath == "FISS/orphan.md" {
			foundOrphan = true
		}
		if issue.FilePath == "FISS/sub/unlinked.md" {
			foundUnlinked = true
		}
	}

	if !foundOrphan {
		t.Errorf("expected orphan issue for FISS/orphan.md")
	}
	if !foundUnlinked {
		t.Errorf("expected orphan issue for FISS/sub/unlinked.md")
	}
}

func TestCheckReachability_UnlinkedAreaIndex(t *testing.T) {
	tempDir := t.TempDir()
	fissDir := filepath.Join(tempDir, "FISS")
	areaDir := filepath.Join(fissDir, "isolated_area")
	if err := os.MkdirAll(areaDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootIndex := `# Root Index
- [Bootstrap](BOOTSTRAP.md)
  Read when: always
`
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(rootIndex), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "BOOTSTRAP.md"), []byte("# Bootstrap\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// isolated_area has INDEX.md and doc.md, but is not linked from Root Index
	if err := os.WriteFile(filepath.Join(areaDir, "INDEX.md"), []byte("# Isolated Area\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(areaDir, "doc.md"), []byte("# Doc\n"), 0644); err != nil {
		t.Fatal(err)
	}

	graph, err := buildNavigationGraph(tempDir)
	if err != nil {
		t.Fatalf("unexpected graph error: %v", err)
	}

	report := model.NewReport()
	if err := checkReachability(tempDir, graph.ReachableFiles, report); err != nil {
		t.Fatalf("unexpected checkReachability error: %v", err)
	}

	if report.ErrorsCount() != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", report.ErrorsCount(), report.Issues)
	}
	for _, issue := range report.Issues {
		if issue.RuleID != "FISS-R008" {
			t.Errorf("expected rule FISS-R008, got %s", issue.RuleID)
		}
	}
}

func TestCheckReachability_IgnoreNonMarkdownAndHidden(t *testing.T) {
	tempDir := t.TempDir()
	fissDir := filepath.Join(tempDir, "FISS")
	hiddenDir := filepath.Join(fissDir, ".hidden")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootIndex := `# Root Index
- [Bootstrap](BOOTSTRAP.md)
  Read when: always
`
	if err := os.WriteFile(filepath.Join(fissDir, "INDEX.md"), []byte(rootIndex), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "BOOTSTRAP.md"), []byte("# Bootstrap\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Hidden file, non-markdown file, and file inside hidden folder
	if err := os.WriteFile(filepath.Join(fissDir, ".gitkeep"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fissDir, "image.png"), []byte("binary"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "secret.md"), []byte("# Secret\n"), 0644); err != nil {
		t.Fatal(err)
	}

	graph, err := buildNavigationGraph(tempDir)
	if err != nil {
		t.Fatalf("unexpected graph error: %v", err)
	}

	report := model.NewReport()
	if err := checkReachability(tempDir, graph.ReachableFiles, report); err != nil {
		t.Fatalf("unexpected checkReachability error: %v", err)
	}

	if report.ErrorsCount() != 0 {
		t.Errorf("expected 0 errors, got %d: %v", report.ErrorsCount(), report.Issues)
	}
}

func TestCheckReachability_MissingFissDir(t *testing.T) {
	tempDir := t.TempDir()
	report := model.NewReport()
	if err := checkReachability(tempDir, nil, report); err != nil {
		t.Fatalf("unexpected error on missing FISS dir: %v", err)
	}
	if report.ErrorsCount() != 0 {
		t.Errorf("expected 0 errors, got %d", report.ErrorsCount())
	}
}

