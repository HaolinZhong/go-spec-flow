package investigate

import (
	"fmt"
	"strings"

	goast "github.com/zhlie/go-spec-flow/internal/ast"
	"github.com/zhlie/go-spec-flow/internal/registry"
)

// Generator builds investigation reports.
type Generator struct {
	project     *goast.Project
	registryDir string
}

// NewGenerator creates a new investigation report generator.
func NewGenerator(project *goast.Project, registryDir string) *Generator {
	return &Generator{project: project, registryDir: registryDir}
}

// InvestigateRoutes generates a report for the given route patterns.
func (g *Generator) InvestigateRoutes(routePatterns []string) (*Report, error) {
	rt := goast.DiscoverRoutes(g.project)
	if len(rt.Routes) == 0 {
		return nil, fmt.Errorf("no routes found in project")
	}

	var matched []*goast.Route
	for _, route := range rt.Routes {
		for _, pattern := range routePatterns {
			routeStr := route.Method + " " + route.Path
			if strings.Contains(routeStr, pattern) || pattern == "*" {
				matched = append(matched, route)
				break
			}
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no routes matched patterns: %v", routePatterns)
	}

	return g.generateFromRoutes(matched)
}

// InvestigateAllRoutes generates a report for all discovered routes.
func (g *Generator) InvestigateAllRoutes() (*Report, error) {
	rt := goast.DiscoverRoutes(g.project)
	if len(rt.Routes) == 0 {
		return nil, fmt.Errorf("no routes found in project")
	}
	return g.generateFromRoutes(rt.Routes)
}

// InvestigateFunc generates a report from a specific function entry point.
func (g *Generator) InvestigateFunc(pkgPath, funcName string) (*Report, error) {
	tracer := goast.NewTracer(g.project, goast.TraceConfig{MaxDepth: 10})
	chain := tracer.Trace(pkgPath, funcName)

	report := &Report{
		Target: fmt.Sprintf("%s.%s", pkgPath, funcName),
		EntryPoints: []*EntryPoint{{
			Package: pkgPath,
			Func:    funcName,
		}},
	}

	g.populateFromChain(report, chain, fmt.Sprintf("%s.%s", pkgPath, funcName))
	return report, nil
}

func (g *Generator) generateFromRoutes(routes []*goast.Route) (*Report, error) {
	report := &Report{
		Target: fmt.Sprintf("%d route(s)", len(routes)),
	}

	tracer := goast.NewTracer(g.project, goast.TraceConfig{MaxDepth: 10})

	for _, route := range routes {
		routeStr := route.Method + " " + route.Path
		report.EntryPoints = append(report.EntryPoints, &EntryPoint{
			Route:   routeStr,
			Handler: route.Handler,
		})

		chain := tracer.TraceFromRoute(route)
		g.populateFromChain(report, chain, routeStr)
	}

	// Deduplicate modules
	report.Modules = deduplicateModules(report.Modules)

	return report, nil
}

func (g *Generator) populateFromChain(report *Report, chain *goast.CallChain, entryLabel string) {
	if chain.Root == nil {
		return
	}

	// Collect steps and modules from the call chain
	var steps []string
	modules := make(map[string][]string) // pkg -> functions
	externalRPCs := make(map[string][]string) // service -> methods
	var mqSteps []string

	var walk func(node *goast.CallNode)
	walk = func(node *goast.CallNode) {
		switch node.Type {
		case goast.NodeExternalRPC:
			step := fmt.Sprintf("[RPC] %s.%s", node.ServiceName, node.MethodName)
			steps = append(steps, step)
			externalRPCs[node.ServiceName] = appendUnique(externalRPCs[node.ServiceName], node.MethodName)
		case goast.NodeMQProducer:
			step := fmt.Sprintf("[MQ] %s.%s", node.Package, node.Name)
			steps = append(steps, step)
			mqSteps = append(mqSteps, step)
		default:
			step := fmt.Sprintf("%s.%s", node.Package, node.Name)
			steps = append(steps, step)
			modules[node.Package] = appendUnique(modules[node.Package], node.Name)
		}

		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(chain.Root)

	report.CallChains = append(report.CallChains, &CallChainInfo{
		Entry: entryLabel,
		Steps: steps,
	})

	// Add modules
	for pkg, funcs := range modules {
		report.Modules = append(report.Modules, &ModuleInfo{
			Package:   pkg,
			Role:      inferRole(pkg),
			Functions: funcs,
		})
	}

	// Add external dependencies with registry lookup
	for svc, methods := range externalRPCs {
		dep := &ExternalDep{
			Service:     svc,
			MethodsUsed: methods,
		}

		// Try to enrich from Service Registry
		if g.registryDir != "" {
			if info, err := registry.LoadServiceInfo(g.registryDir, svc); err == nil {
				dep.Notes = fmt.Sprintf("IDL: %s, %d methods total", info.IDLPath, len(info.Methods))
			}
			if ctx, err := registry.LoadContext(g.registryDir, svc); err == nil && ctx != nil {
				dep.Notes += fmt.Sprintf(" | %s", ctx.Notes)
			}
		}

		report.ExternalDeps = append(report.ExternalDeps, dep)
	}
}

func inferRole(pkgPath string) string {
	parts := strings.Split(pkgPath, "/")
	last := parts[len(parts)-1]
	switch last {
	case "handler":
		return "HTTP handler layer"
	case "service":
		return "Business logic"
	case "dal":
		return "Data access layer"
	case "rpc":
		return "RPC client wrapper"
	case "mq":
		return "Message queue"
	case "router":
		return "Route registration"
	default:
		return ""
	}
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}

func deduplicateModules(modules []*ModuleInfo) []*ModuleInfo {
	merged := make(map[string]*ModuleInfo)
	for _, m := range modules {
		if existing, ok := merged[m.Package]; ok {
			for _, fn := range m.Functions {
				existing.Functions = appendUnique(existing.Functions, fn)
			}
		} else {
			merged[m.Package] = &ModuleInfo{
				Package:   m.Package,
				Role:      m.Role,
				Functions: m.Functions,
			}
		}
	}

	var result []*ModuleInfo
	for _, m := range merged {
		result = append(result, m)
	}
	return result
}
