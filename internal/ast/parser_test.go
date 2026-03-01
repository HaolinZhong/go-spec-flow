package ast

import (
	"testing"
)

func TestLoadProject(t *testing.T) {
	project, err := LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	if len(project.Packages) == 0 {
		t.Fatal("expected at least 1 package")
	}

	// Should find handler, service, dal, rpc, mq, router, kitex_gen packages
	pkgNames := make(map[string]bool)
	for _, pkg := range project.Packages {
		pkgNames[pkg.Name] = true
	}

	expected := []string{"handler", "service", "dal", "rpc", "mq", "router", "orderservice"}
	for _, name := range expected {
		if !pkgNames[name] {
			t.Errorf("expected package %q not found", name)
		}
	}
}

func TestDiscoverRoutes(t *testing.T) {
	project, err := LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	rt := DiscoverRoutes(project)
	if len(rt.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(rt.Routes))
	}

	routes := make(map[string]string)
	for _, r := range rt.Routes {
		routes[r.Method+" "+r.Path] = r.Handler
	}

	if _, ok := routes["POST /api/v1/orders"]; !ok {
		t.Error("POST /api/v1/orders not found")
	}
	if _, ok := routes["GET /api/v1/orders/:id"]; !ok {
		t.Error("GET /api/v1/orders/:id not found")
	}
}

func TestCallChainTracer(t *testing.T) {
	project, err := LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	tracer := NewTracer(project, TraceConfig{MaxDepth: 10})
	chain := tracer.Trace("sample-app/handler", "CreateOrder")

	if chain.Root == nil {
		t.Fatal("expected non-nil root")
	}

	if chain.Root.Name != "CreateOrder" {
		t.Errorf("root name = %q, want CreateOrder", chain.Root.Name)
	}

	// Should have children (calls to service.CreateOrder)
	if len(chain.Root.Children) == 0 {
		t.Error("expected children in call chain")
	}

	// Check that RPC and MQ nodes are present somewhere in the tree
	var hasRPC, hasMQ bool
	var walk func(node *CallNode)
	walk = func(node *CallNode) {
		if node.Type == NodeExternalRPC {
			hasRPC = true
		}
		if node.Type == NodeMQProducer {
			hasMQ = true
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(chain.Root)

	if !hasRPC {
		t.Error("expected RPC node in call chain")
	}
	if !hasMQ {
		t.Error("expected MQ node in call chain")
	}
}

func TestTraceFromRoute(t *testing.T) {
	project, err := LoadProject("../../testdata/sample-app")
	if err != nil {
		t.Fatalf("LoadProject error: %v", err)
	}

	rt := DiscoverRoutes(project)
	if len(rt.Routes) == 0 {
		t.Fatal("no routes found")
	}

	tracer := NewTracer(project, TraceConfig{MaxDepth: 10})
	for _, route := range rt.Routes {
		chain := tracer.TraceFromRoute(route)
		if chain.Root == nil {
			t.Errorf("nil root for route %s %s", route.Method, route.Path)
		}
	}
}
