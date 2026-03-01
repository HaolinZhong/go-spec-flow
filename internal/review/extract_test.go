package review

import (
	"testing"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
)

func TestExtractDiffEntries(t *testing.T) {
	project, err := goast.LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	changedFuncs := []*ChangedFunc{
		{
			Package: "sample-app/handler",
			Name:    "CreateOrder",
			IsNew:   false,
		},
		{
			Package:  "sample-app/service",
			Name:     "CreateOrder",
			Receiver: "OrderService",
			IsNew:    false,
		},
	}

	entries := ExtractDiffEntries(changedFuncs, project.RawPackages())

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Check first entry (handler.CreateOrder)
	e := entries[0]
	if e.Package != "sample-app/handler" {
		t.Errorf("entry[0].Package = %q, want sample-app/handler", e.Package)
	}
	if e.Name != "CreateOrder" {
		t.Errorf("entry[0].Name = %q, want CreateOrder", e.Name)
	}
	if e.Code == "" {
		t.Error("entry[0].Code is empty")
	}
	if e.LineStart == 0 || e.LineEnd == 0 {
		t.Error("entry[0] line range is zero")
	}
	if e.LineEnd < e.LineStart {
		t.Errorf("entry[0] line range invalid: %d-%d", e.LineStart, e.LineEnd)
	}

	// Check second entry (service.CreateOrder)
	e2 := entries[1]
	if e2.Package != "sample-app/service" {
		t.Errorf("entry[1].Package = %q, want sample-app/service", e2.Package)
	}
	if e2.Code == "" {
		t.Error("entry[1].Code is empty")
	}
}

func TestDiffResultString(t *testing.T) {
	dr := &DiffResult{
		ChangedFunctions: []*DiffEntry{
			{
				Package:   "sample-app/handler",
				Name:      "CreateOrder",
				File:      "handler/order.go",
				LineStart: 21,
				LineEnd:   28,
				IsNew:     false,
				Code:      "func (h *OrderHandler) CreateOrder(...) {\n}",
			},
			{
				Package:   "sample-app/service",
				Name:      "NewFunc",
				File:      "service/order.go",
				LineStart: 55,
				LineEnd:   60,
				IsNew:     true,
				Code:      "func NewFunc() {\n}",
			},
		},
	}

	s := dr.String()
	if s == "" {
		t.Error("DiffResult.String() returned empty")
	}

	// Should contain both tags
	if !containsStr(s, "[modified]") {
		t.Error("missing [modified] tag")
	}
	if !containsStr(s, "[new]") {
		t.Error("missing [new] tag")
	}
}

func TestDiffResultStringEmpty(t *testing.T) {
	dr := &DiffResult{}
	s := dr.String()
	if s != "No function-level changes detected." {
		t.Errorf("unexpected empty result string: %q", s)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstr(s, sub))
}

func findSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
