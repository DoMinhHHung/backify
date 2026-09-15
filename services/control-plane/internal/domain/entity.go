package domain

import (
	"regexp"
	"time"
)

const SystemEntityUser = "user"

var entityNamePattern = regexp.MustCompile(`^[a-z0-9_]{1,50}$`)

type Entity struct {
	ID        string
	ProjectID string
	Name      string
	IsSystem  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func validateEntityName(name string) error {
	if !entityNamePattern.MatchString(name) {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "name",
			"reason": "must be lowercase alphanumeric with underscores, 1-50 chars",
		})
	}
	return nil
}

// NewEntity tạo entity với mã định danh và thời gian UTC mới. Entity có tên "user" được đánh dấu là entity hệ thống.
func NewEntity(projectID, name string) (*Entity, error) {
	if projectID == "" {
		return nil, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "project_id",
			"reason": "must not be empty",
		})
	}

	if err := validateEntityName(name); err != nil {
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &Entity{
		ID:        id,
		ProjectID: projectID,
		Name:      name,
		IsSystem:  name == SystemEntityUser,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// CanDelete trả về ErrSystemEntityCannotDelete nếu entity là entity hệ thống.
func (e *Entity) CanDelete() error {
	if e.IsSystem {
		return ErrSystemEntityCannotDelete
	}
	return nil
}
