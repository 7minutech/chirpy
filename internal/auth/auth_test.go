package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {

	password1 := "password123"
	hash1, err := HashPassword(password1)
	if err != nil {
		t.Errorf("HashPassword(%s) err was not nil", password1)
	}

	password2 := "seasame456"
	hash2False, err := HashPassword(password1)
	if err != nil {
		t.Errorf("HashPassword(%s) err was not nil", password1)
	}

	hash2True, err := HashPassword(password2)
	if err != nil {
		t.Errorf("HashPassword(%s) err was not nil", password2)
	}

	cases := []struct {
		input struct {
			password string
			hash     string
		}
		expected bool
	}{
		{
			input: struct {
				password string
				hash     string
			}{password: password1, hash: hash1},
			expected: true,
		},
		{
			input: struct {
				password string
				hash     string
			}{password: password2, hash: hash2False},
			expected: false,
		},
		{
			input: struct {
				password string
				hash     string
			}{password: password2, hash: hash2True},
			expected: true,
		},
	}

	for _, c := range cases {

		ok, err := CheckPasswordHash(c.input.password, c.input.hash)
		if err != nil {
			t.Errorf("CheckPasswordHash(%s, %s) err was not nil", c.input.password, c.input.hash)
		}

		if ok != c.expected {
			t.Errorf("CheckPasswordHash(%s, %s) == %t", c.input.password, c.input.hash, ok)
		}
	}
}

func TestMakeAndValidateJWT_Success(t *testing.T) {
	userID := uuid.New()
	secret := "secret"
	tok, err := MakeJWT(userID, secret, time.Hour)

	if err != nil || tok == "" {
		t.Fatalf("expected token, got err=%v tok=%q", err, tok)
	}
	gotID, err := ValidateJWT(tok, secret)
	if err != nil {
		t.Fatalf("validate err: %v", err)
	}
	if gotID != userID {
		t.Fatalf("want %v, got %v", userID, gotID)
	}
}

func TestMakeAndValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "secret"
	tok, _ := MakeJWT(userID, secret, -time.Minute)

	if _, err := ValidateJWT(tok, secret); err == nil {
		t.Fatalf("expected error for expired token")
	}
}

func TestMakeAndValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	secret := "secret"
	tok, _ := MakeJWT(userID, secret, time.Hour)

	if _, err := ValidateJWT(tok, "wrong"); err == nil {
		t.Fatalf("expected error for wrong secret")
	}
}
