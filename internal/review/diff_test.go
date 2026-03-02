package review

import (
	"strings"
	"testing"
)

func TestExtractFuncDiff_TrimToFuncRange(t *testing.T) {
	// Hunk spans lines 10-25, but function is only lines 15-20.
	// Should only keep lines within 15-20.
	diff := strings.Join([]string{
		"@@ -10,16 +10,16 @@",
		" context line 10",  // newLine=10
		" context line 11",  // newLine=11
		"+added line 12",    // newLine=12
		" context line 13",  // newLine=13
		" context line 14",  // newLine=14
		" context line 15",  // newLine=15 ← func start
		"+added line 16",    // newLine=16
		"-deleted line",     // no newLine (delete)
		"+replaced line 17", // newLine=17
		" context line 18",  // newLine=18
		" context line 19",  // newLine=19
		" context line 20",  // newLine=20 ← func end
		" context line 21",  // newLine=21
		"+added line 22",    // newLine=22
		" context line 23",  // newLine=23
	}, "\n")

	result := extractFuncDiff(diff, 15, 20)

	// Should contain the @@ header and only lines 15-20
	if result == "" {
		t.Fatal("expected non-empty result")
	}

	lines := strings.Split(result, "\n")
	// First line should be @@ header
	if !strings.HasPrefix(lines[0], "@@") {
		t.Errorf("expected @@ header, got: %s", lines[0])
	}

	// Should NOT contain "added line 12" (before func range)
	if strings.Contains(result, "added line 12") {
		t.Error("result should not contain diff lines before func range")
	}

	// Should NOT contain "added line 22" (after func range)
	if strings.Contains(result, "added line 22") {
		t.Error("result should not contain diff lines after func range")
	}

	// Should contain "added line 16" (within func range)
	if !strings.Contains(result, "added line 16") {
		t.Error("result should contain diff lines within func range")
	}

	// Should contain the delete line (between kept lines)
	if !strings.Contains(result, "deleted line") {
		t.Error("result should contain '-' lines between kept lines")
	}
}

func TestExtractFuncDiff_FuncInsideHunk(t *testing.T) {
	// Function lines 5-8, hunk covers exactly that range
	diff := strings.Join([]string{
		"@@ -5,4 +5,5 @@",
		" func foo() {",    // newLine=5
		"+\tnewLine1",      // newLine=6
		"+\tnewLine2",      // newLine=7
		" }",               // newLine=8
	}, "\n")

	result := extractFuncDiff(diff, 5, 8)

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	// All lines should be kept
	if !strings.Contains(result, "newLine1") {
		t.Error("should contain newLine1")
	}
	if !strings.Contains(result, "newLine2") {
		t.Error("should contain newLine2")
	}
}

func TestExtractFuncDiff_NoOverlap(t *testing.T) {
	// Hunk at lines 10-15, function at lines 50-60 — no overlap
	diff := strings.Join([]string{
		"@@ -10,6 +10,6 @@",
		" context",
		"+added",
		" context",
		" context",
		" context",
		" context",
	}, "\n")

	result := extractFuncDiff(diff, 50, 60)

	if result != "" {
		t.Errorf("expected empty result for non-overlapping range, got: %s", result)
	}
}

func TestExtractFuncDiff_DeleteOnlyHunk(t *testing.T) {
	// Hunk with only delete lines adjacent to in-range context
	diff := strings.Join([]string{
		"@@ -10,5 +10,3 @@",
		" func bar() {",    // newLine=10
		"-\toldLine1",      // delete
		"-\toldLine2",      // delete
		" }",               // newLine=11
	}, "\n")

	result := extractFuncDiff(diff, 10, 11)

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	// Should contain the delete lines (between two kept context lines)
	if !strings.Contains(result, "oldLine1") {
		t.Error("should contain deleted lines between kept lines")
	}
}

func TestExtractFuncDiff_MultipleHunks(t *testing.T) {
	// Two hunks: first overlaps, second doesn't
	diff := strings.Join([]string{
		"@@ -5,3 +5,4 @@",
		" line5",           // newLine=5
		"+added6",          // newLine=6
		" line7",           // newLine=7
		"@@ -50,3 +51,3 @@",
		" line51",          // newLine=51
		"+added52",         // newLine=52
		" line53",          // newLine=53
	}, "\n")

	result := extractFuncDiff(diff, 5, 7)

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	// Should contain first hunk content
	if !strings.Contains(result, "added6") {
		t.Error("should contain first hunk")
	}

	// Should NOT contain second hunk content
	if strings.Contains(result, "added52") {
		t.Error("should not contain second hunk outside func range")
	}
}
