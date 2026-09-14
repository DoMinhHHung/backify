package domain

import (
	"errors"
	"testing"
)

func TestNewProject_Valid(t *testing.T) {
	p, err := NewProject("Shop App", "shop-app")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected generated ID")
	}
	if p.SchemaName != "proj_"+p.ID {
		t.Fatalf("expected schema name proj_%s, got %s", p.ID, p.SchemaName)
	}
	if p.Plan != PlanFree {
		t.Fatalf("expected default plan free, got %s", p.Plan)
	}
	if p.Status != ProjectStatusActive {
		t.Fatalf("expected default status active, got %s", p.Status)
	}
}

func TestNewProject_EmptyName(t *testing.T) {
	_, err := NewProject("", "shop-app")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestNewProject_InvalidSubdomain(t *testing.T) {
	cases := []string{"ab", "Shop-App", "shop_app", "shop app", "this-subdomain-is-way-too-long-to-be-valid"}
	for _, subdomain := range cases {
		_, err := NewProject("Shop App", subdomain)
		if err == nil {
			t.Fatalf("expected error for subdomain %q", subdomain)
		}
		var domainErr *Error
		if !errors.As(err, &domainErr) || domainErr.Code != CodeInvalidInput {
			t.Fatalf("expected CodeInvalidInput for subdomain %q, got %v", subdomain, err)
		}
	}
}

func TestPlan_Valid(t *testing.T) {
	valid := []Plan{PlanFree, PlanHalfCore, PlanOneCore, PlanTwoCore}
	for _, plan := range valid {
		if !plan.Valid() {
			t.Fatalf("expected plan %s to be valid", plan)
		}
	}
	if Plan("unknown").Valid() {
		t.Fatal("expected unknown plan to be invalid")
	}
}

func TestProject_ChangePlan(t *testing.T) {
	p, _ := NewProject("Shop App", "shop-app")
	if err := p.ChangePlan(PlanOneCore); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Plan != PlanOneCore {
		t.Fatalf("expected plan 1c-2g, got %s", p.Plan)
	}
	if err := p.ChangePlan("invalid"); err == nil {
		t.Fatal("expected error for invalid plan")
	}
}

func TestProject_StatusTransitions(t *testing.T) {
	p, _ := NewProject("Shop App", "shop-app")
	p.Suspend()
	if p.Status != ProjectStatusSuspended {
		t.Fatalf("expected suspended, got %s", p.Status)
	}
	p.Activate()
	if p.Status != ProjectStatusActive {
		t.Fatalf("expected active, got %s", p.Status)
	}
	p.MarkDeleted()
	if p.Status != ProjectStatusDeleted {
		t.Fatalf("expected deleted, got %s", p.Status)
	}
}
