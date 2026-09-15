package domain

import "testing"

func TestNewField_Valid(t *testing.T) {
	f, err := NewField("entity123", "title", FieldTypeString)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if f.IsSystem {
		t.Fatal("expected non-system field for name title")
	}
}

func TestNewField_SystemNames(t *testing.T) {
	for _, name := range []string{"id", "email", "password"} {
		fieldType := FieldTypeString
		if name == "email" {
			fieldType = FieldTypeEmail
		}
		f, err := NewField("entity123", name, fieldType)
		if err != nil {
			t.Fatalf("expected no error for %q, got %v", name, err)
		}
		if !f.IsSystem {
			t.Fatalf("expected field %q to be system field", name)
		}
	}
}

func TestNewField_ReservedNameWrongType(t *testing.T) {
	_, err := NewField("entity123", "email", FieldTypeString)
	if err == nil {
		t.Fatal("expected error when field named email uses a non-email type")
	}
	domainErr, ok := err.(*Error)
	if !ok || domainErr.Code != CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", err)
	}
}

func TestNewField_ReservedNameCorrectType(t *testing.T) {
	f, err := NewField("entity123", "email", FieldTypeEmail)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !f.IsSystem {
		t.Fatal("expected email field to be system field")
	}
}

func TestNewField_InvalidType(t *testing.T) {
	_, err := NewField("entity123", "title", FieldType("unknown"))
	if err == nil {
		t.Fatal("expected error for invalid field type")
	}
	domainErr, ok := err.(*Error)
	if !ok || domainErr.Code != CodeInvalidFieldType {
		t.Fatalf("expected CodeInvalidFieldType, got %v", err)
	}
}

func TestNewField_InvalidName(t *testing.T) {
	cases := []string{"", "Title", "title-name", "title name"}
	for _, name := range cases {
		_, err := NewField("entity123", name, FieldTypeString)
		if err == nil {
			t.Fatalf("expected error for name %q", name)
		}
	}
}

func TestField_CanDelete(t *testing.T) {
	system, _ := NewField("entity123", "id", FieldTypeUUID)
	if err := system.CanDelete(); err != ErrSystemFieldCannotDelete {
		t.Fatalf("expected ErrSystemFieldCannotDelete, got %v", err)
	}

	custom, _ := NewField("entity123", "title", FieldTypeString)
	if err := custom.CanDelete(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
