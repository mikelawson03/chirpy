package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		return "", errors.New("No valid bearer token found")
	}

	token = strings.TrimPrefix(token, "Bearer ")
	return strings.TrimSpace(token), nil
}

func HashPassword(password string) (string, error) {
	if len(password) < 5 {
		return "", fmt.Errorf("Password should be at least 8 characters")
	}
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userid uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userid.String(),
	})

	return t.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	t, err := jwt.ParseWithClaims(tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	idString, err := t.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func MakeRefreshToken() string {
	s := make([]byte, 32)
	rand.Read(s)
	return hex.EncodeToString(s)
}

func GetApiKey(headers http.Header) (string, error) {
	key := headers.Get("Authorization")
	if key == "" {
		return "", errors.New("No Authorization header provided")
	}
	if !strings.HasPrefix(key, "No API Key found in Authorization header") {

	}

	key = strings.TrimPrefix(key, "ApiKey ")
	return key, nil
}
