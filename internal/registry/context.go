package registry

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadContext loads the human-maintained context.yaml for a service.
func LoadContext(registryDir, serviceName string) (*ServiceContext, error) {
	path := filepath.Join(registryDir, serviceName, "context.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no context file is not an error
		}
		return nil, err
	}

	var ctx ServiceContext
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}

// LoadServiceInfo loads the auto-generated auto.yaml for a service.
func LoadServiceInfo(registryDir, serviceName string) (*ServiceInfo, error) {
	path := filepath.Join(registryDir, serviceName, "auto.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var info ServiceInfo
	if err := yaml.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// LoadIndex loads the registry index.
func LoadIndex(registryDir string) (*RegistryIndex, error) {
	path := filepath.Join(registryDir, "registry-index.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var idx RegistryIndex
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// MergeServiceInfo merges auto-generated and human context into a MergedService.
func MergeServiceInfo(info *ServiceInfo, ctx *ServiceContext) *MergedService {
	ms := &MergedService{
		Service: info.Service,
		IDLPath: info.IDLPath,
		Types:   info.Types,
	}

	if ctx != nil {
		ms.Notes = ctx.Notes
	}

	for _, m := range info.Methods {
		mm := &MergedMethod{
			Name:       m.Name,
			Request:    m.Request,
			Response:   m.Response,
			Exceptions: m.Exceptions,
		}

		if ctx != nil && ctx.Methods != nil {
			if mc, ok := ctx.Methods[m.Name]; ok {
				mm.Idempotent = mc.Idempotent
				mm.TimeoutMs = mc.TimeoutMs
				mm.Notes = mc.Notes
				mm.KnownIssues = mc.KnownIssues
			}
		}

		ms.Methods = append(ms.Methods, mm)
	}

	return ms
}
