package domain

import "testing"

func TestNewModule_Valid(t *testing.T) {
	m, err := NewModule("proj123", ModuleAuth)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Name != ModuleAuth {
		t.Fatalf("expected module name auth, got %s", m.Name)
	}
}

func TestNewModule_InvalidName(t *testing.T) {
	_, err := NewModule("proj123", ModuleName("unknown"))
	if err == nil {
		t.Fatal("expected error for invalid module name")
	}
	domainErr, ok := err.(*Error)
	if !ok || domainErr.Code != CodeInvalidModuleName {
		t.Fatalf("expected CodeInvalidModuleName, got %v", err)
	}
}

func TestNewModule_EmptyProjectID(t *testing.T) {
	_, err := NewModule("", ModuleAuth)
	if err == nil {
		t.Fatal("expected error for empty project id")
	}
}

func TestValidFunction_Auth(t *testing.T) {
	for _, fn := range []string{"signup", "signin", "forgotPassword", "oauth"} {
		if !ValidFunction(ModuleAuth, fn) {
			t.Fatalf("expected %q to be a valid auth function", fn)
		}
	}
	if ValidFunction(ModuleAuth, "unknownFunction") {
		t.Fatal("expected unknownFunction to be invalid")
	}
}

func TestValidFunction_UnspecifiedModule(t *testing.T) {
	if ValidFunction(ModuleCRUD, "anything") {
		t.Fatal("expected crud module to have no functions defined yet")
	}
}
