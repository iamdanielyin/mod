package main

import (
	"github.com/iamdanielyin/mod"
)

// User represents a user entity
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetUserRequest represents get user request
type GetUserRequest struct {
	ID string `json:"id" validate:"required" desc:"用户ID"`
}

// GetUserResponse represents get user response
type GetUserResponse struct {
	User User `json:"user"`
}

// CreateUserRequest represents create user request
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required" desc:"用户姓名"`
	Email string `json:"email" validate:"required,email" desc:"用户邮箱"`
}

// CreateUserResponse represents create user response
type CreateUserResponse struct {
	User    User   `json:"user"`
	Message string `json:"message"`
}

// Simple in-memory user store
var users = map[string]User{
	"1": {ID: "1", Name: "张三", Email: "zhangsan@example.com"},
	"2": {ID: "2", Name: "李四", Email: "lisi@example.com"},
}
var nextID = 3

func main() {
	app := mod.New()

	// Register get user service
	app.Register(mod.Service{
		Name:        "get_user",
		DisplayName: "获取用户信息",
		Description: "根据用户ID获取用户详细信息",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *GetUserRequest, resp *GetUserResponse) error {
			user, exists := users[req.ID]
			if !exists {
				return mod.Reply(404, "用户不存在")
			}
			resp.User = user
			return nil
		}),
		Group: "用户管理",
		Sort:  1,
	})

	// Register create user service
	app.Register(mod.Service{
		Name:        "create_user",
		DisplayName: "创建用户",
		Description: "创建新用户账户",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *CreateUserRequest, resp *CreateUserResponse) error {
			// Create new user
			newUser := User{
				ID:    string(rune(nextID + '0')),
				Name:  req.Name,
				Email: req.Email,
			}

			// Store user
			users[newUser.ID] = newUser
			nextID++

			resp.User = newUser
			resp.Message = "用户创建成功"

			return nil
		}),
		Group: "用户管理",
		Sort:  2,
	})

	app.Run(":8080")
}
