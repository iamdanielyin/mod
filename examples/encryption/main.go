package main

import (
	"fmt"

	"github.com/iamdanielyin/mod"
)

// UserData represents user information
type UserData struct {
	ID       string `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,min=1"`
	Role     string `json:"role" validate:"required"`
	Salary   int    `json:"salary,omitempty"`   // 敏感数据
	Password string `json:"password,omitempty"` // 敏感数据
}

// CreateUserRequest represents create user request
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required" desc:"用户姓名"`
	Email    string `json:"email" validate:"required,email" desc:"用户邮箱"`
	Age      int    `json:"age" validate:"required,min=1" desc:"用户年龄"`
	Role     string `json:"role" validate:"required" desc:"用户角色"`
	Salary   int    `json:"salary,omitempty" desc:"用户薪资（敏感信息）"`
	Password string `json:"password,omitempty" desc:"用户密码（敏感信息）"`
}

// CreateUserResponse represents create user response
type CreateUserResponse struct {
	User    UserData `json:"user"`
	Message string   `json:"message"`
}

// GetUserRequest represents get user request
type GetUserRequest struct {
	ID string `json:"id" validate:"required" desc:"用户ID"`
}

// GetUserResponse represents get user response
type GetUserResponse struct {
	User    UserData `json:"user"`
	Message string   `json:"message"`
}

// PublicUserData represents public user information (non-sensitive)
type PublicUserData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
	Role  string `json:"role"`
}

// GetPublicUserRequest represents get public user request
type GetPublicUserRequest struct {
	ID string `json:"id" validate:"required" desc:"用户ID"`
}

// GetPublicUserResponse represents get public user response
type GetPublicUserResponse struct {
	User    PublicUserData `json:"user"`
	Message string         `json:"message"`
}

// Simple in-memory user store for demo
var users = map[string]UserData{
	"1": {
		ID:       "1",
		Name:     "Alice Admin",
		Email:    "alice@example.com",
		Age:      30,
		Role:     "admin",
		Salary:   100000,
		Password: "secret123",
	},
	"2": {
		ID:       "2",
		Name:     "Bob User",
		Email:    "bob@example.com",
		Age:      25,
		Role:     "user",
		Salary:   50000,
		Password: "password456",
	},
}

var nextID = 3

func main() {
	// Create app instance with encryption configuration
	app := mod.New()

	// Enable encryption middleware for all service routes
	app.UseEncryption()

	// Register create user service (with encryption - contains sensitive data)
	app.Register(mod.Service{
		Name:        "create_user",
		DisplayName: "创建用户",
		Description: "创建新用户，包含敏感信息（薪资、密码），需要加密传输",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *CreateUserRequest, resp *CreateUserResponse) error {
			// Create new user
			newUser := UserData{
				ID:       fmt.Sprintf("%d", nextID),
				Name:     req.Name,
				Email:    req.Email,
				Age:      req.Age,
				Role:     req.Role,
				Salary:   req.Salary,
				Password: req.Password,
			}

			// Store user
			users[newUser.ID] = newUser
			nextID++

			resp.User = newUser
			resp.Message = "用户创建成功"

			ctx.WithFields(map[string]any{
				"user_id": newUser.ID,
				"name":    newUser.Name,
				"role":    newUser.Role,
			}).Info("User created successfully")

			return nil
		}),
		Group: "用户管理（加密）",
		Sort:  1,
	})

	// Register get user service (with encryption - contains sensitive data)
	app.Register(mod.Service{
		Name:        "get_user",
		DisplayName: "获取用户详细信息",
		Description: "获取用户的完整信息，包含敏感数据（薪资、密码），需要加密传输",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *GetUserRequest, resp *GetUserResponse) error {
			// Find user
			user, exists := users[req.ID]
			if !exists {
				return mod.Reply(404, "用户不存在")
			}

			resp.User = user
			resp.Message = "用户信息获取成功"

			ctx.WithFields(map[string]any{
				"user_id": user.ID,
				"name":    user.Name,
			}).Info("User details retrieved")

			return nil
		}),
		Group: "用户管理（加密）",
		Sort:  2,
	})

	// Register get public user service (without encryption - no sensitive data)
	app.Register(mod.Service{
		Name:        "get_public_user",
		DisplayName: "获取公开用户信息",
		Description: "获取用户的公开信息，不包含敏感数据，无需加密传输",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *GetPublicUserRequest, resp *GetPublicUserResponse) error {
			// Find user
			user, exists := users[req.ID]
			if !exists {
				return mod.Reply(404, "用户不存在")
			}

			// Return only public information (no sensitive data)
			resp.User = PublicUserData{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Age:   user.Age,
				Role:  user.Role,
			}
			resp.Message = "公开用户信息获取成功"

			ctx.WithFields(map[string]any{
				"user_id": user.ID,
				"name":    user.Name,
			}).Info("Public user info retrieved")

			return nil
		}),
		Group: "用户管理（公开）",
		Sort:  1,
	})

	app.Info("Encryption Example Server Starting...")
	app.Info("Available endpoints:")
	app.Info("  POST /services/create_user     - Create user (encrypted)")
	app.Info("  POST /services/get_user        - Get user details (encrypted)")
	app.Info("  POST /services/get_public_user - Get public user info (not encrypted)")
	app.Info("  GET  /services/docs            - API documentation")
	app.Info("")
	app.Info("Encryption Configuration:")
	app.Info("  - Symmetric encryption enabled for 'create-user' and 'get-user' services")
	app.Info("  - 'get-public-user' service is in whitelist (no encryption)")
	app.Info("  - HMAC-SHA256 signature verification enabled")
	app.Info("")
	app.Info("Example usage:")
	app.Info("  # For encrypted services, you need to encrypt the request body and add signature")
	app.Info("  # For public services, use normal JSON requests")
	app.Info("  curl -X POST http://localhost:8080/services/get-public-user \\")
	app.Info("    -H 'Content-Type: application/json' \\")
	app.Info("    -d '{\"id\":\"1\"}'")

	app.Run()
}
