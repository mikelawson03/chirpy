package auth

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "password"
	hashedpw, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	match, err := CheckPasswordHash(password, hashedpw)
	if err != nil {
		t.Fatal(err)
	}

	if !match {
		t.Fatal(`Password and Hash do not match`)
	}

}

func TestPasswordFail(t *testing.T) {
	password := "password"
	hashedpw, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	match, err := CheckPasswordHash("test", hashedpw)
	if err != nil {
		t.Fatal(err)
	}

	if match {
		t.Fatal(`Password and Hash should not match`)
	}
}

func TestBadPassword(t *testing.T) {
	password := ""
	_, err := HashPassword(password)
	if err == nil {
		t.Fatal(`Empty password should return error`)
	}
}

func TestJWT(t *testing.T) {
	uid := uuid.New()
	expiration, err := time.ParseDuration("48h")
	if err != nil {
		t.Fatalf(`Failed to make duration: %v`, err)
	}

	tok, err := MakeJWT(uid, "secret", expiration)
	if err != nil {
		t.Fatalf(`Error creating token: %v`, err)
	}

	id, err := ValidateJWT(tok, "secret")
	if err != nil {
		t.Fatalf(`Failed to validate token: %v`, err)
	}

	if uid != id {
		t.Fatalf(`Tokens do not match: %v`, err)
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	tok := "12345abcd"
	bearer := "Bearer  " + tok
	headers.Set("Authorization", bearer)
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf(`Failed to retrieve token: %v`, err)
	}

	if token != tok {
		t.Fatalf(`Incorrect token retrieved. Expected: %v, Received: %v`, tok, token)
	}
}

func TestEmptyBearerToken(t *testing.T) {
	headers := http.Header{}
	token, err := GetBearerToken(headers)
	if err != errors.New("No valid bearer token found") && token != "" {
		t.Fatalf(`Expected token and error. Received token: %v, Received error: %v`, token, err)
	}
}

func TestLoginWithTokens(t *testing.T) {
	s := MakeRefreshToken()
	if s == "" {
		t.Fatalf(`Expected token. Received empty string: %v`, s)
	}

}
