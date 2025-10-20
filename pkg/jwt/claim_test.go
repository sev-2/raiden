package jwt_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	jwtpkg "github.com/sev-2/raiden/pkg/jwt"
	"github.com/stretchr/testify/require"
)

type sampleClaims struct {
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

func encodeSegment(t *testing.T, data any) string {
	t.Helper()
	bytes, err := json.Marshal(data)
	require.NoError(t, err)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func makeUnsignedToken(t *testing.T, payload any) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body := encodeSegment(t, payload)
	return fmt.Sprintf("%s.%s.signature", header, body)
}

func TestJWTClaimsTimeHelpers(t *testing.T) {
	claims := jwtpkg.JWTClaims{ExpiresAt: 1700, IssuedAt: 1600}
	require.Equal(t, time.Unix(1700, 0), claims.ExpiryTime())
	require.Equal(t, time.Unix(1600, 0), claims.IssuedTime())
}

func TestDecode(t *testing.T) {
	token := makeUnsignedToken(t, map[string]any{"role": "admin", "aud": "app"})
	claims, err := jwtpkg.Decode(token)
	require.NoError(t, err)
	require.Equal(t, "admin", claims["role"])

	_, err = jwtpkg.Decode("short")
	require.EqualError(t, err, "invalid token format")

	_, err = jwtpkg.Decode("abc.def?.ghi")
	require.ErrorContains(t, err, "base64 decode")

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte("{invalid"))
	_, err = jwtpkg.Decode(fmt.Sprintf("%s.%s.sig", header, payload))
	require.ErrorContains(t, err, "unmarshal")
}

func TestDecodeTo(t *testing.T) {
	expected := sampleClaims{Role: "member", Exp: 12345}
	token := makeUnsignedToken(t, expected)

	claims, err := jwtpkg.DecodeTo[sampleClaims](token)
	require.NoError(t, err)
	require.Equal(t, expected, *claims)

	_, err = jwtpkg.DecodeTo[sampleClaims]("invalid")
	require.EqualError(t, err, "invalid token format")

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte("not json"))
	_, err = jwtpkg.DecodeTo[sampleClaims](fmt.Sprintf("%s.%s.sig", header, payload))
	require.ErrorContains(t, err, "unmarshal")
}

func signToken(t *testing.T, method jwtv5.SigningMethod, claims jwtv5.MapClaims, secret string) string {
	t.Helper()
	token := jwtv5.NewWithClaims(method, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func TestValidate(t *testing.T) {
	secret := "top-secret"
	now := time.Now()
	claims := jwtv5.MapClaims{
		"role": "admin",
		"exp":  now.Add(time.Hour).Unix(),
		"iat":  now.Unix(),
		"aud":  "app",
	}

	token := signToken(t, jwtv5.SigningMethodHS256, claims, secret)
	parsed, err := jwtpkg.Validate[jwtpkg.JWTClaims](token, secret)
	require.NoError(t, err)
	require.Equal(t, "admin", parsed.Role)

	_, err = jwtpkg.Validate[jwtpkg.JWTClaims](token, "wrong-secret")
	require.ErrorContains(t, err, "jwt parse")

	badAlg := signToken(t, jwtv5.SigningMethodHS512, claims, secret)
	_, err = jwtpkg.Validate[jwtpkg.JWTClaims](badAlg, secret)
	require.ErrorContains(t, err, "unexpected signing method")

	_, err = jwtpkg.Validate[jwtpkg.JWTClaims]("not-a-token", secret)
	require.ErrorContains(t, err, "jwt parse")
}
