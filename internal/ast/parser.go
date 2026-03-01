package ast

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Project holds the loaded packages and extracted structure.
type Project struct {
	Dir      string              `json:"dir" yaml:"dir"`
	Packages []*PackageInfo      `json:"packages" yaml:"packages"`
	pkgMap   map[string]*packages.Package // internal: keyed by package path
}

// PackageInfo holds extracted structure for a single package.
type PackageInfo struct {
	Path       string          `json:"path" yaml:"path"`
	Name       string          `json:"name" yaml:"name"`
	Structs    []*StructInfo   `json:"structs,omitempty" yaml:"structs,omitempty"`
	Interfaces []*InterfaceInfo `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
	Functions  []*FunctionInfo `json:"functions,omitempty" yaml:"functions,omitempty"`
}

type StructInfo struct {
	Name    string        `json:"name" yaml:"name"`
	Methods []*MethodInfo `json:"methods,omitempty" yaml:"methods,omitempty"`
	Fields  []*FieldInfo  `json:"fields,omitempty" yaml:"fields,omitempty"`
}

type FieldInfo struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

type InterfaceInfo struct {
	Name    string        `json:"name" yaml:"name"`
	Methods []*MethodInfo `json:"methods,omitempty" yaml:"methods,omitempty"`
}

type MethodInfo struct {
	Name      string `json:"name" yaml:"name"`
	Signature string `json:"signature" yaml:"signature"`
}

type FunctionInfo struct {
	Name      string `json:"name" yaml:"name"`
	Signature string `json:"signature" yaml:"signature"`
	Receiver  string `json:"receiver,omitempty" yaml:"receiver,omitempty"`
}

// LoadProject loads all Go packages in the given directory with type information.
func LoadProject(dir string) (*Project, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps |
			packages.NeedImports,
		Dir: dir,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	// Check for load errors
	var errs []string
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			errs = append(errs, e.Error())
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("package errors:\n%s", strings.Join(errs, "\n"))
	}

	project := &Project{
		Dir:    dir,
		pkgMap: make(map[string]*packages.Package),
	}

	for _, pkg := range pkgs {
		project.pkgMap[pkg.PkgPath] = pkg
		info := extractPackageInfo(pkg)
		project.Packages = append(project.Packages, info)
	}

	return project, nil
}

// RawPackages returns the underlying loaded packages for advanced analysis.
func (p *Project) RawPackages() map[string]*packages.Package {
	return p.pkgMap
}

func extractPackageInfo(pkg *packages.Package) *PackageInfo {
	info := &PackageInfo{
		Path: pkg.PkgPath,
		Name: pkg.Name,
	}

	if pkg.Types == nil {
		return info
	}

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if !obj.Exported() {
			continue
		}

		switch o := obj.(type) {
		case *types.TypeName:
			named, ok := o.Type().(*types.Named)
			if !ok {
				continue
			}
			underlying := named.Underlying()
			switch underlying.(type) {
			case *types.Struct:
				info.Structs = append(info.Structs, extractStructInfo(named))
			case *types.Interface:
				info.Interfaces = append(info.Interfaces, extractInterfaceInfo(named))
			}
		case *types.Func:
			info.Functions = append(info.Functions, &FunctionInfo{
				Name:      o.Name(),
				Signature: o.Type().(*types.Signature).String(),
			})
		}
	}

	// Extract methods (declared on types in this package)
	for _, s := range info.Structs {
		extractMethods(pkg, s)
	}

	return info
}

func extractStructInfo(named *types.Named) *StructInfo {
	s := named.Underlying().(*types.Struct)
	si := &StructInfo{Name: named.Obj().Name()}
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		if f.Exported() {
			si.Fields = append(si.Fields, &FieldInfo{
				Name: f.Name(),
				Type: types.TypeString(f.Type(), nil),
			})
		}
	}
	return si
}

func extractInterfaceInfo(named *types.Named) *InterfaceInfo {
	iface := named.Underlying().(*types.Interface)
	ii := &InterfaceInfo{Name: named.Obj().Name()}
	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		ii.Methods = append(ii.Methods, &MethodInfo{
			Name:      m.Name(),
			Signature: m.Type().(*types.Signature).String(),
		})
	}
	return ii
}

func extractMethods(pkg *packages.Package, si *StructInfo) {
	scope := pkg.Types.Scope()
	obj := scope.Lookup(si.Name)
	if obj == nil {
		return
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return
	}
	for i := 0; i < named.NumMethods(); i++ {
		m := named.Method(i)
		if m.Exported() {
			si.Methods = append(si.Methods, &MethodInfo{
				Name:      m.Name(),
				Signature: m.Type().(*types.Signature).String(),
			})
		}
	}
}

// FindFuncDecl finds a function/method declaration by package path and function name.
func (p *Project) FindFuncDecl(pkgPath, funcName string) *ast.FuncDecl {
	pkg, ok := p.pkgMap[pkgPath]
	if !ok {
		return nil
	}
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Name == funcName {
				return fn
			}
		}
	}
	return nil
}

// String returns a text summary of the project structure.
func (p *Project) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Project: %s\n", p.Dir)
	fmt.Fprintf(&sb, "Packages: %d\n\n", len(p.Packages))
	for _, pkg := range p.Packages {
		fmt.Fprintf(&sb, "  %s (%s)\n", pkg.Path, pkg.Name)
		for _, s := range pkg.Structs {
			fmt.Fprintf(&sb, "    struct %s\n", s.Name)
			for _, m := range s.Methods {
				fmt.Fprintf(&sb, "      func %s %s\n", m.Name, m.Signature)
			}
		}
		for _, iface := range pkg.Interfaces {
			fmt.Fprintf(&sb, "    interface %s\n", iface.Name)
			for _, m := range iface.Methods {
				fmt.Fprintf(&sb, "      %s %s\n", m.Name, m.Signature)
			}
		}
		for _, fn := range pkg.Functions {
			fmt.Fprintf(&sb, "    func %s %s\n", fn.Name, fn.Signature)
		}
	}
	return sb.String()
}
