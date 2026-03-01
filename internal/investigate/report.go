package investigate

import (
	"fmt"
	"strings"
)

// Report holds the structured investigation output.
type Report struct {
	Target      string              `json:"target" yaml:"target"`
	EntryPoints []*EntryPoint       `json:"entry_points" yaml:"entry_points"`
	Modules     []*ModuleInfo       `json:"modules" yaml:"modules"`
	CallChains  []*CallChainInfo    `json:"call_chains" yaml:"call_chains"`
	ExternalDeps []*ExternalDep     `json:"external_dependencies,omitempty" yaml:"external_dependencies,omitempty"`
}

type EntryPoint struct {
	Route   string `json:"route,omitempty" yaml:"route,omitempty"`
	Handler string `json:"handler,omitempty" yaml:"handler,omitempty"`
	Package string `json:"package,omitempty" yaml:"package,omitempty"`
	Func    string `json:"func,omitempty" yaml:"func,omitempty"`
}

type ModuleInfo struct {
	Package   string   `json:"package" yaml:"package"`
	Role      string   `json:"role,omitempty" yaml:"role,omitempty"`
	Functions []string `json:"functions" yaml:"functions"`
}

type CallChainInfo struct {
	Entry string   `json:"entry" yaml:"entry"`
	Steps []string `json:"steps" yaml:"steps"`
}

type ExternalDep struct {
	Service     string   `json:"service" yaml:"service"`
	MethodsUsed []string `json:"methods_used" yaml:"methods_used"`
	Notes       string   `json:"notes,omitempty" yaml:"notes,omitempty"`
}

func (r *Report) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Investigation: %s\n", r.Target)
	fmt.Fprintln(&sb, strings.Repeat("=", 50))

	fmt.Fprintln(&sb, "\nEntry Points:")
	for _, ep := range r.EntryPoints {
		if ep.Route != "" {
			fmt.Fprintf(&sb, "  %s → %s\n", ep.Route, ep.Handler)
		} else {
			fmt.Fprintf(&sb, "  %s.%s\n", ep.Package, ep.Func)
		}
	}

	fmt.Fprintln(&sb, "\nModules Involved:")
	for _, m := range r.Modules {
		role := ""
		if m.Role != "" {
			role = fmt.Sprintf(" (%s)", m.Role)
		}
		fmt.Fprintf(&sb, "  %s%s\n", m.Package, role)
		for _, fn := range m.Functions {
			fmt.Fprintf(&sb, "    - %s\n", fn)
		}
	}

	fmt.Fprintln(&sb, "\nCall Chains:")
	for _, cc := range r.CallChains {
		fmt.Fprintf(&sb, "  %s:\n", cc.Entry)
		for _, step := range cc.Steps {
			fmt.Fprintf(&sb, "    → %s\n", step)
		}
	}

	if len(r.ExternalDeps) > 0 {
		fmt.Fprintln(&sb, "\nExternal Dependencies:")
		for _, dep := range r.ExternalDeps {
			fmt.Fprintf(&sb, "  %s: %s\n", dep.Service, strings.Join(dep.MethodsUsed, ", "))
			if dep.Notes != "" {
				fmt.Fprintf(&sb, "    Notes: %s\n", dep.Notes)
			}
		}
	}

	return sb.String()
}
