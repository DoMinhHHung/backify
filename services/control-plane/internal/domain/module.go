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

// Valid cho biết tên module có thuộc tập tên được nhận diện hay không.
func (m ModuleName) Valid() bool {
	switch m {
	case ModuleAuth, ModuleCRUD, ModuleStorage, ModuleNotification, ModulePayment:
		return true
	default:
		return false
	}
}

// Enabled cho biết module có được bật trong phiên bản MVP hay không.
func (m ModuleName) Enabled() bool {
	return m == ModuleAuth
}

var authFunctions = []string{"signup", "signin", "forgotPassword", "oauth"}

var moduleFunctions = map[ModuleName][]string{
	ModuleAuth: authFunctions,
}

// FunctionsForModule trả về bản sao danh sách hàm của module, hoặc nil nếu chưa định nghĩa.
func FunctionsForModule(name ModuleName) []string {
	fns, ok := moduleFunctions[name]
	if !ok {
		return nil
	}
	result := make([]string, len(fns))
	copy(result, fns)
	return result
}

// ValidFunction cho biết tên hàm có được định nghĩa cho module hay không.
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

// NewModule tạo module mới với mã định danh và thời điểm theo UTC sau khi kiểm tra đầu vào.
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

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	return &Module{
		ID:        id,
		ProjectID: projectID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}, nil
}
