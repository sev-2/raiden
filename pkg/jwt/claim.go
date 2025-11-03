package jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Audience     string          `json:"aud"`
	ExpiresAt    int64           `json:"exp"`
	IssuedAt     int64           `json:"iat"`
	Issuer       string          `json:"iss"`
	Subject      string          `json:"sub"`
	Email        string          `json:"email"`
	Phone        string          `json:"phone"`
	AppMetadata  map[string]any  `json:"app_metadata"`
	UserMetadata map[string]any  `json:"user_metadata"`
	Role         string          `json:"role"`
	AAL          string          `json:"aal"`
	AMR          []AuthMethodRef `json:"amr"`
	SessionID    string          `json:"session_id"`
}

type AuthMethodRef struct {
	Method    string `json:"method"`
	Timestamp int64  `json:"timestamp"`
}

func (c *JWTClaims) ExpiryTime() time.Time { return time.Unix(c.ExpiresAt, 0) }
func (c *JWTClaims) IssuedTime() time.Time { return time.Unix(c.IssuedAt, 0) }

// ---------------------------------------------------------------------------
// Decode: No signature verification — just read claims from any JWT
// ---------------------------------------------------------------------------
func Decode(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return claims, nil
}

// ---------------------------------------------------------------------------
// DecodeTo: Decode into a provided generic struct (no signature validation)
// ---------------------------------------------------------------------------
func DecodeTo[T any](token string) (*T, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	var out T
	if err := json.Unmarshal(payload, &out); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Validate: Verify JWT signature using JWT secret
// ---------------------------------------------------------------------------
func Validate[T any](tokenStr, jwtSecret string) (*T, error) {
	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		// Expect HMAC signing
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt parse: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token signature or expired")
	}

	// Marshal → Unmarshal into generic type
	claims := token.Claims.(jwt.MapClaims)
	data, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("marshal claims: %w", err)
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("bind claims: %w", err)
	}
	return &result, nil
}
