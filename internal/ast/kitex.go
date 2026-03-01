package ast

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// KitexCall represents a detected Kitex RPC client call.
type KitexCall struct {
	ServiceName string `json:"service_name" yaml:"service_name"`
	MethodName  string `json:"method_name" yaml:"method_name"`
	Package     string `json:"package" yaml:"package"`
	Receiver    string `json:"receiver" yaml:"receiver"`
}

// DetectKitexCalls scans a function body for Kitex client method calls.
// It checks if the receiver type originates from a kitex_gen package.
func DetectKitexCalls(pkg *packages.Package, fn *ast.FuncDecl) []*KitexCall {
	if fn.Body == nil {
		return nil
	}

	var calls []*KitexCall
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Check if the receiver type is from a kitex_gen package
		t := pkg.TypesInfo.TypeOf(sel.X)
		if t == nil {
			return true
		}

		if !isKitexGenType(t) {
			return true
		}

		calls = append(calls, &KitexCall{
			ServiceName: extractServiceName(t),
			MethodName:  sel.Sel.Name,
			Package:     pkg.PkgPath,
			Receiver:    exprToString(sel.X),
		})

		return true
	})

	return calls
}

// isKitexGenType checks if a type originates from a kitex_gen package.
func isKitexGenType(t types.Type) bool {
	typeStr := t.String()
	return strings.Contains(typeStr, "kitex_gen")
}

// extractServiceName extracts the service name from a kitex_gen type path.
// e.g., "*sample-app/kitex_gen/orderservice.Client" -> "orderservice"
func extractServiceName(t types.Type) string {
	typeStr := t.String()
	// Remove pointer prefix
	typeStr = strings.TrimPrefix(typeStr, "*")

	// Find kitex_gen/ and extract the next path segment
	idx := strings.Index(typeStr, "kitex_gen/")
	if idx < 0 {
		return "unknown"
	}
	remaining := typeStr[idx+len("kitex_gen/"):]
	// Take everything before the next dot or slash
	for i, c := range remaining {
		if c == '.' || c == '/' {
			return remaining[:i]
		}
	}
	return remaining
}

// MQCall represents a detected message queue producer call.
type MQCall struct {
	Topic    string `json:"topic,omitempty" yaml:"topic,omitempty"`
	Method   string `json:"method" yaml:"method"`
	Package  string `json:"package" yaml:"package"`
	Receiver string `json:"receiver" yaml:"receiver"`
}

// DetectMQCalls scans a function body for MQ producer method calls.
// Currently detects calls on types with "Producer" in the name.
func DetectMQCalls(pkg *packages.Package, fn *ast.FuncDecl) []*MQCall {
	if fn.Body == nil {
		return nil
	}

	var calls []*MQCall
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		t := pkg.TypesInfo.TypeOf(sel.X)
		if t == nil {
			return true
		}

		typeStr := t.String()
		if !strings.Contains(typeStr, "Producer") && !strings.Contains(typeStr, "producer") {
			return true
		}

		// Check if this is a "send" type method
		method := sel.Sel.Name
		if !isMQSendMethod(method) {
			return true
		}

		calls = append(calls, &MQCall{
			Method:   method,
			Package:  pkg.PkgPath,
			Receiver: exprToString(sel.X),
		})

		return true
	})

	return calls
}

func isMQSendMethod(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "send") || strings.Contains(lower, "publish") ||
		strings.Contains(lower, "produce") || strings.Contains(lower, "emit")
}
