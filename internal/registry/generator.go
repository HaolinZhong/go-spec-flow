package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zhlie/go-spec-flow/internal/thrift"
	"gopkg.in/yaml.v3"
)

// GenerateFromIDL parses all Thrift files in idlDir and generates
// auto.yaml for each service into registryDir.
func GenerateFromIDL(idlDir, registryDir string) ([]*ServiceInfo, error) {
	docs, err := thrift.ParseDir(idlDir)
	if err != nil {
		return nil, fmt.Errorf("parsing IDL directory: %w", err)
	}

	var services []*ServiceInfo
	for _, doc := range docs {
		for _, svc := range doc.Services {
			info := convertService(svc, doc)
			info.IDLPath = doc.Filename
			services = append(services, info)

			// Write auto.yaml
			svcDir := filepath.Join(registryDir, svc.Name)
			if err := os.MkdirAll(svcDir, 0o755); err != nil {
				return nil, err
			}

			data, err := yaml.Marshal(info)
			if err != nil {
				return nil, err
			}
			if err := os.WriteFile(filepath.Join(svcDir, "auto.yaml"), data, 0o644); err != nil {
				return nil, err
			}
		}
	}

	// Write registry index
	index := &RegistryIndex{}
	for _, svc := range services {
		hasCtx := false
		ctxPath := filepath.Join(registryDir, svc.Service, "context.yaml")
		if _, err := os.Stat(ctxPath); err == nil {
			hasCtx = true
		}
		index.Services = append(index.Services, &ServiceEntry{
			Name:       svc.Service,
			IDLPath:    svc.IDLPath,
			HasContext: hasCtx,
		})
	}

	indexData, err := yaml.Marshal(index)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(registryDir, "registry-index.yaml"), indexData, 0o644); err != nil {
		return nil, err
	}

	return services, nil
}

func convertService(svc *thrift.Service, doc *thrift.Document) *ServiceInfo {
	info := &ServiceInfo{
		Service: svc.Name,
	}

	for _, m := range svc.Methods {
		mi := &MethodInfo{Name: m.Name}

		for _, p := range m.Params {
			mi.Request = append(mi.Request, &FieldInfo{
				Name:     p.Name,
				Type:     p.Type.String(),
				Required: p.Required,
			})
		}

		// For response, use the return type struct's fields if it's a struct reference
		if m.ReturnType != nil && m.ReturnType.Name != "void" {
			// Try to find the struct in the document
			retStruct := findStruct(doc, m.ReturnType.Name)
			if retStruct != nil {
				for _, f := range retStruct.Fields {
					mi.Response = append(mi.Response, &FieldInfo{
						Name: f.Name,
						Type: f.Type.String(),
					})
				}
			} else {
				mi.Response = append(mi.Response, &FieldInfo{
					Name: "result",
					Type: m.ReturnType.String(),
				})
			}
		}

		for _, t := range m.Throws {
			mi.Exceptions = append(mi.Exceptions, &FieldInfo{
				Name: t.Name,
				Type: t.Type.String(),
			})
		}

		info.Methods = append(info.Methods, mi)
	}

	// Collect types
	for _, s := range doc.Structs {
		ti := &TypeInfo{Name: s.Name, Kind: "struct"}
		for _, f := range s.Fields {
			ti.Fields = append(ti.Fields, &FieldInfo{Name: f.Name, Type: f.Type.String()})
		}
		info.Types = append(info.Types, ti)
	}
	for _, e := range doc.Enums {
		ti := &TypeInfo{Name: e.Name, Kind: "enum"}
		for _, v := range e.Values {
			ti.Values = append(ti.Values, fmt.Sprintf("%s = %d", v.Name, v.Value))
		}
		info.Types = append(info.Types, ti)
	}
	for _, ex := range doc.Exceptions {
		ti := &TypeInfo{Name: ex.Name, Kind: "exception"}
		for _, f := range ex.Fields {
			ti.Fields = append(ti.Fields, &FieldInfo{Name: f.Name, Type: f.Type.String()})
		}
		info.Types = append(info.Types, ti)
	}

	return info
}

func findStruct(doc *thrift.Document, name string) *thrift.Struct {
	for _, s := range doc.Structs {
		if s.Name == name {
			return s
		}
	}
	return nil
}
