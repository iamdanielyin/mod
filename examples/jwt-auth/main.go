package main

import (
	"log"
	"time"

	"github.com/iamdanielyin/mod"
)

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"-"` // Don't return password in JSON
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" validate:"required" desc:"用户名"`
	Password string `json:"password" validate:"required" desc:"密码"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User  User               `json:"user"`
	Token *mod.TokenResponse `json:"token"`
}

// UserInfoRequest represents user info request
type UserInfoRequest struct {
	// Empty struct for user info request
}

// UserInfoResponse represents user info response
type UserInfoResponse struct {
	User    User   `json:"user"`
	Message string `json:"message"`
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	// Empty struct for logout request
}

// LogoutResponse represents logout response
type LogoutResponse struct {
	Message string `json:"message"`
}

// RefreshRequest represents refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" desc:"刷新令牌"`
}

// Simple in-memory user store for demo
var users = map[string]User{
	"admin": {
		ID:       "1",
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
		Password: "admin123",
	},
	"user": {
		ID:       "2",
		Username: "user",
		Email:    "user@example.com",
		Role:     "user",
		Password: "user123",
	},
}

func main() {
	app := mod.New()

	// Enable optional JWT middleware
	app.UseOptionalJWT()

	// Register login service (no auth required)
	app.Register(mod.Service{
		Name:        "login",
		DisplayName: "用户登录",
		Description: "用户登录获取JWT令牌",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *LoginRequest, resp *LoginResponse) error {
			// Find user
			user, exists := users[req.Username]
			if !exists || user.Password != req.Password {
				ctx.Warn("Login failed for username:", req.Username)
				return mod.Reply(401, "用户名或密码错误")
			}

			// Generate JWT tokens
			tokens, err := app.GenerateJWT(
				user.ID,
				user.Username,
				user.Email,
				user.Role,
				map[string]any{
					"login_time": time.Now().Unix(),
					"login_ip":   ctx.IP(),
				},
			)
			if err != nil {
				ctx.WithFields(map[string]any{
					"user_id": user.ID,
					"error":   err.Error(),
				}).Error("生成JWT令牌失败")
				return mod.Reply(500, "生成令牌失败")
			}

			// Store token in cache
			tokenData := map[string]any{
				"user_id":    user.ID,
				"username":   user.Username,
				"email":      user.Email,
				"role":       user.Role,
				"login_time": time.Now().Unix(),
				"login_ip":   ctx.IP(),
			}

			if err := app.SetToken(tokens.AccessToken, tokenData); err != nil {
				ctx.WithFields(map[string]any{
					"user_id": user.ID,
					"error":   err.Error(),
				}).Warn("存储令牌到缓存失败")
			}

			resp.User = user
			resp.Token = tokens

			ctx.WithFields(map[string]any{
				"user_id":  user.ID,
				"username": user.Username,
			}).Info("用户登录成功")

			return nil
		}),
		Group: "认证",
		Sort:  1,
	})

	// Register logout service
	app.Register(mod.Service{
		Name:        "logout",
		DisplayName: "用户登出",
		Description: "注销JWT令牌",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *LogoutRequest, resp *LogoutResponse) error {
			// Get JWT token from context
			token := ctx.GetJWTToken()
			if token == "" {
				return mod.Reply(401, "未提供令牌")
			}

			// Revoke the token
			if err := app.RevokeJWT(token); err != nil {
				ctx.WithFields(map[string]any{
					"error": err.Error(),
				}).Error("撤销JWT令牌失败")
				return mod.Reply(500, "登出失败")
			}

			// Remove token from cache
			if err := app.RemoveToken(token); err != nil {
				ctx.WithFields(map[string]any{
					"error": err.Error(),
				}).Warn("从缓存移除令牌失败")
			}

			resp.Message = "登出成功"

			ctx.WithFields(map[string]any{
				"user_id": ctx.GetUserID(),
			}).Info("用户登出成功")

			return nil
		}),
		Group: "认证",
		Sort:  2,
	})

	// Register refresh token service
	app.Register(mod.Service{
		Name:        "refresh",
		DisplayName: "刷新令牌",
		Description: "使用刷新令牌获取新的访问令牌",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *RefreshRequest, resp *mod.TokenResponse) error {
			// Refresh the token
			tokens, err := app.RefreshJWT(req.RefreshToken)
			if err != nil {
				ctx.WithFields(map[string]any{
					"error": err.Error(),
				}).Error("刷新JWT令牌失败")
				return mod.Reply(401, "刷新令牌无效")
			}

			// Store new token in cache
			if err := app.SetToken(tokens.AccessToken, map[string]any{
				"refreshed_at": time.Now().Unix(),
			}); err != nil {
				ctx.WithFields(map[string]any{
					"error": err.Error(),
				}).Warn("存储刷新后的令牌失败")
			}

			*resp = *tokens

			ctx.Info("JWT令牌刷新成功")

			return nil
		}),
		Group: "认证",
		Sort:  3,
	})

	// Register user info service (requires JWT authentication)
	app.Register(mod.Service{
		Name:        "user_info",
		DisplayName: "获取用户信息",
		Description: "获取当前登录用户的信息",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *UserInfoRequest, resp *UserInfoResponse) error {
			// Check if user is authenticated
			if !ctx.IsAuthenticated() {
				return mod.Reply(401, "需要身份认证")
			}

			// Get user information from JWT claims
			userID := ctx.GetUserID()
			username := ctx.GetUsername()
			email := ctx.GetUserEmail()
			role := ctx.GetUserRole()

			resp.User = User{
				ID:       userID,
				Username: username,
				Email:    email,
				Role:     role,
			}
			resp.Message = "用户信息获取成功"

			ctx.WithFields(map[string]any{
				"user_id": userID,
			}).Info("获取用户信息")

			return nil
		}),
		Group: "用户",
		Sort:  1,
	})

	log.Println("JWT认证示例服务器启动中...")
	log.Println("可用端点:")
	log.Println("  POST /services/login      - 用户登录")
	log.Println("  POST /services/logout     - 用户登出")
	log.Println("  POST /services/refresh    - 刷新令牌")
	log.Println("  POST /services/user_info  - 获取用户信息")
	log.Println("  GET  /services/docs       - API文档")
	log.Println()
	log.Println("测试用户:")
	log.Println("  admin/admin123  - 管理员")
	log.Println("  user/user123    - 普通用户")
	log.Println()
	log.Println("使用示例:")
	log.Println("  curl -X POST http://localhost:8080/services/login \\")
	log.Println("    -H 'Content-Type: application/json' \\")
	log.Println("    -d '{\"username\":\"admin\",\"password\":\"admin123\"}'")

	app.Run()
}
