package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/auth"
	"github.com/ileego/go_react_ai/internal/security"
	apperrors "github.com/ileego/go_react_ai/pkg/errors"
	"github.com/ileego/go_react_ai/pkg/response"
)

const (
	userIDContextKey   = "user_id"
	userRoleContextKey = "user_role"
)

// JWTAuth 解析并校验 Access Token，将 user_id 与 role 注入 gin 上下文
func JWTAuth(secret string, blacklist security.TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			response.FromError(c, apperrors.NewUnauthorized("缺少认证令牌").WithCode("MISSING_TOKEN"))
			c.Abort()
			return
		}

		claims, err := auth.ParseAndValidate(tokenString, auth.TokenTypeAccess, secret)
		if err != nil {
			if errors.Is(err, auth.ErrTokenExpired) {
				response.FromError(c, apperrors.NewUnauthorized("令牌已过期").WithCode("TOKEN_EXPIRED"))
			} else {
				response.FromError(c, apperrors.NewUnauthorized("无效的令牌").WithCode("INVALID_TOKEN"))
			}
			c.Abort()
			return
		}

		// 黑名单检查
		if blacklist != nil {
			blacklisted, err := blacklist.IsBlacklisted(c.Request.Context(), claims.ID)
			if err != nil {
				response.FromError(c, apperrors.NewInternal("检查令牌黑名单失败", err))
				c.Abort()
				return
			}
			if blacklisted {
				response.FromError(c, apperrors.NewUnauthorized("令牌已被吊销").WithCode("TOKEN_REVOKED"))
				c.Abort()
				return
			}
		}

		c.Set(userIDContextKey, claims.UserID)
		c.Set(userRoleContextKey, claims.Role)
		c.Next()
	}
}

// GetUserID 从 gin 上下文获取当前用户 ID
func GetUserID(c *gin.Context) int64 {
	if v, ok := c.Get(userIDContextKey); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// GetUserRole 从 gin 上下文获取当前用户角色
func GetUserRole(c *gin.Context) string {
	if v, ok := c.Get(userRoleContextKey); ok {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

// RequireRole 要求用户具有指定角色之一
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		response.FromError(c, apperrors.NewForbidden("权限不足").WithCode("FORBIDDEN"))
		c.Abort()
	}
}

// RequirePermission 要求用户具有指定权限
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if hasPermission(role, perm) {
			c.Next()
			return
		}
		response.FromError(c, apperrors.NewForbidden("权限不足").WithCode("FORBIDDEN"))
		c.Abort()
	}
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}

// hasPermission 基于角色判断是否具有指定权限
func hasPermission(role, perm string) bool {
	permissions := rolePermissions[role]
	for _, p := range permissions {
		if p == perm || p == "*" {
			return true
		}
	}
	return false
}

var rolePermissions = map[string][]string{
	"system": {"*"},
	"admin": {
		"reports:dispatch",
		"reports:read:any",
		"users:read:any",
		"users:role:update",
	},
	"user": {
		"reports:create",
		"reports:read:own",
		"reports:cancel:own",
	},
}
