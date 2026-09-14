package domain

import "time"

type ModuleName string

const (
	ModuleAuth         ModuleName = "auth"
	ModuleCRUD         ModuleName = "crud"
	ModuleStorage      ModuleName = "storage"
	ModuleNotification ModuleName = "notification"
	ModulePayment      ModuleName = "payment"
)

func (m ModuleName) Valid() bool {
	switch m {
	case ModuleAuth, ModuleCRUD, ModuleStorage, ModuleNotification, ModulePayment:
		return true
	default:
		return false
	}
}

var authFunctions = []string{"signup", "signin", "forgotPassword", "oauth"}

var moduleFunctions = map[ModuleName][]string{
	ModuleAuth: authFunctions,
}

func FunctionsForModule(name ModuleName) []string {
	fns, ok := moduleFunctions[name]
	if !ok {
		return nil
	}
	result := make([]string, len(fns))
	copy(result, fns)
	return result
}

func ValidFunction(name ModuleName, function string) bool {
	for _, fn := range FunctionsForModule(name) {
		if fn == function {
			return true
		}
	}
	return false
}

type Module struct {
	ID        string
	ProjectID string
	Name      ModuleName
	CreatedAt time.Time
}

func NewModule(projectID string, name ModuleName) (*Module, error) {
	if projectID == "" {
		return nil, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "project_id",
			"reason": "must not be empty",
		})
	}

	if !name.Valid() {
		return nil, ErrInvalidModuleName.WithDetails(map[string]interface{}{
			"field": "name",
			"value": string(name),
		})
	}

	return &Module{
		ID:        generateID(),
		ProjectID: projectID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}, nil
}
