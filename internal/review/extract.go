package review

import (
	"bufio"
	"fmt"
	"go/ast"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

// DiffEntry represents a changed function with its full source code.
type DiffEntry struct {
	Package   string `json:"package" yaml:"package"`
	Name      string `json:"name" yaml:"name"`
	Receiver  string `json:"receiver,omitempty" yaml:"receiver,omitempty"`
	File      string `json:"file" yaml:"file"`
	LineStart int    `json:"line_start" yaml:"line_start"`
	LineEnd   int    `json:"line_end" yaml:"line_end"`
	IsNew     bool   `json:"is_new" yaml:"is_new"`
	Code      string `json:"code" yaml:"code"`
}

// DiffResult holds all changed functions with their code.
type DiffResult struct {
	ChangedFunctions []*DiffEntry `json:"changed_functions" yaml:"changed_functions"`
}

func (dr *DiffResult) String() string {
	if len(dr.ChangedFunctions) == 0 {
		return "No function-level changes detected."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Changed Functions (%d):\n", len(dr.ChangedFunctions))
	fmt.Fprintln(&sb, strings.Repeat("=", 50))

	for _, entry := range dr.ChangedFunctions {
		tag := "modified"
		if entry.IsNew {
			tag = "new"
		}
		name := entry.Name
		if entry.Receiver != "" {
			name = entry.Receiver + "." + entry.Name
		}
		fmt.Fprintf(&sb, "\n[%s] %s.%s\n", tag, entry.Package, name)
		fmt.Fprintf(&sb, "  File: %s (lines %d-%d)\n", entry.File, entry.LineStart, entry.LineEnd)
		fmt.Fprintln(&sb, strings.Repeat("-", 40))
		fmt.Fprintln(&sb, entry.Code)
	}

	return sb.String()
}

// ExtractDiffEntries converts ChangedFuncs into DiffEntries with full source code.
func ExtractDiffEntries(changedFuncs []*ChangedFunc, pkgs map[string]*packages.Package) []*DiffEntry {
	var entries []*DiffEntry

	for _, cf := range changedFuncs {
		pkg, ok := pkgs[cf.Package]
		if !ok {
			continue
		}

		fn := findFuncDeclByName(pkg, cf.Name)
		if fn == nil {
			continue
		}

		fset := pkg.Fset
		startPos := fset.Position(fn.Pos())
		endPos := fset.Position(fn.End())

		code, err := readLines(startPos.Filename, startPos.Line, endPos.Line)
		if err != nil {
			continue
		}

		entries = append(entries, &DiffEntry{
			Package:   cf.Package,
			Name:      cf.Name,
			Receiver:  cf.Receiver,
			File:      startPos.Filename,
			LineStart: startPos.Line,
			LineEnd:   endPos.Line,
			IsNew:     cf.IsNew,
			Code:      code,
		})
	}

	return entries
}

// findFuncDeclByName finds a function/method declaration by name in a package.
func findFuncDeclByName(pkg *packages.Package, name string) *ast.FuncDecl {
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

// readLines reads lines [start, end] (1-based, inclusive) from a file.
func readLines(filename string, start, end int) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum >= start && lineNum <= end {
			lines = append(lines, scanner.Text())
		}
		if lineNum > end {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}
