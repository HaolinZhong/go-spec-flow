package registry

import (
	"fmt"
	"strings"
)

// ServiceInfo represents the auto-generated registry info for a service.
type ServiceInfo struct {
	Service string       `json:"service" yaml:"service"`
	IDLPath string       `json:"idl_path" yaml:"idl_path"`
	Methods []*MethodInfo `json:"methods" yaml:"methods"`
	Types   []*TypeInfo  `json:"types,omitempty" yaml:"types,omitempty"`
}

type MethodInfo struct {
	Name       string       `json:"name" yaml:"name"`
	Request    []*FieldInfo `json:"request" yaml:"request"`
	Response   []*FieldInfo `json:"response" yaml:"response"`
	Exceptions []*FieldInfo `json:"exceptions,omitempty" yaml:"exceptions,omitempty"`
}

type FieldInfo struct {
	Name     string `json:"name" yaml:"name"`
	Type     string `json:"type" yaml:"type"`
	Required bool   `json:"required,omitempty" yaml:"required,omitempty"`
}

type TypeInfo struct {
	Name   string       `json:"name" yaml:"name"`
	Kind   string       `json:"kind" yaml:"kind"` // struct, enum, exception, typedef
	Fields []*FieldInfo `json:"fields,omitempty" yaml:"fields,omitempty"`
	Values []string     `json:"values,omitempty" yaml:"values,omitempty"` // for enums
}

// ServiceContext represents human-maintained behavioral context.
type ServiceContext struct {
	Service string                    `json:"service" yaml:"service"`
	Notes   string                    `json:"notes,omitempty" yaml:"notes,omitempty"`
	Methods map[string]*MethodContext `json:"methods,omitempty" yaml:"methods,omitempty"`
}

type MethodContext struct {
	Idempotent  *bool    `json:"idempotent,omitempty" yaml:"idempotent,omitempty"`
	TimeoutMs   int      `json:"timeout_ms,omitempty" yaml:"timeout_ms,omitempty"`
	Notes       string   `json:"notes,omitempty" yaml:"notes,omitempty"`
	KnownIssues []string `json:"known_issues,omitempty" yaml:"known_issues,omitempty"`
	ErrorCodes  map[int]string `json:"error_codes,omitempty" yaml:"error_codes,omitempty"`
}

// RegistryIndex holds the list of all registered services.
type RegistryIndex struct {
	Services []*ServiceEntry `json:"services" yaml:"services"`
}

type ServiceEntry struct {
	Name       string `json:"name" yaml:"name"`
	IDLPath    string `json:"idl_path" yaml:"idl_path"`
	HasContext bool   `json:"has_context" yaml:"has_context"`
}

func (ri *RegistryIndex) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-25s %-25s %s\n", "SERVICE", "IDL", "CONTEXT")
	fmt.Fprintf(&sb, "%-25s %-25s %s\n", "-------", "---", "-------")
	for _, s := range ri.Services {
		ctx := "no"
		if s.HasContext {
			ctx = "yes"
		}
		fmt.Fprintf(&sb, "%-25s %-25s %s\n", s.Name, s.IDLPath, ctx)
	}
	return sb.String()
}

// MergedService combines auto-generated and human-maintained context.
type MergedService struct {
	Service string                `json:"service" yaml:"service"`
	IDLPath string                `json:"idl_path" yaml:"idl_path"`
	Notes   string                `json:"notes,omitempty" yaml:"notes,omitempty"`
	Methods []*MergedMethod       `json:"methods" yaml:"methods"`
	Types   []*TypeInfo           `json:"types,omitempty" yaml:"types,omitempty"`
}

type MergedMethod struct {
	Name       string       `json:"name" yaml:"name"`
	Request    []*FieldInfo `json:"request" yaml:"request"`
	Response   []*FieldInfo `json:"response" yaml:"response"`
	Exceptions []*FieldInfo `json:"exceptions,omitempty" yaml:"exceptions,omitempty"`
	// Context (from human)
	Idempotent  *bool    `json:"idempotent,omitempty" yaml:"idempotent,omitempty"`
	TimeoutMs   int      `json:"timeout_ms,omitempty" yaml:"timeout_ms,omitempty"`
	Notes       string   `json:"notes,omitempty" yaml:"notes,omitempty"`
	KnownIssues []string `json:"known_issues,omitempty" yaml:"known_issues,omitempty"`
}

func (ms *MergedService) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Service: %s\n", ms.Service)
	fmt.Fprintf(&sb, "IDL: %s\n", ms.IDLPath)
	if ms.Notes != "" {
		fmt.Fprintf(&sb, "Notes: %s\n", ms.Notes)
	}
	fmt.Fprintln(&sb)

	for _, m := range ms.Methods {
		fmt.Fprintf(&sb, "  %s\n", m.Name)
		fmt.Fprintf(&sb, "    Request:\n")
		for _, f := range m.Request {
			req := ""
			if f.Required {
				req = " (required)"
			}
			fmt.Fprintf(&sb, "      - %s: %s%s\n", f.Name, f.Type, req)
		}
		fmt.Fprintf(&sb, "    Response:\n")
		for _, f := range m.Response {
			fmt.Fprintf(&sb, "      - %s: %s\n", f.Name, f.Type)
		}
		if len(m.Exceptions) > 0 {
			fmt.Fprintf(&sb, "    Throws:\n")
			for _, f := range m.Exceptions {
				fmt.Fprintf(&sb, "      - %s: %s\n", f.Name, f.Type)
			}
		}
		if m.Notes != "" {
			fmt.Fprintf(&sb, "    Notes: %s\n", m.Notes)
		}
		if m.Idempotent != nil {
			fmt.Fprintf(&sb, "    Idempotent: %v\n", *m.Idempotent)
		}
		if m.TimeoutMs > 0 {
			fmt.Fprintf(&sb, "    Timeout: %dms\n", m.TimeoutMs)
		}
	}

	if len(ms.Types) > 0 {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "Types:")
		for _, t := range ms.Types {
			fmt.Fprintf(&sb, "  %s (%s)\n", t.Name, t.Kind)
			for _, f := range t.Fields {
				fmt.Fprintf(&sb, "    - %s: %s\n", f.Name, f.Type)
			}
			for _, v := range t.Values {
				fmt.Fprintf(&sb, "    - %s\n", v)
			}
		}
	}

	return sb.String()
}
