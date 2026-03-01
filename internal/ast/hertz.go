package ast

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Route represents a discovered Hertz route registration.
type Route struct {
	Method  string `json:"method" yaml:"method"`
	Path    string `json:"path" yaml:"path"`
	Handler string `json:"handler" yaml:"handler"`
	Package string `json:"package" yaml:"package"`
	File    string `json:"file,omitempty" yaml:"file,omitempty"`
}

// RouteTable holds discovered routes.
type RouteTable struct {
	Routes []*Route `json:"routes" yaml:"routes"`
}

func (rt *RouteTable) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-8s %-30s %s\n", "METHOD", "PATH", "HANDLER")
	fmt.Fprintf(&sb, "%-8s %-30s %s\n", "------", "----", "-------")
	for _, r := range rt.Routes {
		fmt.Fprintf(&sb, "%-8s %-30s %s\n", r.Method, r.Path, r.Handler)
	}
	return sb.String()
}

var httpMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "DELETE": true,
	"PATCH": true, "HEAD": true, "OPTIONS": true, "Any": true,
}

// DiscoverRoutes scans the project for Hertz route registrations.
func DiscoverRoutes(project *Project) *RouteTable {
	rt := &RouteTable{}
	for _, pkg := range project.pkgMap {
		for _, file := range pkg.Syntax {
			routes := discoverRoutesInFile(pkg, file)
			rt.Routes = append(rt.Routes, routes...)
		}
	}
	return rt
}

func discoverRoutesInFile(pkg *packages.Package, file *ast.File) []*Route {
	// Step 1: Find Hertz engine variables (server.Default() / server.New())
	// Step 2: Track Group() calls to accumulate path prefixes
	// Step 3: Find HTTP method registrations

	// varPrefixes maps variable names to their accumulated path prefix
	varPrefixes := make(map[string]string)

	// First pass: find function parameters of Hertz server types
	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Type.Params == nil {
			return true
		}
		for _, field := range fn.Type.Params.List {
			t := pkg.TypesInfo.TypeOf(field.Type)
			if t != nil && strings.Contains(t.String(), "hertz") {
				for _, name := range field.Names {
					varPrefixes[name.Name] = ""
				}
			}
		}
		return true
	})

	// Second pass: find engine variables and group chains
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, rhs := range assign.Rhs {
			if i >= len(assign.Lhs) {
				continue
			}
			ident, ok := assign.Lhs[i].(*ast.Ident)
			if !ok {
				continue
			}

			call, ok := rhs.(*ast.CallExpr)
			if !ok {
				continue
			}

			// Check for server.Default() / server.New()
			if isHertzServerInit(pkg, call) {
				varPrefixes[ident.Name] = ""
				continue
			}

			// Check for .Group("/path")
			if prefix, receiver, ok := isGroupCall(pkg, call); ok {
				if parentPrefix, exists := varPrefixes[receiver]; exists {
					varPrefixes[ident.Name] = parentPrefix + prefix
				}
			}
		}
		return true
	})

	// Find route registrations
	var routes []*Route
	ast.Inspect(file, func(n ast.Node) bool {
		stmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}
		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		method := sel.Sel.Name
		if !httpMethods[method] {
			return true
		}

		// Get receiver variable name
		receiverName := exprToString(sel.X)
		prefix, exists := varPrefixes[receiverName]
		if !exists {
			// Could be a direct call on a known Hertz type
			if !isHertzType(pkg, sel.X) {
				return true
			}
		}

		// Extract path (first argument)
		if len(call.Args) < 2 {
			return true
		}
		path := stringLitValue(call.Args[0])
		if path == "" {
			return true
		}

		// Extract handler (second argument)
		handlerName := exprToString(call.Args[1])

		routes = append(routes, &Route{
			Method:  method,
			Path:    prefix + path,
			Handler: handlerName,
			Package: pkg.PkgPath,
		})

		return true
	})

	return routes
}

// isHertzServerInit checks if a call is server.Default() or server.New()
func isHertzServerInit(pkg *packages.Package, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Default" && sel.Sel.Name != "New" {
		return false
	}

	// Check if the receiver is the Hertz server package
	if ident, ok := sel.X.(*ast.Ident); ok {
		if obj, exists := pkg.TypesInfo.Uses[ident]; exists {
			if pkgName, ok := obj.(*types.PkgName); ok {
				return strings.HasSuffix(pkgName.Imported().Path(), "hertz/pkg/app/server")
			}
		}
	}
	return false
}

// isGroupCall checks if a call is .Group("/path") and returns the path prefix and receiver name.
func isGroupCall(pkg *packages.Package, call *ast.CallExpr) (prefix, receiver string, ok bool) {
	sel, okSel := call.Fun.(*ast.SelectorExpr)
	if !okSel || sel.Sel.Name != "Group" {
		return "", "", false
	}
	if len(call.Args) < 1 {
		return "", "", false
	}

	path := stringLitValue(call.Args[0])
	if path == "" {
		return "", "", false
	}

	receiverName := exprToString(sel.X)
	return path, receiverName, true
}

// isHertzType checks if an expression is of a Hertz router type.
func isHertzType(pkg *packages.Package, expr ast.Expr) bool {
	t := pkg.TypesInfo.TypeOf(expr)
	if t == nil {
		return false
	}
	return strings.Contains(t.String(), "hertz")
}

// stringLitValue extracts the value from a string literal expression.
func stringLitValue(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		return ""
	}
	// Remove quotes
	s := lit.Value
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return ""
}

// exprToString converts an expression to a readable string representation.
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.IndexExpr:
		return exprToString(e.X)
	default:
		return fmt.Sprintf("%T", expr)
	}
}
