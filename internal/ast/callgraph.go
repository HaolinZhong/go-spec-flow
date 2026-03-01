package ast

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// NodeType represents the type of a call chain node.
type NodeType string

const (
	NodeFunction    NodeType = "function"
	NodeExternalRPC NodeType = "external-rpc"
	NodeMQProducer  NodeType = "mq-producer"
)

// CallNode represents a node in the call chain tree.
type CallNode struct {
	Name     string      `json:"name" yaml:"name"`
	Package  string      `json:"package" yaml:"package"`
	Type     NodeType    `json:"type" yaml:"type"`
	Children []*CallNode `json:"children,omitempty" yaml:"children,omitempty"`

	// Extra info for special node types
	ServiceName string `json:"service_name,omitempty" yaml:"service_name,omitempty"`
	MethodName  string `json:"method_name,omitempty" yaml:"method_name,omitempty"`
}

// CallChain holds the traced call tree from an entry point.
type CallChain struct {
	Root *CallNode `json:"root" yaml:"root"`
}

func (cc *CallChain) String() string {
	if cc.Root == nil {
		return "(empty call chain)"
	}
	var sb strings.Builder
	printTree(&sb, cc.Root, "", true)
	return sb.String()
}

func nodeLabel(node *CallNode) string {
	switch node.Type {
	case NodeExternalRPC:
		return fmt.Sprintf("[RPC] %s.%s", node.ServiceName, node.MethodName)
	case NodeMQProducer:
		return fmt.Sprintf("[MQ] %s.%s", node.Package, node.Name)
	default:
		return fmt.Sprintf("%s.%s", node.Package, node.Name)
	}
}

func printTree(sb *strings.Builder, node *CallNode, indent string, isRoot bool) {
	fmt.Fprintln(sb, nodeLabel(node))
	for i, child := range node.Children {
		isLast := i == len(node.Children)-1
		if isLast {
			fmt.Fprintf(sb, "%s└── ", indent)
			printTree(sb, child, indent+"    ", false)
		} else {
			fmt.Fprintf(sb, "%s├── ", indent)
			printTree(sb, child, indent+"│   ", false)
		}
	}
}

// TraceConfig controls the trace behavior.
type TraceConfig struct {
	MaxDepth int
}

// Tracer builds call chains from a loaded project.
type Tracer struct {
	project *Project
	config  TraceConfig
	visited map[string]bool // cycle detection: "pkg.FuncName"
}

// NewTracer creates a new call chain tracer.
func NewTracer(project *Project, config TraceConfig) *Tracer {
	if config.MaxDepth <= 0 {
		config.MaxDepth = 10
	}
	return &Tracer{
		project: project,
		config:  config,
	}
}

// Trace builds a call chain starting from the given function.
func (t *Tracer) Trace(pkgPath, funcName string) *CallChain {
	t.visited = make(map[string]bool)
	root := t.traceFunc(pkgPath, funcName, 0)
	return &CallChain{Root: root}
}

// TraceFromRoute traces from a route handler, resolving the handler reference.
func (t *Tracer) TraceFromRoute(route *Route) *CallChain {
	// Handler is like "orderHandler.CreateOrder" - need to resolve to actual function
	parts := strings.Split(route.Handler, ".")
	if len(parts) != 2 {
		return &CallChain{}
	}

	methodName := parts[1]
	// Find which package/type has this method
	pkgPath, resolvedName := t.resolveHandler(route.Package, parts[0], methodName)
	if pkgPath == "" {
		return &CallChain{}
	}

	t.visited = make(map[string]bool)
	root := t.traceFunc(pkgPath, resolvedName, 0)
	return &CallChain{Root: root}
}

func (t *Tracer) traceFunc(pkgPath, funcName string, depth int) *CallNode {
	key := pkgPath + "." + funcName
	if t.visited[key] || depth > t.config.MaxDepth {
		return &CallNode{
			Name:    funcName + " (cycle/depth-limit)",
			Package: pkgPath,
			Type:    NodeFunction,
		}
	}
	t.visited[key] = true
	defer func() { t.visited[key] = false }()

	node := &CallNode{
		Name:    funcName,
		Package: pkgPath,
		Type:    NodeFunction,
	}

	pkg, ok := t.project.pkgMap[pkgPath]
	if !ok {
		return node
	}

	// Find the function declaration
	fn := findFuncDecl(pkg, funcName)
	if fn == nil || fn.Body == nil {
		return node
	}

	// Walk function body and classify each call
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check for Kitex RPC calls first
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			t2 := pkg.TypesInfo.TypeOf(sel.X)
			if t2 != nil && isKitexGenType(t2) {
				node.Children = append(node.Children, &CallNode{
					Name:        sel.Sel.Name,
					Package:     pkgPath,
					Type:        NodeExternalRPC,
					ServiceName: extractServiceName(t2),
					MethodName:  sel.Sel.Name,
				})
				return true
			}

			// Check for MQ producer calls
			if t2 != nil {
				typeStr := t2.String()
				if (strings.Contains(typeStr, "Producer") || strings.Contains(typeStr, "producer")) && isMQSendMethod(sel.Sel.Name) {
					node.Children = append(node.Children, &CallNode{
						Name:    sel.Sel.Name,
						Package: pkgPath,
						Type:    NodeMQProducer,
					})
					return true
				}
			}
		}

		targetPkg, targetFunc := t.resolveCallTarget(pkg, call)
		if targetPkg == "" || targetFunc == "" {
			return true
		}

		// Skip kitex_gen calls
		if strings.Contains(targetPkg, "kitex_gen") {
			return true
		}
		if _, hasPkg := t.project.pkgMap[targetPkg]; !hasPkg {
			return true // external package, skip
		}

		child := t.traceFunc(targetPkg, targetFunc, depth+1)
		node.Children = append(node.Children, child)

		return true
	})

	return node
}

// resolveCallTarget resolves a CallExpr to its target package and function name.
func (t *Tracer) resolveCallTarget(pkg *packages.Package, call *ast.CallExpr) (pkgPath, funcName string) {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		// Direct function call in same package
		if obj, ok := pkg.TypesInfo.Uses[fn]; ok {
			if f, ok := obj.(*types.Func); ok {
				if f.Pkg() != nil {
					return f.Pkg().Path(), f.Name()
				}
			}
		}

	case *ast.SelectorExpr:
		// pkg.Func() or receiver.Method()
		sel := pkg.TypesInfo.Selections[fn]
		if sel != nil {
			// Method call on a receiver
			obj := sel.Obj()
			if f, ok := obj.(*types.Func); ok {
				sig := f.Type().(*types.Signature)
				recv := sig.Recv()
				if recv != nil {
					recvType := recv.Type()
					// Unwrap pointer
					if ptr, ok := recvType.(*types.Pointer); ok {
						recvType = ptr.Elem()
					}
					if named, ok := recvType.(*types.Named); ok {
						if named.Obj().Pkg() != nil {
							return named.Obj().Pkg().Path(), f.Name()
						}
					}
				}
				if f.Pkg() != nil {
					return f.Pkg().Path(), f.Name()
				}
			}
		}

		// Qualified identifier (package.Function)
		if ident, ok := fn.X.(*ast.Ident); ok {
			if obj, exists := pkg.TypesInfo.Uses[ident]; exists {
				if pkgName, ok := obj.(*types.PkgName); ok {
					return pkgName.Imported().Path(), fn.Sel.Name
				}
			}
		}
	}

	return "", ""
}

// resolveHandler resolves a handler reference like "orderHandler.CreateOrder"
// to the actual package and method name.
func (t *Tracer) resolveHandler(routePkgPath, varName, methodName string) (string, string) {
	pkg, ok := t.project.pkgMap[routePkgPath]
	if !ok {
		return "", ""
	}

	// Find the variable declaration to get its type
	for _, file := range pkg.Syntax {
		var foundType types.Type
		ast.Inspect(file, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i, lhs := range assign.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || ident.Name != varName {
					continue
				}
				if i < len(assign.Rhs) {
					foundType = pkg.TypesInfo.TypeOf(assign.Rhs[i])
				}
			}
			return true
		})

		if foundType != nil {
			// Unwrap pointer
			if ptr, ok := foundType.(*types.Pointer); ok {
				foundType = ptr.Elem()
			}
			if named, ok := foundType.(*types.Named); ok {
				return named.Obj().Pkg().Path(), methodName
			}
		}
	}

	return "", ""
}

// findFuncDecl finds a function or method declaration by name.
func findFuncDecl(pkg *packages.Package, name string) *ast.FuncDecl {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Name == name {
				return fn
			}
		}
	}
	return nil
}
