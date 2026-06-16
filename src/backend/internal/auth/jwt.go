package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType 标识令牌类型
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims 自定义 JWT Claims
type Claims struct {
	UserID int64     `json:"uid"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

// TokenPair 一对令牌及其 JTI
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessJTI    string
	RefreshJTI   string
}

// Config JWT 配置
type Config struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrWrongType    = errors.New("wrong token type")
)

// GenerateTokenPair 生成 Access + Refresh 双令牌
func GenerateTokenPair(cfg Config, userID int64, role string) (*TokenPair, error) {
	accessJTI := uuid.NewString()
	refreshJTI := uuid.NewString()

	accessToken, err := generateToken(cfg, userID, role, TokenTypeAccess, accessJTI, cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := generateToken(cfg, userID, role, TokenTypeRefresh, refreshJTI, cfg.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessJTI:    accessJTI,
		RefreshJTI:   refreshJTI,
	}, nil
}

func generateToken(cfg Config, userID int64, role string, tokenType TokenType, jti string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseAndValidate 解析并校验 JWT
func ParseAndValidate(tokenString string, expectedType TokenType, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != expectedType {
		return nil, ErrWrongType
	}

	return claims, nil
}

// ExtractJTI 从令牌字符串中提取 JTI（不校验签名，仅用于黑名单快速检查）
func ExtractJTI(tokenString string) (string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*Claims); ok {
		return claims.ID, nil
	}
	return "", ErrInvalidToken
}
