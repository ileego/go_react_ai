package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/middleware"
	"github.com/ileego/go_react_ai/internal/security"
	"github.com/ileego/go_react_ai/internal/service"
	apperrors "github.com/ileego/go_react_ai/pkg/errors"
	"github.com/ileego/go_react_ai/pkg/response"
	"golang.org/x/oauth2"
)

// OAuthStateStore OAuth state 存储接口，用于防止 CSRF
type OAuthStateStore interface {
	Save(ctx context.Context, state string, ttl time.Duration) error
	Verify(ctx context.Context, state string) (bool, error)
}

// AuthHandler 认证相关 HTTP 接口
type AuthHandler struct {
	authSvc    service.AuthService
	oauthCfg   *oauth2.Config
	oauthState OAuthStateStore
	rl         security.RateLimiter
}

// NewAuthHandler 创建 AuthHandler
func NewAuthHandler(
	authSvc service.AuthService,
	oauthCfg *oauth2.Config,
	oauthState OAuthStateStore,
	rl security.RateLimiter,
) *AuthHandler {
	return &AuthHandler{
		authSvc:    authSvc,
		oauthCfg:   oauthCfg,
		oauthState: oauthState,
		rl:         rl,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
}

// Register 用户注册
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), req.Email, req.Password, req.Nickname)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, toUserResponse(user))
}

// Login 用户登录
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Data(c, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60,
	})
}

// Refresh 刷新令牌
// POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authSvc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Data(c, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60,
	})
}

// Logout 用户登出
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), req.AccessToken, req.RefreshToken); err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c)
}

// Me 获取当前用户信息
// GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.FromError(c, apperrors.NewUnauthorized("未登录").WithCode("UNAUTHORIZED"))
		return
	}

	user, err := h.authSvc.Me(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Data(c, toUserResponse(user))
}

// GithubLogin GitHub OAuth2 登录入口
// GET /api/auth/github/login
func (h *AuthHandler) GithubLogin(c *gin.Context) {
	if h.oauthCfg == nil {
		response.FromError(c, apperrors.NewInternal("GitHub OAuth 未配置", nil))
		return
	}

	state, err := generateState()
	if err != nil {
		response.FromError(c, apperrors.NewInternal("生成 state 失败", err))
		return
	}

	if h.oauthState != nil {
		if err := h.oauthState.Save(c.Request.Context(), state, 5*time.Minute); err != nil {
			response.FromError(c, apperrors.NewInternal("保存 state 失败", err))
			return
		}
	}

	url := h.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusFound, url)
}

// GithubCallback GitHub OAuth2 回调
// GET /api/auth/github/callback
func (h *AuthHandler) GithubCallback(c *gin.Context) {
	if h.oauthCfg == nil {
		response.FromError(c, apperrors.NewInternal("GitHub OAuth 未配置", nil))
		return
	}

	state := c.Query("state")
	code := c.Query("code")

	if h.oauthState != nil {
		valid, err := h.oauthState.Verify(c.Request.Context(), state)
		if err != nil {
			response.FromError(c, apperrors.NewInternal("校验 state 失败", err))
			return
		}
		if !valid {
			response.FromError(c, apperrors.NewUnauthorized("state 校验失败").WithCode("INVALID_STATE"))
			return
		}
	}

	accessToken, refreshToken, err := h.authSvc.GithubLogin(c.Request.Context(), code)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Data(c, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60,
	})
}

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Role:      string(user.Role),
	}
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
