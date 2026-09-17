package postgres

import "testing"

func TestDatabaseName(t *testing.T) {
	name, err := databaseName("a1b2c3d4e5f60789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "auth_proj_a1b2c3d4e5f60789" {
		t.Fatalf("unexpected name: %s", name)
	}
}

func TestDatabaseName_RejectsUnsafeInput(t *testing.T) {
	cases := []string{
		"",
		"a; DROP DATABASE auth;--",
		"has space",
		"has\"quote",
		"has'quote",
	}
	for _, c := range cases {
		if _, err := databaseName(c); err == nil {
			t.Fatalf("expected error for input %q", c)
		}
	}
}
