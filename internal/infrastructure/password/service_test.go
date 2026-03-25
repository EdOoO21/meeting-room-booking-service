package password

import "testing"

func TestService_HashAndCompare(t *testing.T) {
	t.Parallel()

	svc := New()
	password := "strong-password-123"

	hash, err := svc.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned empty hash")
	}
	if hash == password {
		t.Fatal("Hash() returned original password")
	}

	if err := svc.Compare(hash, password); err != nil {
		t.Fatalf("Compare() error = %v, want nil for correct password", err)
	}
	if err := svc.Compare(hash, "wrong-password"); err == nil {
		t.Fatal("Compare() error = nil, want error for wrong password")
	}
}
