package review

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
)

// collectChangedFuncs extracts functions that have actual diffs from changed Go files.
// Only functions where extractFuncDiff returns non-empty are included.
func collectChangedFuncs(project *goast.Project, diffs []*GitDiffFile) []ChangedFunc {
	if project == nil {
		return nil
	}

	var result []ChangedFunc
	pkgs := project.RawPackages()

	for _, df := range diffs {
		if !strings.HasSuffix(df.Path, ".go") {
			continue
		}

		for pkgPath, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				pos := pkg.Fset.Position(file.Pos())
				if !strings.HasSuffix(pos.Filename, df.Path) {
					continue
				}

				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || fn.Body == nil {
						continue
					}

					startPos := pkg.Fset.Position(fn.Pos())
					endPos := pkg.Fset.Position(fn.End())

					funcDiff := extractFuncDiff(df.Content, startPos.Line, endPos.Line)
					if funcDiff == "" {
						continue
					}

					nodeType := "function"
					if fn.Recv != nil && len(fn.Recv.List) > 0 {
						nodeType = "method"
					}

					code, _ := readFileLines(startPos.Filename, startPos.Line, endPos.Line)

					result = append(result, ChangedFunc{
						Package:   pkgPath,
						Name:      fn.Name.Name,
						File:      startPos.Filename,
						LineStart: startPos.Line,
						LineEnd:   endPos.Line,
						Code:      code,
						FuncDiff:  funcDiff,
						NodeType:  nodeType,
						IsNew:     df.IsNew,
					})
				}
			}
		}
	}

	return result
}

// changedFuncKey returns a unique key for a changed function.
func changedFuncKey(cf ChangedFunc) string {
	return cf.Package + "." + cf.Name
}

// callEdge represents a directed edge from caller to callee in the call graph.
type callEdge struct {
	from    string // changedFuncKey of caller
	to      string // changedFuncKey of callee
	bridges []bridgeFunc // intermediate unchanged functions on the path
}

// bridgeFunc represents an unchanged function that connects two changed functions.
type bridgeFunc struct {
	Package   string
	Name      string
	File      string
	LineStart int
	LineEnd   int
	Code      string
	NodeType  string
}

// callGraph holds the directed edges between changed functions.
type callGraph struct {
	edges      []callEdge
	changedSet map[string]ChangedFunc // key → ChangedFunc
}

// buildCallGraph builds a directed graph of call relationships among changed functions.
// It uses Tracer.Trace() to discover paths between changed functions, including bridge nodes.
func buildCallGraph(project *goast.Project, changedFuncs []ChangedFunc, maxDepth int) *callGraph {
	graph := &callGraph{
		changedSet: make(map[string]ChangedFunc),
	}

	for _, cf := range changedFuncs {
		graph.changedSet[changedFuncKey(cf)] = cf
	}

	if project == nil {
		return graph
	}

	// Use a higher depth for entry detection to find connections
	traceDepth := maxDepth
	if traceDepth < 6 {
		traceDepth = 6
	}
	tracer := goast.NewTracer(project, goast.TraceConfig{MaxDepth: traceDepth})

	for _, cf := range changedFuncs {
		chain := tracer.Trace(cf.Package, cf.Name)
		if chain.Root == nil {
			continue
		}

		fromKey := changedFuncKey(cf)
		// Search trace tree for other changed functions
		searchTraceForChanged(project, chain.Root, fromKey, nil, graph)
	}

	return graph
}

// searchTraceForChanged recursively searches a trace tree for changed functions,
// recording edges and bridge functions along the path.
func searchTraceForChanged(project *goast.Project, node *goast.CallNode, fromKey string, path []bridgeFunc, graph *callGraph) {
	for _, child := range node.Children {
		childKey := child.Package + "." + child.Name
		if childKey == fromKey {
			continue // skip self
		}

		if _, isChanged := graph.changedSet[childKey]; isChanged {
			// Found a changed function — record edge with bridges
			edge := callEdge{
				from:    fromKey,
				to:      childKey,
				bridges: make([]bridgeFunc, len(path)),
			}
			copy(edge.bridges, path)
			graph.edges = append(graph.edges, edge)
		} else {
			// Not changed — it's a potential bridge, recurse
			bf := buildBridgeFunc(project, child)
			newPath := append(path, bf)
			searchTraceForChanged(project, child, fromKey, newPath, graph)
		}
	}
}

// buildBridgeFunc creates a bridgeFunc from a CallNode by reading its source.
func buildBridgeFunc(project *goast.Project, cn *goast.CallNode) bridgeFunc {
	bf := bridgeFunc{
		Package:  cn.Package,
		Name:     cn.Name,
		NodeType: string(cn.Type),
	}

	if project == nil {
		return bf
	}

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
				bf.File = startPos.Filename
				bf.LineStart = startPos.Line
				bf.LineEnd = endPos.Line
				bf.Code, _ = readFileLines(startPos.Filename, startPos.Line, endPos.Line)
				return bf
			}
		}
	}

	return bf
}

// findEntries returns changed functions with in-degree 0 in the call graph (not called by other changed functions).
func findEntries(graph *callGraph) []ChangedFunc {
	inDegree := make(map[string]int)
	for key := range graph.changedSet {
		inDegree[key] = 0
	}
	for _, edge := range graph.edges {
		inDegree[edge.to]++
	}

	var entries []ChangedFunc
	for key, deg := range inDegree {
		if deg == 0 {
			entries = append(entries, graph.changedSet[key])
		}
	}

	// Sort by name for deterministic output
	sort.Slice(entries, func(i, j int) bool {
		ki := changedFuncKey(entries[i])
		kj := changedFuncKey(entries[j])
		return ki < kj
	})

	return entries
}

// flowBuildResult holds the nodes for a single flow and which changed functions are covered.
type flowBuildResult struct {
	root    *FlowNode
	covered map[string]bool // set of changedFuncKeys included in this flow
}

// buildFlowRoots constructs flow tree roots from the call graph.
// Returns: chain flows, isolated flows, in order.
func buildFlowRoots(project *goast.Project, graph *callGraph, changedFuncs []ChangedFunc) []*FlowNode {
	entries := findEntries(graph)

	// Build adjacency list: from → [(to, bridges)]
	adj := make(map[string][]callEdge)
	for _, edge := range graph.edges {
		adj[edge.from] = append(adj[edge.from], edge)
	}

	// Track which changed functions are covered by chain flows
	covered := make(map[string]bool)

	var chainFlows []*FlowNode

	for i, entry := range entries {
		entryKey := changedFuncKey(entry)

		// Check if this entry actually has outgoing edges (is part of a chain)
		result := buildChainFlow(entry, entryKey, adj, graph.changedSet, i)
		if result == nil {
			continue
		}

		// Only create a chain flow if it covers more than just the entry
		if len(result.covered) > 1 {
			chainFlows = append(chainFlows, result.root)
			for k := range result.covered {
				covered[k] = true
			}
		}
	}

	// Isolated flows: changed functions not covered by any chain flow
	var isolatedFlows []*FlowNode
	for _, cf := range changedFuncs {
		key := changedFuncKey(cf)
		if covered[key] {
			continue
		}
		node := changedFuncToFlowNode(cf)
		node.ID = fmt.Sprintf("isolated-%s-%s", cf.Package, cf.Name)
		isolatedFlows = append(isolatedFlows, node)
	}

	// Sort isolated flows by key
	sort.Slice(isolatedFlows, func(i, j int) bool {
		return isolatedFlows[i].Package+"."+isolatedFlows[i].Label < isolatedFlows[j].Package+"."+isolatedFlows[j].Label
	})

	var roots []*FlowNode
	roots = append(roots, chainFlows...)
	roots = append(roots, isolatedFlows...)

	return roots
}

// buildChainFlow builds a flow tree starting from an entry function,
// following call edges to other changed functions with bridge nodes in between.
func buildChainFlow(entry ChangedFunc, entryKey string, adj map[string][]callEdge, changedSet map[string]ChangedFunc, flowIdx int) *flowBuildResult {
	covered := make(map[string]bool)
	covered[entryKey] = true

	entryNode := changedFuncToFlowNode(entry)
	entryNode.ID = fmt.Sprintf("flow-%d-entry-%s-%s", flowIdx, entry.Package, entry.Name)

	// BFS/DFS through the call graph from entry
	var children []*FlowNode
	visited := make(map[string]bool)
	visited[entryKey] = true

	var walk func(fromKey string)
	walk = func(fromKey string) {
		edges := adj[fromKey]
		for _, edge := range edges {
			if visited[edge.to] {
				continue
			}
			visited[edge.to] = true
			covered[edge.to] = true

			// Add bridge nodes
			for _, bf := range edge.bridges {
				bridgeNode := bridgeFuncToFlowNode(bf, flowIdx)
				children = append(children, bridgeNode)
			}

			// Add the target changed function
			targetCF := changedSet[edge.to]
			targetNode := changedFuncToFlowNode(targetCF)
			targetNode.ID = fmt.Sprintf("flow-%d-func-%s-%s", flowIdx, targetCF.Package, targetCF.Name)
			children = append(children, targetNode)

			// Continue walking from the target
			walk(edge.to)
		}
	}

	walk(entryKey)

	if len(children) == 0 {
		// Entry with no outgoing edges to other changed functions
		return &flowBuildResult{root: entryNode, covered: covered}
	}

	// Create a flow root that groups entry + children
	flowRoot := &FlowNode{
		ID:       fmt.Sprintf("flow-%d", flowIdx),
		Label:    entry.Name,
		NodeType: "file", // use "file" nodeType for proper tree rendering
		Code:     entry.Code,
		Package:  entry.Package,
		File:     entry.File,
	}

	// Entry function is the first child
	flowRoot.Children = append(flowRoot.Children, entryNode)
	flowRoot.Children = append(flowRoot.Children, children...)

	return &flowBuildResult{root: flowRoot, covered: covered}
}

// changedFuncToFlowNode converts a ChangedFunc to a FlowNode.
func changedFuncToFlowNode(cf ChangedFunc) *FlowNode {
	return &FlowNode{
		ID:        fmt.Sprintf("func-%s-%s", cf.Package, cf.Name),
		Label:     cf.Name,
		Package:   cf.Package,
		File:      cf.File,
		LineStart: cf.LineStart,
		LineEnd:   cf.LineEnd,
		Code:      cf.Code,
		Diff:      cf.FuncDiff,
		NodeType:  cf.NodeType,
		IsNew:     cf.IsNew,
	}
}

// bridgeFuncToFlowNode converts a bridgeFunc to a FlowNode with IsBridge=true.
func bridgeFuncToFlowNode(bf bridgeFunc, flowIdx int) *FlowNode {
	return &FlowNode{
		ID:        fmt.Sprintf("flow-%d-bridge-%s-%s", flowIdx, bf.Package, bf.Name),
		Label:     bf.Name,
		Package:   bf.Package,
		File:      bf.File,
		LineStart: bf.LineStart,
		LineEnd:   bf.LineEnd,
		Code:      bf.Code,
		NodeType:  bf.NodeType,
		IsBridge:  true,
	}
}

// buildNonCodeRoot creates a "Non-code Files" root node grouping non-Go changed files.
func buildNonCodeRoot(dir string, diffs []*GitDiffFile) *FlowNode {
	var children []*FlowNode

	for _, df := range diffs {
		if strings.HasSuffix(df.Path, ".go") {
			continue
		}

		child := &FlowNode{
			ID:       fmt.Sprintf("file-%s", strings.ReplaceAll(df.Path, "/", "-")),
			Label:    df.Path,
			File:     df.Path,
			Diff:     df.Content,
			NodeType: "file",
			IsNew:    df.IsNew,
		}
		children = append(children, child)
	}

	if len(children) == 0 {
		return nil
	}

	return &FlowNode{
		ID:       "non-code-files",
		Label:    "Non-code Files",
		NodeType: "file",
		Children: children,
	}
}
