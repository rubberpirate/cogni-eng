package access

import "testing"

func TestPasswordAndSecretBoundaries(t *testing.T) {
	hash, err := Password("original test phrase with 83!")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "original test phrase with 83!") || CheckPassword(hash, "different test phrase") {
		t.Fatal("password verification")
	}
	for _, p := range []string{"short", "passwordpassword123", "aaaaaaaaaaaaaaaa"} {
		if _, err := Password(p); err == nil {
			t.Fatalf("accepted weak password %q", p)
		}
	}
	key := make([]byte, 32)
	sealed, err := Seal(key, "secret")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Open(key, sealed)
	if err != nil || got != "secret" {
		t.Fatal("secret round trip")
	}
	sealed[len(sealed)-1] ^= 1
	if _, err = Open(key, sealed); err == nil {
		t.Fatal("tamper accepted")
	}
}
func TestRolesFailClosed(t *testing.T) {
	for _, role := range []string{"learner", "guardian", "reviewer", "instructor", "operator", "unknown"} {
		for _, cap := range []string{"organization.manage", "access.manage", "people.manage", "audit.read"} {
			if Grants(role, cap) {
				t.Fatalf("%s unexpectedly grants %s", role, cap)
			}
		}
	}
	if !Grants("coordinator", "people.manage") || Grants("coordinator", "access.manage") {
		t.Fatal("coordinator boundary")
	}
}
