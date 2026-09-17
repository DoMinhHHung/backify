package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type mockProjectRepo struct {
	projects map[string]*domain.Project
}

func newMockProjectRepo() *mockProjectRepo {
	return &mockProjectRepo{projects: make(map[string]*domain.Project)}
}

func (m *mockProjectRepo) Create(ctx context.Context, project *domain.Project) error {
	for _, p := range m.projects {
		if p.Subdomain == project.Subdomain {
			return domain.ErrSubdomainTaken
		}
	}
	m.projects[project.ID] = project
	return nil
}

func (m *mockProjectRepo) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}
	return p, nil
}

func (m *mockProjectRepo) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Subdomain == subdomain {
			return p, nil
		}
	}
	return nil, domain.ErrProjectNotFound
}

func (m *mockProjectRepo) List(ctx context.Context) ([]*domain.Project, error) {
	var result []*domain.Project
	for _, p := range m.projects {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockProjectRepo) Update(ctx context.Context, project *domain.Project) error {
	if _, ok := m.projects[project.ID]; !ok {
		return domain.ErrProjectNotFound
	}
	m.projects[project.ID] = project
	return nil
}

type mockEntityRepo struct {
	entities map[string]*domain.Entity
}

func newMockEntityRepo() *mockEntityRepo {
	return &mockEntityRepo{entities: make(map[string]*domain.Entity)}
}

func (m *mockEntityRepo) Create(ctx context.Context, entity *domain.Entity) error {
	for _, e := range m.entities {
		if e.ProjectID == entity.ProjectID && e.Name == entity.Name {
			return domain.ErrEntityNameTaken
		}
	}
	m.entities[entity.ID] = entity
	return nil
}

func (m *mockEntityRepo) GetByID(ctx context.Context, id string) (*domain.Entity, error) {
	e, ok := m.entities[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return e, nil
}

func (m *mockEntityRepo) GetByProjectAndName(ctx context.Context, projectID, name string) (*domain.Entity, error) {
	for _, e := range m.entities {
		if e.ProjectID == projectID && e.Name == name {
			return e, nil
		}
	}
	return nil, domain.ErrEntityNotFound
}

func (m *mockEntityRepo) ListByProject(ctx context.Context, projectID string) ([]*domain.Entity, error) {
	var result []*domain.Entity
	for _, e := range m.entities {
		if e.ProjectID == projectID {
			result = append(result, e)
		}
	}
	return result, nil
}

type mockFieldRepo struct {
	fields map[string]*domain.Field
}

func newMockFieldRepo() *mockFieldRepo {
	return &mockFieldRepo{fields: make(map[string]*domain.Field)}
}

func (m *mockFieldRepo) Create(ctx context.Context, field *domain.Field) error {
	for _, f := range m.fields {
		if f.EntityID == field.EntityID && f.Name == field.Name {
			return domain.ErrFieldNameTaken
		}
	}
	m.fields[field.ID] = field
	return nil
}

func (m *mockFieldRepo) GetByID(ctx context.Context, id string) (*domain.Field, error) {
	f, ok := m.fields[id]
	if !ok {
		return nil, domain.ErrFieldNotFound
	}
	return f, nil
}

func (m *mockFieldRepo) GetByEntityAndName(ctx context.Context, entityID, name string) (*domain.Field, error) {
	for _, f := range m.fields {
		if f.EntityID == entityID && f.Name == name {
			return f, nil
		}
	}
	return nil, domain.ErrFieldNotFound
}

func (m *mockFieldRepo) ListByEntity(ctx context.Context, entityID string) ([]*domain.Field, error) {
	var result []*domain.Field
	for _, f := range m.fields {
		if f.EntityID == entityID {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *mockFieldRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.fields[id]; !ok {
		return domain.ErrFieldNotFound
	}
	delete(m.fields, id)
	return nil
}

type mockModuleRepo struct {
	modules       map[string]*domain.Module
	functions     map[string]string
	functionField map[string]bool
	usages        map[string][]port.FieldUsage
}

func newMockModuleRepo() *mockModuleRepo {
	return &mockModuleRepo{
		modules:       make(map[string]*domain.Module),
		functions:     make(map[string]string),
		functionField: make(map[string]bool),
		usages:        make(map[string][]port.FieldUsage),
	}
}

func (m *mockModuleRepo) Create(ctx context.Context, module *domain.Module) error {
	key := module.ProjectID + ":" + string(module.Name)
	if existing, ok := m.modules[key]; ok {
		module.ID = existing.ID
		return nil
	}
	m.modules[key] = module
	return nil
}

func (m *mockModuleRepo) GetByProjectAndName(ctx context.Context, projectID string, name domain.ModuleName) (*domain.Module, error) {
	key := projectID + ":" + string(name)
	module, ok := m.modules[key]
	if !ok {
		return nil, domain.ErrModuleNotFound
	}
	return module, nil
}

func (m *mockModuleRepo) ListByProject(ctx context.Context, projectID string) ([]*domain.Module, error) {
	var result []*domain.Module
	for _, mod := range m.modules {
		if mod.ProjectID == projectID {
			result = append(result, mod)
		}
	}
	return result, nil
}

func (m *mockModuleRepo) EnsureFunction(ctx context.Context, moduleID, functionName string) (string, error) {
	key := moduleID + ":" + functionName
	if id, ok := m.functions[key]; ok {
		return id, nil
	}
	id := "fn_" + key
	m.functions[key] = id
	return id, nil
}

func (m *mockModuleRepo) ToggleFunctionField(ctx context.Context, functionID, fieldID string, enabled bool) error {
	key := functionID + ":" + fieldID
	m.functionField[key] = enabled
	return nil
}

func (m *mockModuleRepo) ListFunctionsByModule(ctx context.Context, moduleID string) ([]port.FunctionConfig, error) {
	var result []port.FunctionConfig
	for key, functionID := range m.functions {
		prefix := moduleID + ":"
		if len(key) <= len(prefix) || key[:len(prefix)] != prefix {
			continue
		}
		fn := port.FunctionConfig{ID: functionID, Name: key[len(prefix):]}
		for fkey, enabled := range m.functionField {
			if !enabled {
				continue
			}
			fPrefix := functionID + ":"
			if len(fkey) > len(fPrefix) && fkey[:len(fPrefix)] == fPrefix {
				fn.EnabledFieldIDs = append(fn.EnabledFieldIDs, fkey[len(fPrefix):])
			}
		}
		result = append(result, fn)
	}
	return result, nil
}

func (m *mockModuleRepo) ListFieldUsages(ctx context.Context, fieldID string) ([]port.FieldUsage, error) {
	return m.usages[fieldID], nil
}

func (m *mockModuleRepo) DisableFieldEverywhere(ctx context.Context, fieldID string) error {
	delete(m.usages, fieldID)
	return nil
}

func (m *mockModuleRepo) DisableAndDeleteField(ctx context.Context, fieldID string) error {
	delete(m.usages, fieldID)
	return nil
}

type mockEventPublisher struct {
	events []port.Event
}

func (m *mockEventPublisher) Publish(ctx context.Context, event port.Event) error {
	m.events = append(m.events, event)
	return nil
}
