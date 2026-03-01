package thrift

import (
	"testing"
)

func TestLexer(t *testing.T) {
	input := `
namespace go order

struct CreateOrderRequest {
    1: required i64 user_id
    2: optional string address
}

enum Status {
    CREATED = 0
    DONE = 1
}

service OrderService {
    CreateOrderRequest CreateOrder(1: CreateOrderRequest req) throws (1: OrderException e)
}
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	// Should have tokens for namespace, struct, enum, service
	var kwCount int
	for _, tok := range tokens {
		if tok.Type == TokenNamespace || tok.Type == TokenStruct ||
			tok.Type == TokenEnum || tok.Type == TokenService {
			kwCount++
		}
	}
	if kwCount != 4 {
		t.Errorf("expected 4 keyword tokens, got %d", kwCount)
	}
}

func TestParser(t *testing.T) {
	input := `
namespace go order

include "common.thrift"

struct CreateOrderRequest {
    1: required i64 user_id
    2: required i64 product_id
    3: optional string address
}

struct CreateOrderResponse {
    1: string order_id
}

enum OrderStatus {
    CREATED = 0
    PAID = 1
}

exception OrderException {
    1: i32 code
    2: string message
}

service OrderService {
    CreateOrderResponse CreateOrder(1: CreateOrderRequest req) throws (1: OrderException e)
}
`
	p := &Parser{parsed: make(map[string]*Document)}
	doc, err := p.parse("test.thrift", input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if doc.Namespace != "order" {
		t.Errorf("namespace = %q, want %q", doc.Namespace, "order")
	}

	if len(doc.Includes) != 1 || doc.Includes[0] != "common.thrift" {
		t.Errorf("includes = %v, want [common.thrift]", doc.Includes)
	}

	if len(doc.Structs) != 2 {
		t.Errorf("structs count = %d, want 2", len(doc.Structs))
	}

	if len(doc.Enums) != 1 {
		t.Errorf("enums count = %d, want 1", len(doc.Enums))
	}

	if len(doc.Exceptions) != 1 {
		t.Errorf("exceptions count = %d, want 1", len(doc.Exceptions))
	}

	if len(doc.Services) != 1 {
		t.Fatalf("services count = %d, want 1", len(doc.Services))
	}

	svc := doc.Services[0]
	if svc.Name != "OrderService" {
		t.Errorf("service name = %q, want %q", svc.Name, "OrderService")
	}

	if len(svc.Methods) != 1 {
		t.Fatalf("methods count = %d, want 1", len(svc.Methods))
	}

	method := svc.Methods[0]
	if method.Name != "CreateOrder" {
		t.Errorf("method name = %q, want %q", method.Name, "CreateOrder")
	}

	if len(method.Params) != 1 {
		t.Errorf("params count = %d, want 1", len(method.Params))
	}

	if len(method.Throws) != 1 {
		t.Errorf("throws count = %d, want 1", len(method.Throws))
	}
}

func TestParserCollectionTypes(t *testing.T) {
	input := `
struct Response {
    1: list<string> items
    2: map<string, i64> counts
    3: set<i32> ids
}
`
	p := &Parser{parsed: make(map[string]*Document)}
	doc, err := p.parse("test.thrift", input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(doc.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(doc.Structs))
	}

	fields := doc.Structs[0].Fields
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}

	if fields[0].Type.String() != "list<string>" {
		t.Errorf("field 0 type = %q, want list<string>", fields[0].Type.String())
	}
	if fields[1].Type.String() != "map<string,i64>" {
		t.Errorf("field 1 type = %q, want map<string,i64>", fields[1].Type.String())
	}
	if fields[2].Type.String() != "set<i32>" {
		t.Errorf("field 2 type = %q, want set<i32>", fields[2].Type.String())
	}
}

func TestParseDir(t *testing.T) {
	docs, err := ParseDir("../../testdata/idl")
	if err != nil {
		t.Fatalf("ParseDir error: %v", err)
	}

	if len(docs) < 2 {
		t.Errorf("expected at least 2 documents, got %d", len(docs))
	}

	var foundOrder bool
	for _, doc := range docs {
		if doc.Filename == "order.thrift" {
			foundOrder = true
			if len(doc.Services) != 1 {
				t.Errorf("order.thrift: expected 1 service, got %d", len(doc.Services))
			}
		}
	}
	if !foundOrder {
		t.Error("order.thrift not found in parsed documents")
	}
}
