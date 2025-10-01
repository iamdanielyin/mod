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
	"manager": {
		ID:       "3",
		Username: "manager",
		Email:    "manager@example.com",
		Role:     "manager",
		Password: "manager123", // In real app, this should be hashed
	},
	"vip": {
		ID:       "4",
		Username: "vip",
		Email:    "vip@example.com",
		Role:     "user",
		Password: "vip123", // In real app, this should be hashed
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
		Name:        "user_info",
		DisplayName: "获取用户信息",
		Description: "获取当前登录用户的信息",
		SkipAuth:    true, // Skip traditional auth, use JWT context check
		Handler:     mod.MakeHandler(handleUserInfo),
		Group:       "用户",
		Sort:        1,
	})

	// Register protected data service (requires JWT authentication)
	app.Register(mod.Service{
		Name:        "protected_data",
		DisplayName: "受保护的数据",
		Description: "需要JWT认证才能访问的数据",
		SkipAuth:    true, // Skip traditional auth, use JWT context check
		Handler:     mod.MakeHandler(handleProtectedData),
		Group:       "数据",
		Sort:        1,
	})

	// Register admin-only service using new permission system
	app.Register(mod.Service{
		Name:        "admin_data",
		DisplayName: "管理员专用数据",
		Description: "只有管理员角色才能访问的数据",
		SkipAuth:    true,
		Handler:     mod.MakeHandler(handleAdminData),
		Group:       "管理员功能",
		Sort:        1,
		Permission: &mod.PermissionConfig{
			Rules: []mod.PermissionRule{
				{Field: "user.role", Operator: "eq", Value: "admin"},
			},
			Logic: "AND",
		},
	})

	// Register VIP service using permission system
	app.Register(mod.Service{
		Name:        "vip_service",
		DisplayName: "VIP服务",
		Description: "需要VIP级别2或以上才能访问",
		SkipAuth:    true,
		Handler:     mod.MakeHandler(handleVipService),
		Group:       "VIP功能",
		Sort:        1,
		Permission: &mod.PermissionConfig{
			Rules: []mod.PermissionRule{
				{Field: "user.vip_level", Operator: "gte", Value: 2},
				{Field: "user.status", Operator: "eq", Value: "active"},
			},
			Logic: "AND",
		},
	})

	// Register multi-role service
	app.Register(mod.Service{
		Name:        "manager_or_admin_data",
		DisplayName: "管理层数据",
		Description: "管理员或经理角色可访问",
		SkipAuth:    true,
		Handler:     mod.MakeHandler(handleManagerOrAdminData),
		Group:       "管理功能",
		Sort:        1,
		Permission: &mod.PermissionConfig{
			Rules: []mod.PermissionRule{
				{Field: "user.role", Operator: "in", Value: []string{"admin", "manager"}},
			},
			Logic: "AND",
		},
	})

	log.Println("JWT Example Server Starting...")
	log.Println("Available endpoints:")
	log.Println("  POST /services/login              - Login to get JWT token")
	log.Println("  POST /services/logout             - Logout (requires JWT)")
	log.Println("  POST /services/refresh            - Refresh JWT token")
	log.Println("  POST /services/user_info          - Get user info (requires JWT)")
	log.Println("  POST /services/protected_data     - Get protected data (requires JWT)")
	log.Println("  POST /services/admin_data         - Admin-only data (requires admin role)")
	log.Println("  POST /services/vip_service        - VIP service (requires VIP level 2+)")
	log.Println("  POST /services/manager_or_admin_data - Management data (admin or manager role)")
	log.Println("  GET  /services/docs               - API documentation")
	log.Println()
	log.Println("Test users:")
	log.Println("  admin/admin123    - Admin role")
	log.Println("  manager/manager123 - Manager role")
	log.Println("  user/user123      - User role")
	log.Println("  vip/vip123        - VIP user")
	log.Println()
	log.Println("Example usage:")
	log.Println("  curl -X POST http://localhost:8080/services/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin123\"}'")
	log.Println("  curl -X POST http://localhost:8080/services/admin_data -H 'Authorization: Bearer <token>'")

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

	// Store token in cache for validation with rich user data
	tokenData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"status":   "active",
		},
		"session": map[string]interface{}{
			"issued_at":  time.Now().Unix(),
			"login_ip":   ctx.IP(),
			"login_time": time.Now().Unix(),
		},
	}

	// Add VIP level for VIP users (demo purpose)
	if user.Username == "vip" {
		tokenData["user"].(map[string]interface{})["vip_level"] = 3
	} else if user.Role == "admin" {
		tokenData["user"].(map[string]interface{})["vip_level"] = 5
	} else {
		tokenData["user"].(map[string]interface{})["vip_level"] = 1
	}

	if err := modApp.SetToken(tokens.AccessToken, tokenData); err != nil {
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

// Handle admin data
func handleAdminData(ctx *mod.Context, req *struct{}, resp *AdminDataResponse) error {
	resp.AdminData = "This is admin-only data accessible through permission system"
	resp.Stats = map[string]int{
		"total_users":   len(users),
		"admin_actions": 42,
		"system_uptime": 3600,
	}

	ctx.Info("Admin data accessed via permission system")
	return nil
}

// VIP service response
type VipServiceResponse struct {
	VipData   string `json:"vip_data"`
	VipLevel  int    `json:"vip_level"`
	Timestamp int64  `json:"timestamp"`
}

// Handle VIP service
func handleVipService(ctx *mod.Context, req *struct{}, resp *VipServiceResponse) error {
	resp.VipData = "This is VIP service content for level 2+ users"
	resp.VipLevel = 2
	resp.Timestamp = time.Now().Unix()

	ctx.Info("VIP service accessed")
	return nil
}

// Manager data response
type ManagerDataResponse struct {
	Data        string `json:"data"`
	AccessLevel string `json:"access_level"`
	Timestamp   int64  `json:"timestamp"`
}

// Handle manager or admin data
func handleManagerOrAdminData(ctx *mod.Context, req *struct{}, resp *ManagerDataResponse) error {
	resp.Data = "This is management-level data accessible to admin or manager roles"
	resp.AccessLevel = "management"
	resp.Timestamp = time.Now().Unix()

	ctx.Info("Management data accessed")
	return nil
}
