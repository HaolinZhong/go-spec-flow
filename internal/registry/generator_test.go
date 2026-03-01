package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateFromIDL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gsf-registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	services, err := GenerateFromIDL("../../testdata/idl", tmpDir)
	if err != nil {
		t.Fatalf("GenerateFromIDL error: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	svc := services[0]
	if svc.Service != "OrderService" {
		t.Errorf("service name = %q, want OrderService", svc.Service)
	}

	if len(svc.Methods) != 3 {
		t.Errorf("expected 3 methods, got %d", len(svc.Methods))
	}

	// Verify auto.yaml was written
	autoPath := filepath.Join(tmpDir, "OrderService", "auto.yaml")
	if _, err := os.Stat(autoPath); os.IsNotExist(err) {
		t.Error("auto.yaml not created")
	}

	// Verify registry-index.yaml was written
	indexPath := filepath.Join(tmpDir, "registry-index.yaml")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("registry-index.yaml not created")
	}

	// Verify we can load it back
	info, err := LoadServiceInfo(tmpDir, "OrderService")
	if err != nil {
		t.Fatalf("LoadServiceInfo error: %v", err)
	}
	if info.Service != "OrderService" {
		t.Errorf("loaded service = %q, want OrderService", info.Service)
	}
}

func TestMergeServiceInfo(t *testing.T) {
	info := &ServiceInfo{
		Service: "TestService",
		IDLPath: "test.thrift",
		Methods: []*MethodInfo{
			{
				Name: "DoWork",
				Request: []*FieldInfo{
					{Name: "id", Type: "i64"},
				},
			},
		},
	}

	idempotent := true
	ctx := &ServiceContext{
		Service: "TestService",
		Notes:   "Test service for unit testing",
		Methods: map[string]*MethodContext{
			"DoWork": {
				Idempotent: &idempotent,
				TimeoutMs:  2000,
				Notes:      "Does work",
			},
		},
	}

	merged := MergeServiceInfo(info, ctx)
	if merged.Notes != "Test service for unit testing" {
		t.Errorf("notes = %q, want 'Test service for unit testing'", merged.Notes)
	}

	if len(merged.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(merged.Methods))
	}

	m := merged.Methods[0]
	if m.Idempotent == nil || !*m.Idempotent {
		t.Error("expected idempotent = true")
	}
	if m.TimeoutMs != 2000 {
		t.Errorf("timeout = %d, want 2000", m.TimeoutMs)
	}
}
