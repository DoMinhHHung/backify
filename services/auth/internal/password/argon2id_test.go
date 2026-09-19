package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	encoded, err := Hash("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := Verify("correct-horse-battery-staple", encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("expected password to match its own hash")
	}

	ok, err = Verify("wrong-password", encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if ok {
		t.Fatal("expected wrong password not to match")
	}
}

func TestHash_SaltIsRandom(t *testing.T) {
	a, _ := Hash("same-password")
	b, _ := Hash("same-password")
	if a == b {
		t.Fatal("expected two hashes of the same password to differ (random salt)")
	}
}

func TestVerify_MalformedHash(t *testing.T) {
	cases := []string{
		"",
		"not-a-hash-at-all",
		"$argon2id$v=19$m=65536,t=3,p=4$onlyfourparts",
		"$bcrypt$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA",
	}
	for _, c := range cases {
		if _, err := Verify("anything", c); err != ErrHashMalformed {
			t.Fatalf("input %q: expected ErrHashMalformed, got %v", c, err)
		}
	}
}
