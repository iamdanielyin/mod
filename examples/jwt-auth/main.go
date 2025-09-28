package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
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
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User  User               `json:"user"`
	Token *mod.TokenResponse `json:"token"`
}

// UserInfoResponse represents user info response
type UserInfoResponse struct {
	User    User   `json:"user"`
	Message string `json:"message"`
}

// Protected data response
type ProtectedDataResponse struct {
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
	UserID    string `json:"user_id"`
}

// Admin data response
type AdminDataResponse struct {
	AdminData string         `json:"admin_data"`
	Stats     map[string]int `json:"stats"`
}

// Simple in-memory user store for demo
var users = map[string]User{
	"admin": {
		ID:       "1",
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
		Password: "admin123", // In real app, this should be hashed
	},
	"user": {
		ID:       "2",
		Username: "user",
		Email:    "user@example.com",
		Role:     "user",
		Password: "user123", // In real app, this should be hashed
	},
}

// Global app instance for access in handlers
var modApp *mod.App

func main() {
	// Create app instance - JWT configuration comes from mod.yml
	modApp = mod.New()
	app := modApp

	// Enable optional JWT middleware (validates JWT if present but allows services with SkipAuth)
	app.UseOptionalJWT() // This will validate JWT tokens when present but not require them for SkipAuth services

	// Register login service (no auth required)
	app.Register(mod.Service{
		Name:        "login",
		DisplayName: "用户登录",
		Description: "用户登录获取JWT令牌",
		SkipAuth:    true, // Skip token validation for login
		Handler:     mod.MakeHandler(handleLogin),
		Group:       "认证",
		Sort:        1,
	})

	// Register logout service (requires JWT authentication)
	app.Register(mod.Service{
		Name:        "logout",
		DisplayName: "用户登出",
		Description: "注销JWT令牌",
		SkipAuth:    true, // Skip traditional auth, use JWT context check
		Handler:     mod.MakeHandler(handleLogout),
		Group:       "认证",
		Sort:        2,
	})

	// Register refresh token service (no auth required as it uses refresh token)
	app.Register(mod.Service{
		Name:        "refresh",
		DisplayName: "刷新令牌",
		Description: "使用刷新令牌获取新的访问令牌",
		SkipAuth:    true,
		Handler:     mod.MakeHandler(handleRefresh),
		Group:       "认证",
		Sort:        3,
	})

	// Register user info service (requires JWT authentication)
	app.Register(mod.Service{
		Name:        "userinfo",
		DisplayName: "获取用户信息",
		Description: "获取当前登录用户的信息",
		SkipAuth:    true, // Skip traditional auth, use JWT context check
		Handler:     mod.MakeHandler(handleUserInfo),
		Group:       "用户",
		Sort:        1,
	})

	// Register protected data service (requires JWT authentication)
	app.Register(mod.Service{
		Name:        "protected-data",
		DisplayName: "受保护的数据",
		Description: "需要JWT认证才能访问的数据",
		SkipAuth:    true, // Skip traditional auth, use JWT context check
		Handler:     mod.MakeHandler(handleProtectedData),
		Group:       "数据",
		Sort:        1,
	})

	// Add a route with role-based access control using middleware
	app.Post("/admin/data", mod.JWTMiddleware(app), mod.RoleMiddleware("admin"), func(c *fiber.Ctx) error {
		ctx := &mod.Context{Ctx: c}
		response := AdminDataResponse{
			AdminData: "This is admin-only data",
			Stats: map[string]int{
				"total_users":   len(users),
				"admin_actions": 42,
				"system_uptime": 3600,
			},
		}
		return c.JSON(mod.NewSuccessResponse(ctx, response))
	})

	log.Println("JWT Example Server Starting...")
	log.Println("Available endpoints:")
	log.Println("  POST /services/login     - Login to get JWT token")
	log.Println("  POST /services/logout    - Logout (requires JWT)")
	log.Println("  POST /services/refresh   - Refresh JWT token")
	log.Println("  POST /services/userinfo  - Get user info (requires JWT)")
	log.Println("  POST /services/protected-data - Get protected data (requires JWT)")
	log.Println("  POST /admin/data         - Admin-only data (requires JWT + admin role)")
	log.Println("  GET  /services/docs      - API documentation")
	log.Println()
	log.Println("Example usage:")
	log.Println("  curl -X POST http://localhost:8080/services/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin123\"}'")
	log.Println("  curl -X POST http://localhost:8080/services/userinfo -H 'Authorization: Bearer <token>'")

	app.Run(":8080")
}

// Handle user login
func handleLogin(ctx *mod.Context, req *LoginRequest, resp *LoginResponse) error {
	// Find user
	user, exists := users[req.Username]
	if !exists || user.Password != req.Password {
		ctx.Warn("Login failed for username:", req.Username)
		return mod.Reply(401, "Invalid username or password")
	}

	// Generate JWT tokens
	tokens, err := modApp.GenerateJWT(
		user.ID,
		user.Username,
		user.Email,
		user.Role,
		map[string]interface{}{
			"login_time": time.Now().Unix(),
			"login_ip":   ctx.IP(),
		},
	)
	if err != nil {
		ctx.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate JWT tokens")
		return mod.Reply(500, "Failed to generate tokens")
	}

	// Store token in cache for validation (optional - depends on your token validation strategy)
	if err := modApp.SetToken(tokens.AccessToken, map[string]interface{}{
		"user_id":   user.ID,
		"username":  user.Username,
		"role":      user.Role,
		"issued_at": time.Now().Unix(),
	}); err != nil {
		ctx.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Warn("Failed to store token in cache")
	}

	resp.User = user
	resp.Token = tokens

	ctx.WithFields(map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
	}).Info("User logged in successfully")

	return nil
}

// Handle user logout
func handleLogout(ctx *mod.Context, req *struct{}, resp *struct{ Message string }) error {
	// Get JWT token from context
	token := ctx.GetJWTToken()
	if token == "" {
		return mod.Reply(401, "No token provided")
	}

	// Revoke the token
	if err := modApp.RevokeJWT(token); err != nil {
		ctx.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to revoke JWT token")
		return mod.Reply(500, "Failed to logout")
	}

	// Remove token from cache
	if err := modApp.RemoveToken(token); err != nil {
		ctx.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Failed to remove token from cache")
	}

	resp.Message = "Logged out successfully"

	ctx.WithFields(map[string]interface{}{
		"user_id": ctx.GetUserID(),
	}).Info("User logged out successfully")

	return nil
}

// Handle token refresh
func handleRefresh(ctx *mod.Context, req *struct{ RefreshToken string }, resp *mod.TokenResponse) error {
	if req.RefreshToken == "" {
		return mod.Reply(400, "Refresh token is required")
	}

	// Refresh the token
	tokens, err := modApp.RefreshJWT(req.RefreshToken)
	if err != nil {
		ctx.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to refresh JWT token")
		return mod.Reply(401, "Invalid refresh token")
	}

	// Store new token in cache
	if err := modApp.SetToken(tokens.AccessToken, map[string]interface{}{
		"refreshed_at": time.Now().Unix(),
	}); err != nil {
		ctx.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Failed to store refreshed token in cache")
	}

	*resp = *tokens

	ctx.Info("JWT token refreshed successfully")

	return nil
}

// Handle get user info
func handleUserInfo(ctx *mod.Context, req *struct{}, resp *UserInfoResponse) error {
	// Check if user is authenticated
	if !ctx.IsAuthenticated() {
		return mod.Reply(401, "Authentication required")
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
	resp.Message = "User info retrieved successfully"

	ctx.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("User info retrieved")

	return nil
}

// Handle protected data
func handleProtectedData(ctx *mod.Context, req *struct{}, resp *ProtectedDataResponse) error {
	// Check if user is authenticated
	if !ctx.IsAuthenticated() {
		return mod.Reply(401, "Authentication required")
	}

	userID := ctx.GetUserID()
	role := ctx.GetUserRole()

	// Provide different data based on role
	var data string
	switch role {
	case "admin":
		data = "This is admin-level protected data with full access"
	case "user":
		data = "This is user-level protected data with limited access"
	default:
		data = "This is basic protected data"
	}

	resp.Data = data
	resp.Timestamp = time.Now().Unix()
	resp.UserID = userID

	ctx.WithFields(map[string]interface{}{
		"user_id": userID,
		"role":    role,
	}).Info("Protected data accessed")

	return nil
}
