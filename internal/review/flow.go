package review

import (
	"fmt"
	"strings"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
)

// ChangeType represents the change status of a call chain node.
type ChangeType string

const (
	ChangeModified  ChangeType = "modified"
	ChangeNew       ChangeType = "new"
	ChangeUnchanged ChangeType = "unchanged"
	ChangeRPC       ChangeType = "external-rpc"
	ChangeMQ        ChangeType = "mq-producer"
)

// FlowNode is a call chain node annotated with change information.
type FlowNode struct {
	Name     string      `json:"name" yaml:"name"`
	Package  string      `json:"package" yaml:"package"`
	Change   ChangeType  `json:"change" yaml:"change"`
	Children []*FlowNode `json:"children,omitempty" yaml:"children,omitempty"`
	// For RPC nodes
	ServiceName string `json:"service_name,omitempty" yaml:"service_name,omitempty"`
	MethodName  string `json:"method_name,omitempty" yaml:"method_name,omitempty"`
}

// FlowReview holds the complete flow-based review result.
type FlowReview struct {
	Flows      []*FlowEntry    `json:"flows" yaml:"flows"`
	Standalone []*ChangedFunc  `json:"standalone,omitempty" yaml:"standalone,omitempty"`
}

type FlowEntry struct {
	Route string    `json:"route" yaml:"route"`
	Root  *FlowNode `json:"root" yaml:"root"`
}

func (fr *FlowReview) String() string {
	var sb strings.Builder
	fmt.Fprintln(&sb, "Flow-Based Review")
	fmt.Fprintln(&sb, strings.Repeat("=", 50))

	for _, flow := range fr.Flows {
		fmt.Fprintf(&sb, "\n%s\n", flow.Route)
		printFlowNode(&sb, flow.Root, "", true)
	}

	if len(fr.Standalone) > 0 {
		fmt.Fprintln(&sb, "\nStandalone Changes (not in any flow):")
		for _, cf := range fr.Standalone {
			tag := "modified"
			if cf.IsNew {
				tag = "new"
			}
			fmt.Fprintf(&sb, "  [%s] %s.%s\n", tag, cf.Package, cf.FullName())
		}
	}

	return sb.String()
}

func printFlowNode(sb *strings.Builder, node *FlowNode, indent string, isRoot bool) {
	label := fmt.Sprintf("[%s] %s.%s", node.Change, node.Package, node.Name)
	if node.Change == ChangeRPC {
		label = fmt.Sprintf("[RPC] %s.%s", node.ServiceName, node.MethodName)
	} else if node.Change == ChangeMQ {
		label = fmt.Sprintf("[MQ] %s.%s", node.Package, node.Name)
	}

	if isRoot {
		fmt.Fprintln(sb, label)
	} else {
		fmt.Fprintln(sb, label)
	}

	for i, child := range node.Children {
		isLast := i == len(node.Children)-1
		prefix := "├── "
		nextIndent := indent + "│   "
		if isLast {
			prefix = "└── "
			nextIndent = indent + "    "
		}
		fmt.Fprintf(sb, "%s%s", indent+prefix, "")
		printFlowNode(sb, child, nextIndent, false)
	}
}

// BuildFlowReview creates a flow-based review by overlaying change info on call chains.
func BuildFlowReview(project *goast.Project, changedFuncs []*ChangedFunc) *FlowReview {
	// Build lookup map: "pkg.FuncName" -> change type
	changeMap := make(map[string]ChangeType)
	for _, cf := range changedFuncs {
		key := cf.Package + "." + cf.Name
		if cf.IsNew {
			changeMap[key] = ChangeNew
		} else {
			changeMap[key] = ChangeModified
		}
	}

	// Build call chains from routes
	rt := goast.DiscoverRoutes(project)
	tracer := goast.NewTracer(project, goast.TraceConfig{MaxDepth: 10})

	review := &FlowReview{}
	touchedFuncs := make(map[string]bool)

	for _, route := range rt.Routes {
		chain := tracer.TraceFromRoute(route)
		if chain.Root == nil {
			continue
		}

		flowRoot := annotateNode(chain.Root, changeMap, touchedFuncs)
		// Only include if any node in the chain has changes
		if hasChanges(flowRoot) {
			review.Flows = append(review.Flows, &FlowEntry{
				Route: route.Method + " " + route.Path + " → " + route.Handler,
				Root:  flowRoot,
			})
		}
	}

	// Collect standalone changes (not in any flow)
	for _, cf := range changedFuncs {
		key := cf.Package + "." + cf.Name
		if !touchedFuncs[key] {
			review.Standalone = append(review.Standalone, cf)
		}
	}

	return review
}

func annotateNode(node *goast.CallNode, changeMap map[string]ChangeType, touched map[string]bool) *FlowNode {
	fn := &FlowNode{
		Name:    node.Name,
		Package: node.Package,
	}

	key := node.Package + "." + node.Name
	touched[key] = true

	switch node.Type {
	case goast.NodeExternalRPC:
		fn.Change = ChangeRPC
		fn.ServiceName = node.ServiceName
		fn.MethodName = node.MethodName
	case goast.NodeMQProducer:
		fn.Change = ChangeMQ
	default:
		if ct, ok := changeMap[key]; ok {
			fn.Change = ct
		} else {
			fn.Change = ChangeUnchanged
		}
	}

	for _, child := range node.Children {
		fn.Children = append(fn.Children, annotateNode(child, changeMap, touched))
	}

	return fn
}

func hasChanges(node *FlowNode) bool {
	if node.Change == ChangeModified || node.Change == ChangeNew {
		return true
	}
	for _, child := range node.Children {
		if hasChanges(child) {
			return true
		}
	}
	return false
}
