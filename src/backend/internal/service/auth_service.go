// Package service 实现业务逻辑层。
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ileego/go_react_ai/internal/auth"
	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
	"github.com/ileego/go_react_ai/internal/security"
	apperrors "github.com/ileego/go_react_ai/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	bcryptCost        = 12
	minPasswordLength = 8
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// authService 实现 AuthService 接口
type authService struct {
	userRepo   repository.UserRepository
	jwtCfg     auth.Config
	oauthCfg   *oauth2.Config
	rl         security.RateLimiter
	blacklist  security.TokenBlacklist
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(
	userRepo repository.UserRepository,
	jwtCfg auth.Config,
	oauthCfg *oauth2.Config,
	rl security.RateLimiter,
	blacklist security.TokenBlacklist,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtCfg:    jwtCfg,
		oauthCfg:  oauthCfg,
		rl:        rl,
		blacklist: blacklist,
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, email, password, nickname string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailRegex.MatchString(email) {
		return nil, apperrors.NewValidation("email", "邮箱格式不正确")
	}
	if err := validatePassword(password); err != nil {
		return nil, apperrors.NewValidation("password", err.Error())
	}
	if strings.TrimSpace(nickname) == "" {
		return nil, apperrors.NewValidation("nickname", "昵称不能为空")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, apperrors.NewInternal("密码加密失败", err)
	}

	user := &domain.User{
		Email:    email,
		Nickname: nickname,
		Role:     domain.UserRoleUser,
	}

	if err := s.userRepo.CreateWithPassword(ctx, user, string(hash)); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, apperrors.NewDuplicate("user", email).WithCode("EMAIL_ALREADY_EXISTS")
		}
		return nil, apperrors.NewInternal("创建用户失败", err)
	}

	return user, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	key := "email:" + email

	// 检查是否被锁定
	allowed, lockout, err := s.rl.AllowLogin(ctx, key)
	if err != nil {
		return "", "", apperrors.NewInternal("检查登录锁定失败", err)
	}
	if !allowed {
		return "", "", apperrors.NewForbidden(fmt.Sprintf("登录失败次数过多，请 %v 后再试", lockout.Round(time.Second))).
			WithCode("LOGIN_LOCKED")
	}

	user, hash, err := s.userRepo.GetByEmailWithPassword(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.rl.RecordLoginFailure(ctx, key)
			return "", "", apperrors.NewUnauthorized("邮箱或密码错误").WithCode("INVALID_CREDENTIALS")
		}
		return "", "", apperrors.NewInternal("查询用户失败", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		s.rl.RecordLoginFailure(ctx, key)
		return "", "", apperrors.NewUnauthorized("邮箱或密码错误").WithCode("INVALID_CREDENTIALS")
	}

	// 登录成功，重置失败计数
	_ = s.rl.ResetLoginFailures(ctx, key)

	pair, err := auth.GenerateTokenPair(s.jwtCfg, user.ID, string(user.Role))
	if err != nil {
		return "", "", apperrors.NewInternal("生成令牌失败", err)
	}

	return pair.AccessToken, pair.RefreshToken, nil
}

// Refresh 刷新令牌
func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := auth.ParseAndValidate(refreshToken, auth.TokenTypeRefresh, s.jwtCfg.Secret)
	if err != nil {
		return "", "", apperrors.NewUnauthorized("无效的刷新令牌").WithCode("INVALID_REFRESH_TOKEN")
	}

	// 检查黑名单
	blacklisted, err := s.blacklist.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		return "", "", apperrors.NewInternal("检查黑名单失败", err)
	}
	if blacklisted {
		return "", "", apperrors.NewUnauthorized("刷新令牌已被吊销").WithCode("TOKEN_REVOKED")
	}

	// 生成新的令牌对
	pair, err := auth.GenerateTokenPair(s.jwtCfg, claims.UserID, claims.Role)
	if err != nil {
		return "", "", apperrors.NewInternal("生成令牌失败", err)
	}

	// 将旧的 refresh token 加入黑名单
	ttl := s.jwtCfg.RefreshTokenTTL
	_ = s.blacklist.Add(ctx, claims.ID, ttl)

	return pair.AccessToken, pair.RefreshToken, nil
}

// Logout 登出
func (s *authService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken != "" {
		if jti, err := auth.ExtractJTI(accessToken); err == nil && jti != "" {
			_ = s.blacklist.Add(ctx, jti, s.jwtCfg.AccessTokenTTL)
		}
	}
	if refreshToken != "" {
		if jti, err := auth.ExtractJTI(refreshToken); err == nil && jti != "" {
			_ = s.blacklist.Add(ctx, jti, s.jwtCfg.RefreshTokenTTL)
		}
	}
	return nil
}

// Me 获取当前用户信息
func (s *authService) Me(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperrors.NewNotFound("user", userID)
		}
		return nil, apperrors.NewInternal("查询用户失败", err)
	}
	return user, nil
}

// GithubLogin 使用 GitHub OAuth2 code 登录或注册
func (s *authService) GithubLogin(ctx context.Context, code string) (string, string, error) {
	if s.oauthCfg == nil {
		return "", "", apperrors.NewInternal("GitHub OAuth 未配置", nil)
	}

	token, err := s.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return "", "", apperrors.NewUnauthorized("GitHub 授权失败").WithCode("GITHUB_EXCHANGE_FAILED")
	}

	githubUser, err := s.fetchGithubUser(ctx, token.AccessToken)
	if err != nil {
		return "", "", apperrors.NewInternal("获取 GitHub 用户信息失败", err)
	}

	githubID := strconv.FormatInt(githubUser.ID, 10)
	user, err := s.userRepo.GetByGithubID(ctx, githubID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return "", "", apperrors.NewInternal("查询 GitHub 用户失败", err)
	}

	if user == nil {
		// 根据邮箱查找是否已存在用户
		existing, err := s.userRepo.GetByEmail(ctx, githubUser.Email)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return "", "", apperrors.NewInternal("查询用户邮箱失败", err)
		}

		if existing != nil {
			// 绑定 GitHub 信息到已有账号
			if err := s.userRepo.UpdateGithubInfo(ctx, existing.ID, githubID, githubUser.AvatarURL); err != nil {
				return "", "", apperrors.NewInternal("绑定 GitHub 账号失败", err)
			}
			user = existing
		} else {
			// 创建新用户
			newUser := &domain.User{
				Email:     strings.ToLower(githubUser.Email),
				Nickname:  githubUser.Login,
				AvatarURL: githubUser.AvatarURL,
				Role:      domain.UserRoleUser,
			}
			if err := s.userRepo.CreateWithGithub(ctx, newUser, githubID); err != nil {
				if errors.Is(err, repository.ErrDuplicate) {
					return "", "", apperrors.NewDuplicate("user", githubUser.Email).WithCode("EMAIL_ALREADY_EXISTS")
				}
				return "", "", apperrors.NewInternal("创建 GitHub 用户失败", err)
			}
			user = newUser
		}
	}

	pair, err := auth.GenerateTokenPair(s.jwtCfg, user.ID, string(user.Role))
	if err != nil {
		return "", "", apperrors.NewInternal("生成令牌失败", err)
	}

	return pair.AccessToken, pair.RefreshToken, nil
}

func (s *authService) fetchGithubUser(ctx context.Context, accessToken string) (*githubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api status %d: %s", resp.StatusCode, string(body))
	}

	var info githubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

type githubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
}

// NewGithubOAuthConfig 创建 GitHub OAuth2 配置
func NewGithubOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  redirectURL,
		Scopes:       []string{"read:user", "user:email"},
	}
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("密码长度至少 %d 位", minPasswordLength)
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("密码必须同时包含大写字母、小写字母和数字")
	}
	return nil
}
