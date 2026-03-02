package review

import (
	"testing"
)

func TestCollectChangedFuncs_FiltersUnchanged(t *testing.T) {
	// Without a project, collectChangedFuncs returns nil
	result := collectChangedFuncs(nil, []*GitDiffFile{
		{Path: "foo.go", Content: "some diff"},
	})
	if result != nil {
		t.Errorf("expected nil without project, got %d", len(result))
	}
}

func TestCollectChangedFuncs_SkipsNonGo(t *testing.T) {
	result := collectChangedFuncs(nil, []*GitDiffFile{
		{Path: "config.yaml", Content: "some diff"},
	})
	if result != nil {
		t.Errorf("expected nil for non-Go file, got %d", len(result))
	}
}

func TestFindEntries_NoEdges(t *testing.T) {
	// All functions are isolated — all are entries
	graph := &callGraph{
		changedSet: map[string]ChangedFunc{
			"pkg.A": {Package: "pkg", Name: "A"},
			"pkg.B": {Package: "pkg", Name: "B"},
		},
	}

	entries := findEntries(graph)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestFindEntries_WithEdge(t *testing.T) {
	graph := &callGraph{
		changedSet: map[string]ChangedFunc{
			"pkg.A": {Package: "pkg", Name: "A"},
			"pkg.B": {Package: "pkg", Name: "B"},
		},
		edges: []callEdge{
			{from: "pkg.A", to: "pkg.B"},
		},
	}

	entries := findEntries(graph)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "A" {
		t.Errorf("expected entry A, got %s", entries[0].Name)
	}
}

func TestFindEntries_Cycle(t *testing.T) {
	// A → B, B → A — both have in-degree 1, both are NOT entries
	// In practice, buildCallGraph uses visited sets to break cycles,
	// but if edges exist, in-degree is tracked
	graph := &callGraph{
		changedSet: map[string]ChangedFunc{
			"pkg.A": {Package: "pkg", Name: "A"},
			"pkg.B": {Package: "pkg", Name: "B"},
		},
		edges: []callEdge{
			{from: "pkg.A", to: "pkg.B"},
			{from: "pkg.B", to: "pkg.A"},
		},
	}

	entries := findEntries(graph)
	// Both have in-degree > 0, neither is entry
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries in cycle, got %d", len(entries))
	}
}

func TestBuildFlowRoots_AllIsolated(t *testing.T) {
	graph := &callGraph{
		changedSet: map[string]ChangedFunc{
			"pkg.A": {Package: "pkg", Name: "A", Code: "func A() {}"},
			"pkg.B": {Package: "pkg", Name: "B", Code: "func B() {}"},
		},
	}
	funcs := []ChangedFunc{
		{Package: "pkg", Name: "A", Code: "func A() {}"},
		{Package: "pkg", Name: "B", Code: "func B() {}"},
	}

	roots := buildFlowRoots(nil, graph, funcs)
	if len(roots) != 2 {
		t.Fatalf("expected 2 isolated roots, got %d", len(roots))
	}
	for _, r := range roots {
		if len(r.Children) != 0 {
			t.Errorf("isolated root %s should have no children, got %d", r.Label, len(r.Children))
		}
	}
}

func TestBuildFlowRoots_ChainFlow(t *testing.T) {
	cfA := ChangedFunc{Package: "pkg", Name: "A", Code: "func A() { B() }", NodeType: "function"}
	cfB := ChangedFunc{Package: "pkg", Name: "B", Code: "func B() {}", NodeType: "function"}

	graph := &callGraph{
		changedSet: map[string]ChangedFunc{
			"pkg.A": cfA,
			"pkg.B": cfB,
		},
		edges: []callEdge{
			{from: "pkg.A", to: "pkg.B"},
		},
	}
	funcs := []ChangedFunc{cfA, cfB}

	roots := buildFlowRoots(nil, graph, funcs)
	if len(roots) != 1 {
		t.Fatalf("expected 1 chain flow root, got %d", len(roots))
	}

	flow := roots[0]
	// Flow root should have children: entry + target
	if len(flow.Children) < 2 {
		t.Fatalf("expected at least 2 children in chain flow, got %d", len(flow.Children))
	}
}

func TestBuildFlowRoots_WithBridge(t *testing.T) {
	cfA := ChangedFunc{Package: "pkg", Name: "A", Code: "func A() {}", NodeType: "function"}
	cfC := ChangedFunc{Package: "pkg", Name: "C", Code: "func C() {}", NodeType: "function"}

	graph := &callGraph{
		changedSet: map[string]ChangedFunc{
			"pkg.A": cfA,
			"pkg.C": cfC,
		},
		edges: []callEdge{
			{
				from: "pkg.A",
				to:   "pkg.C",
				bridges: []bridgeFunc{
					{Package: "pkg", Name: "B", Code: "func B() { C() }", NodeType: "function"},
				},
			},
		},
	}
	funcs := []ChangedFunc{cfA, cfC}

	roots := buildFlowRoots(nil, graph, funcs)
	if len(roots) != 1 {
		t.Fatalf("expected 1 chain flow root, got %d", len(roots))
	}

	// Check for bridge node
	hasBridge := false
	for _, child := range roots[0].Children {
		if child.IsBridge && child.Label == "B" {
			hasBridge = true
		}
	}
	if !hasBridge {
		t.Error("expected bridge node B in flow children")
	}
}

func TestBuildNonCodeRoot_NoNonGo(t *testing.T) {
	diffs := []*GitDiffFile{
		{Path: "main.go", Content: "diff"},
	}
	root := buildNonCodeRoot(".", diffs)
	if root != nil {
		t.Error("expected nil for all-Go diffs")
	}
}

func TestBuildNonCodeRoot_GroupsNonGo(t *testing.T) {
	diffs := []*GitDiffFile{
		{Path: "main.go", Content: "diff"},
		{Path: "config.yaml", Content: "diff yaml"},
		{Path: "README.md", Content: "diff md"},
	}
	root := buildNonCodeRoot(".", diffs)
	if root == nil {
		t.Fatal("expected non-code root")
	}
	if root.Label != "Non-code Files" {
		t.Errorf("expected label 'Non-code Files', got %s", root.Label)
	}
	if len(root.Children) != 2 {
		t.Errorf("expected 2 non-Go children, got %d", len(root.Children))
	}
}
