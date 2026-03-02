package ast

import (
	"testing"
)

func TestFindCallers(t *testing.T) {
	project, err := LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	// CreateOrder in service is called by handler.CreateOrder
	result := FindCallers(project, "sample-app/service", "CreateOrder")

	if result.Target.Package != "sample-app/service" {
		t.Errorf("target package = %q, want sample-app/service", result.Target.Package)
	}
	if result.Target.Name != "CreateOrder" {
		t.Errorf("target name = %q, want CreateOrder", result.Target.Name)
	}

	if len(result.Callers) == 0 {
		t.Fatal("expected at least 1 caller for service.CreateOrder")
	}

	// Check that handler package is among callers
	foundHandler := false
	for _, c := range result.Callers {
		if c.Package == "sample-app/handler" && c.Name == "CreateOrder" {
			foundHandler = true
			break
		}
	}
	if !foundHandler {
		t.Error("expected handler.CreateOrder as a caller of service.CreateOrder")
	}
}

func TestFindCallersNoCallers(t *testing.T) {
	project, err := LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	// Look for callers of a function that nobody calls
	result := FindCallers(project, "sample-app/handler", "NonExistentFunc")
	if len(result.Callers) != 0 {
		t.Errorf("expected 0 callers, got %d", len(result.Callers))
	}
}

func TestFindCallersCrossPackage(t *testing.T) {
	project, err := LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	// dal.Create is called by service.CreateOrder
	result := FindCallers(project, "sample-app/dal", "Create")

	if len(result.Callers) == 0 {
		t.Fatal("expected at least 1 caller for dal.Create")
	}

	foundService := false
	for _, c := range result.Callers {
		if c.Package == "sample-app/service" {
			foundService = true
			break
		}
	}
	if !foundService {
		t.Error("expected service package as a caller of dal.Create")
	}
}

func TestFindCallersInFuncLit(t *testing.T) {
	project, err := LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	// service.GetOrder is called from cli/cmd.go inside a package-level FuncLit (orderCmd.RunE)
	result := FindCallers(project, "sample-app/service", "GetOrder")

	// Should find callers from both handler (FuncDecl) and cli (FuncLit)
	foundHandler := false
	foundCLI := false
	for _, c := range result.Callers {
		if c.Package == "sample-app/handler" && c.Name == "GetOrder" {
			foundHandler = true
		}
		if c.Package == "sample-app/cli" && c.Name == "orderCmd" {
			foundCLI = true
		}
	}

	if !foundHandler {
		t.Error("expected handler.GetOrder as a caller (FuncDecl)")
	}
	if !foundCLI {
		t.Error("expected cli.orderCmd as a caller (FuncLit in package-level var)")
	}
}

func TestCallersResultString(t *testing.T) {
	result := &CallersResult{
		Target: CallerTarget{Package: "pkg/a", Name: "Foo"},
		Callers: []*CallerInfo{
			{Package: "pkg/b", Name: "Bar", File: "b/bar.go", Line: 10},
		},
	}

	s := result.String()
	if s == "" {
		t.Error("CallersResult.String() returned empty")
	}
}
