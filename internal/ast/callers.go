package ast

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// CallerInfo represents a direct caller of a function.
type CallerInfo struct {
	Package  string `json:"package" yaml:"package"`
	Name     string `json:"name" yaml:"name"`
	File     string `json:"file" yaml:"file"`
	Line     int    `json:"line" yaml:"line"`
}

// CallersResult holds the target function and its direct callers.
type CallersResult struct {
	Target  CallerTarget  `json:"target" yaml:"target"`
	Callers []*CallerInfo `json:"callers" yaml:"callers"`
}

// CallerTarget identifies the function being looked up.
type CallerTarget struct {
	Package string `json:"package" yaml:"package"`
	Name    string `json:"name" yaml:"name"`
}

func (cr *CallersResult) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Callers of %s.%s\n", cr.Target.Package, cr.Target.Name)
	fmt.Fprintln(&sb, strings.Repeat("=", 50))

	if len(cr.Callers) == 0 {
		fmt.Fprintln(&sb, "  (no callers found)")
		return sb.String()
	}

	for _, c := range cr.Callers {
		fmt.Fprintf(&sb, "  %s.%s\n", c.Package, c.Name)
		fmt.Fprintf(&sb, "    %s:%d\n", c.File, c.Line)
	}
	return sb.String()
}

// FindCallers finds all direct callers (one level) of the specified function in the project.
// It searches both function declarations and package-level variable initializers (e.g., cobra RunE).
func FindCallers(project *Project, targetPkg, targetFunc string) *CallersResult {
	result := &CallersResult{
		Target: CallerTarget{
			Package: targetPkg,
			Name:    targetFunc,
		},
	}

	for pkgPath, pkg := range project.pkgMap {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Body == nil {
						continue
					}
					scanBodyForCalls(pkg, d.Body, pkgPath, d.Name.Name, targetPkg, targetFunc, result)

				case *ast.GenDecl:
					// Walk variable declarations for FuncLit (e.g., cobra RunE)
					for _, spec := range d.Specs {
						vs, ok := spec.(*ast.ValueSpec)
						if !ok {
							continue
						}
						varName := ""
						if len(vs.Names) > 0 {
							varName = vs.Names[0].Name
						}
						for _, val := range vs.Values {
							ast.Inspect(val, func(n ast.Node) bool {
								fl, ok := n.(*ast.FuncLit)
								if !ok {
									return true
								}
								scanBodyForCalls(pkg, fl.Body, pkgPath, varName, targetPkg, targetFunc, result)
								return true
							})
						}
					}
				}
			}
		}
	}

	return result
}

// scanBodyForCalls inspects a function body for calls to the target function.
func scanBodyForCalls(pkg *packages.Package, body *ast.BlockStmt, callerPkg, callerName, targetPkg, targetFunc string, result *CallersResult) {
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		callPkg, callFunc := ResolveCallTarget(pkg, call)
		if callPkg == targetPkg && callFunc == targetFunc {
			pos := pkg.Fset.Position(call.Pos())
			result.Callers = append(result.Callers, &CallerInfo{
				Package: callerPkg,
				Name:    callerName,
				File:    pos.Filename,
				Line:    pos.Line,
			})
		}

		return true
	})
}

// ResolveCallTarget resolves a call expression to its target package and function name.
// Shared by both FindCallers and Tracer to ensure consistent resolution.
func ResolveCallTarget(pkg *packages.Package, call *ast.CallExpr) (pkgPath, funcName string) {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		if obj, ok := pkg.TypesInfo.Uses[fn]; ok {
			if f, ok := obj.(*types.Func); ok {
				if f.Pkg() != nil {
					return f.Pkg().Path(), f.Name()
				}
			}
		}

	case *ast.SelectorExpr:
		sel := pkg.TypesInfo.Selections[fn]
		if sel != nil {
			obj := sel.Obj()
			if f, ok := obj.(*types.Func); ok {
				sig := f.Type().(*types.Signature)
				recv := sig.Recv()
				if recv != nil {
					recvType := recv.Type()
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
