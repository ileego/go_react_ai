package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateTokenPair(t *testing.T) {
	cfg := Config{
		Secret:          "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}

	pair, err := GenerateTokenPair(cfg, 1, "user")
	if err != nil {
		t.Fatalf("generate token pair failed: %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("access token is empty")
	}
	if pair.RefreshToken == "" {
		t.Error("refresh token is empty")
	}
	if pair.AccessJTI == "" || pair.RefreshJTI == "" {
		t.Error("jti is empty")
	}
}

func TestParseAndValidateAccessToken(t *testing.T) {
	cfg := Config{
		Secret:          "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}

	pair, err := GenerateTokenPair(cfg, 42, "admin")
	if err != nil {
		t.Fatalf("generate token pair failed: %v", err)
	}

	claims, err := ParseAndValidate(pair.AccessToken, TokenTypeAccess, cfg.Secret)
	if err != nil {
		t.Fatalf("parse access token failed: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("want userID 42, got %d", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("want role admin, got %s", claims.Role)
	}
	if claims.Type != TokenTypeAccess {
		t.Errorf("want type access, got %s", claims.Type)
	}
}

func TestParseAndValidateWrongType(t *testing.T) {
	cfg := Config{
		Secret:          "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	pair, err := GenerateTokenPair(cfg, 1, "user")
	if err != nil {
		t.Fatalf("generate token pair failed: %v", err)
	}

	_, err = ParseAndValidate(pair.AccessToken, TokenTypeRefresh, cfg.Secret)
	if err != ErrWrongType {
		t.Fatalf("want ErrWrongType, got %v", err)
	}
}

func TestParseAndValidateExpired(t *testing.T) {
	cfg := Config{
		Secret:          "test-secret",
		AccessTokenTTL:  -time.Second,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	pair, err := GenerateTokenPair(cfg, 1, "user")
	if err != nil {
		t.Fatalf("generate token pair failed: %v", err)
	}

	_, err = ParseAndValidate(pair.AccessToken, TokenTypeAccess, cfg.Secret)
	if err != ErrTokenExpired {
		t.Fatalf("want ErrTokenExpired, got %v", err)
	}
}

func TestExtractJTI(t *testing.T) {
	cfg := Config{
		Secret:          "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	pair, err := GenerateTokenPair(cfg, 1, "user")
	if err != nil {
		t.Fatalf("generate token pair failed: %v", err)
	}

	jti, err := ExtractJTI(pair.AccessToken)
	if err != nil {
		t.Fatalf("extract jti failed: %v", err)
	}
	if jti != pair.AccessJTI {
		t.Errorf("want jti %s, got %s", pair.AccessJTI, jti)
	}
}

func TestParseWithInvalidSecret(t *testing.T) {
	cfg := Config{
		Secret:         "test-secret",
		AccessTokenTTL: 15 * time.Minute,
	}

	token, err := generateToken(cfg, 1, "user", TokenTypeAccess, "jti", cfg.AccessTokenTTL)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	_, err = ParseAndValidate(token, TokenTypeAccess, "wrong-secret")
	if err != ErrInvalidToken {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}

func TestClaimsRegisteredClaims(t *testing.T) {
	now := time.Now()
	claims := Claims{
		UserID: 1,
		Role:   "user",
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	_, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign claims failed: %v", err)
	}
}
