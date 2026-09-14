package domain

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

type Plan string

const (
	PlanFree     Plan = "free"
	PlanHalfCore Plan = "0.5c-512mb"
	PlanOneCore  Plan = "1c-2g"
	PlanTwoCore  Plan = "2c-4g"
)

func (p Plan) Valid() bool {
	switch p {
	case PlanFree, PlanHalfCore, PlanOneCore, PlanTwoCore:
		return true
	default:
		return false
	}
}

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusSuspended ProjectStatus = "suspended"
	ProjectStatusDeleted   ProjectStatus = "deleted"
)

func (s ProjectStatus) Valid() bool {
	switch s {
	case ProjectStatusActive, ProjectStatusSuspended, ProjectStatusDeleted:
		return true
	default:
		return false
	}
}

var subdomainPattern = regexp.MustCompile(`^[a-z0-9-]{3,30}$`)

type Project struct {
	ID         string
	Name       string
	Subdomain  string
	SchemaName string
	Plan       Plan
	Status     ProjectStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func validateSubdomain(subdomain string) error {
	if !subdomainPattern.MatchString(subdomain) {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "subdomain",
			"reason": "must be lowercase alphanumeric with hyphens, 3-30 chars",
		})
	}
	return nil
}

func NewProject(name, subdomain string) (*Project, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "name",
			"reason": "must not be empty",
		})
	}

	if err := validateSubdomain(subdomain); err != nil {
		return nil, err
	}

	id := generateID()
	now := time.Now().UTC()

	return &Project{
		ID:         id,
		Name:       trimmedName,
		Subdomain:  subdomain,
		SchemaName: "proj_" + id,
		Plan:       PlanFree,
		Status:     ProjectStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (p *Project) ChangePlan(plan Plan) error {
	if !plan.Valid() {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "plan",
			"reason": "unknown plan",
		})
	}
	p.Plan = plan
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Project) Suspend() {
	p.Status = ProjectStatusSuspended
	p.UpdatedAt = time.Now().UTC()
}

func (p *Project) Activate() {
	p.Status = ProjectStatusActive
	p.UpdatedAt = time.Now().UTC()
}

func (p *Project) MarkDeleted() {
	p.Status = ProjectStatusDeleted
	p.UpdatedAt = time.Now().UTC()
}
