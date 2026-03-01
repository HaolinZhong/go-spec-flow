package review

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ChangedFunc represents a function that was modified in the diff.
type ChangedFunc struct {
	Package  string
	Name     string
	Receiver string // empty for standalone functions
	IsNew    bool
}

func (cf *ChangedFunc) FullName() string {
	if cf.Receiver != "" {
		return cf.Receiver + "." + cf.Name
	}
	return cf.Name
}

// MapDiffToFunctions maps diff hunks to function/method declarations.
func MapDiffToFunctions(diffs []*FileDiff, pkgs map[string]*packages.Package) []*ChangedFunc {
	var changed []*ChangedFunc
	seen := make(map[string]bool)

	for _, d := range diffs {
		if !strings.HasSuffix(d.Path, ".go") {
			continue
		}

		// Find the package that contains this file
		for _, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				fset := pkg.Fset
				pos := fset.Position(file.Pos())
				if !strings.HasSuffix(pos.Filename, d.Path) && !strings.HasSuffix(d.Path, pos.Filename) {
					// Try matching just the file name within the package
					continue
				}

				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok {
						continue
					}

					if d.IsNew || funcOverlapsHunks(fset, fn, d.Hunks) {
						cf := &ChangedFunc{
							Package: pkg.PkgPath,
							Name:    fn.Name.Name,
							IsNew:   d.IsNew,
						}
						if fn.Recv != nil && len(fn.Recv.List) > 0 {
							cf.Receiver = receiverTypeName(fn.Recv.List[0].Type)
						}

						key := cf.Package + "." + cf.FullName()
						if !seen[key] {
							seen[key] = true
							changed = append(changed, cf)
						}
					}
				}
			}
		}
	}

	return changed
}

// funcOverlapsHunks checks if any diff hunk overlaps with the function's line range.
func funcOverlapsHunks(fset *token.FileSet, fn *ast.FuncDecl, hunks []*Hunk) bool {
	start := fset.Position(fn.Pos()).Line
	end := fset.Position(fn.End()).Line

	for _, hunk := range hunks {
		hunkEnd := hunk.NewStart + hunk.NewCount - 1
		if hunkEnd < hunk.NewStart {
			hunkEnd = hunk.NewStart
		}
		// Check overlap
		if hunk.NewStart <= end && hunkEnd >= start {
			return true
		}
	}
	return false
}

func receiverTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(e.X)
	case *ast.Ident:
		return e.Name
	default:
		return ""
	}
}
