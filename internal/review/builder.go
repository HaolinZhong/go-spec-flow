package review

import (
	"bufio"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
)

// BuildDiffTree builds a flow tree from a git diff range.
// The tree is organized by changed files, with each file's diff as the code.
// When a project is loadable, trace/callers are used to add structural context.
func BuildDiffTree(dir, diffRange string, maxDepth int) (*FlowTree, error) {
	if maxDepth <= 0 {
		maxDepth = 2
	}
	// Get per-file diffs
	diffs, err := RunGitDiff(dir, diffRange)
	if err != nil {
		return nil, err
	}

	if len(diffs) == 0 {
		return &FlowTree{Mode: "diff", Title: "No changes found"}, nil
	}

	tree := &FlowTree{
		Mode:  "diff",
		Title: fmt.Sprintf("Diff Review: %s", diffRange),
	}

	// Try loading project for structural context (non-fatal if fails)
	project, _ := goast.LoadProject(dir)

	for i, df := range diffs {
		node := &FlowNode{
			ID:       fmt.Sprintf("file-%d", i),
			Label:    df.Path,
			File:     df.Path,
			Diff:     df.Content,
			NodeType: "file",
			IsNew:    df.IsNew,
		}

		// Try to read current file source for source view
		if source, err := readEntireFile(filepath.Join(dir, df.Path)); err == nil {
			node.Code = source
		}

		// If project loaded and it's a Go file, add function-level children via trace
		if project != nil && strings.HasSuffix(df.Path, ".go") {
			children := buildFuncNodesFromDiff(project, dir, df, maxDepth)
			if len(children) > 0 {
				node.Children = children
			}
		}

		tree.Roots = append(tree.Roots, node)
	}

	return tree, nil
}

// buildFuncNodesFromDiff extracts function declarations from a changed Go file
// and creates child nodes with trace information.
func buildFuncNodesFromDiff(project *goast.Project, dir string, df *GitDiffFile, maxDepth int) []*FlowNode {
	// Find package that contains this file
	var nodes []*FlowNode
	pkgs := project.RawPackages()

	for pkgPath, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			pos := pkg.Fset.Position(file.Pos())
			if !strings.HasSuffix(pos.Filename, df.Path) {
				continue
			}

			// Collect sibling function names for dedup (layer 1)
			siblings := make(map[string]bool)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				siblings[fn.Name.Name] = true
			}

			// Per-file seen set for cross-branch dedup (layer 2)
			seen := make(map[string]bool)

			// Extract function declarations from this file
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}

				startPos := pkg.Fset.Position(fn.Pos())
				endPos := pkg.Fset.Position(fn.End())

				funcName := fn.Name.Name
				nodeType := "function"
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					nodeType = "method"
				}

				code, _ := readFileLines(startPos.Filename, startPos.Line, endPos.Line)

				funcDiff := extractFuncDiff(df.Content, startPos.Line, endPos.Line)

				// Skip functions with no actual diff
				if funcDiff == "" {
					continue
				}

				child := &FlowNode{
					ID:        fmt.Sprintf("func-%s-%s", pkgPath, funcName),
					Label:     funcName,
					Package:   pkgPath,
					File:      startPos.Filename,
					LineStart: startPos.Line,
					LineEnd:   endPos.Line,
					Code:      code,
					Diff:      funcDiff,
					NodeType:  nodeType,
				}

				// Mark this function as seen (it's a top-level node)
				seen[pkgPath+"."+funcName] = true

				// Add trace children with sibling filtering + seen dedup
				tracer := goast.NewTracer(project, goast.TraceConfig{MaxDepth: maxDepth})
				chain := tracer.Trace(pkgPath, funcName)
				if chain.Root != nil {
					for _, callChild := range chain.Root.Children {
						childNode := callNodeToFlowNode(project, callChild, siblings, seen)
						if childNode != nil {
							child.Children = append(child.Children, childNode)
						}
					}
				}

				nodes = append(nodes, child)
			}
		}
	}

	return nodes
}

// BuildCodebaseTree builds a flow tree from a project entry point.
// If entryPkg is empty, all packages are included.
func BuildCodebaseTree(project *goast.Project, entryPkg string, maxDepth int) (*FlowTree, error) {
	if maxDepth <= 0 {
		maxDepth = 4
	}

	tree := &FlowTree{
		Mode:  "codebase",
		Title: fmt.Sprintf("Codebase Review: %s", entryPkg),
	}

	if entryPkg == "" {
		tree.Title = "Codebase Review: all packages"
	}

	pkgs := project.RawPackages()
	tracer := goast.NewTracer(project, goast.TraceConfig{MaxDepth: maxDepth})

	for pkgPath, pkg := range pkgs {
		// Filter by entry package if specified
		if entryPkg != "" && !strings.Contains(pkgPath, entryPkg) {
			continue
		}

		for _, file := range pkg.Syntax {
			pos := pkg.Fset.Position(file.Pos())

			// Read entire file
			fileCode, err := readEntireFile(pos.Filename)
			if err != nil {
				continue
			}

			fileNode := &FlowNode{
				ID:       fmt.Sprintf("pkg-%s-%s", pkgPath, filepath.Base(pos.Filename)),
				Label:    fmt.Sprintf("%s/%s", pkg.Name, filepath.Base(pos.Filename)),
				Package:  pkgPath,
				File:     pos.Filename,
				Code:     fileCode,
				NodeType: "file",
			}

			// Collect sibling function names for dedup (layer 1)
			siblings := make(map[string]bool)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				siblings[fn.Name.Name] = true
			}

			// Per-file seen set for cross-branch dedup (layer 2)
			seen := make(map[string]bool)

			// Add function-level children with trace
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}

				startPos := pkg.Fset.Position(fn.Pos())
				endPos := pkg.Fset.Position(fn.End())

				funcName := fn.Name.Name
				nodeType := "function"
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					nodeType = "method"
				}

				code, _ := readFileLines(startPos.Filename, startPos.Line, endPos.Line)

				funcNode := &FlowNode{
					ID:        fmt.Sprintf("func-%s-%s", pkgPath, funcName),
					Label:     funcName,
					Package:   pkgPath,
					File:      startPos.Filename,
					LineStart: startPos.Line,
					LineEnd:   endPos.Line,
					Code:      code,
					NodeType:  nodeType,
				}

				// Mark this function as seen
				seen[pkgPath+"."+funcName] = true

				// Trace call chain with sibling filtering + seen dedup
				chain := tracer.Trace(pkgPath, funcName)
				if chain.Root != nil {
					for _, callChild := range chain.Root.Children {
						childNode := callNodeToFlowNode(project, callChild, siblings, seen)
						if childNode != nil {
							funcNode.Children = append(funcNode.Children, childNode)
						}
					}
				}

				fileNode.Children = append(fileNode.Children, funcNode)
			}

			tree.Roots = append(tree.Roots, fileNode)
		}
	}

	return tree, nil
}

// callNodeToFlowNode converts an ast.CallNode to a FlowNode.
// siblings: set of function labels declared in the same file (filtered out).
// seen: per-file dedup set (key: "pkg.FuncName"). First occurrence gets full
// code+children; subsequent occurrences get a minimal reference node.
func callNodeToFlowNode(project *goast.Project, cn *goast.CallNode, siblings map[string]bool, seen map[string]bool) *FlowNode {
	// Skip siblings — they already exist as top-level nodes in the same file
	if siblings[cn.Name] {
		return nil
	}

	key := cn.Package + "." + cn.Name
	isRef := seen[key]

	idPrefix := "call"
	if isRef {
		idPrefix = "ref"
	}

	node := &FlowNode{
		ID:       fmt.Sprintf("%s-%s-%s", idPrefix, cn.Package, cn.Name),
		Label:    cn.Name,
		Package:  cn.Package,
		NodeType: string(cn.Type),
	}

	// Populate source code for all nodes (including refs)
	if project != nil {
		pkgs := project.RawPackages()
		if pkg, ok := pkgs[cn.Package]; ok {
			for _, file := range pkg.Syntax {
				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || fn.Name.Name != cn.Name {
						continue
					}
					startPos := pkg.Fset.Position(fn.Pos())
					endPos := pkg.Fset.Position(fn.End())
					node.File = startPos.Filename
					node.LineStart = startPos.Line
					node.LineEnd = endPos.Line
					node.Code, _ = readFileLines(startPos.Filename, startPos.Line, endPos.Line)
				}
			}
		}
	}

	// Ref nodes: have code but no children (avoids duplication)
	if isRef {
		return node
	}
	seen[key] = true

	// Recurse children
	for _, child := range cn.Children {
		childNode := callNodeToFlowNode(project, child, siblings, seen)
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}

	return node
}

// readFileLines reads specific line range from a file.
func readFileLines(path string, start, end int) (string, error) {
	f, err := os.Open(path)
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
	return strings.Join(lines, "\n"), nil
}

// readEntireFile reads the entire contents of a file.
func readEntireFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
