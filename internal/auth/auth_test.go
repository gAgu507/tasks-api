package auth

import (
	"testing"
	"time"
)

var secret = []byte("test-secret-test-secret-test-secret")

func TestRoundTrip(t *testing.T) {
	tok, err := NewToken(secret, 42, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	id, err := ParseToken(secret, tok)
	if err != nil || id != 42 {
		t.Fatalf("id = %d, err = %v, want 42, nil", id, err)
	}
}

func TestWrongSecret(t *testing.T) {
	tok, _ := NewToken(secret, 42, time.Hour)
	if _, err := ParseToken([]byte("another-secret-another-secret-123"), tok); err == nil {
		t.Error("token with wrong secret was accepted")
	}
}

func TestExpired(t *testing.T) {
	tok, _ := NewToken(secret, 42, -time.Minute)
	if _, err := ParseToken(secret, tok); err == nil {
		t.Error("expired token was accepted")
	}
}

func TestGarbage(t *testing.T) {
	if _, err := ParseToken(secret, "not.a.token"); err == nil {
		t.Error("garbage was accepted")
	}
}
