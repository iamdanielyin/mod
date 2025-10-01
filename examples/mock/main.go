package main

import (
	"github.com/iamdanielyin/mod"
)

// 用户信息请求
type GetUserRequest struct {
	UserID string `json:"user_id" validate:"required" desc:"用户ID"`
}

// 用户信息响应
type GetUserResponse struct {
	ID    string `json:"id" desc:"用户ID"`
	Name  string `json:"name" desc:"用户名"`
	Email string `json:"email" desc:"邮箱"`
}

// 用户列表请求
type ListUsersRequest struct {
	Page int `json:"page" validate:"min=1" desc:"页码"`
	Size int `json:"size" validate:"min=1,max=100" desc:"每页数量"`
}

// 用户列表响应
type ListUsersResponse struct {
	Users []GetUserResponse `json:"users" desc:"用户列表"`
	Total int               `json:"total" desc:"总数"`
}

// 订单信息请求
type GetOrderRequest struct {
	OrderID string `json:"order_id" validate:"required" desc:"订单ID"`
}

// 订单信息响应
type GetOrderResponse struct {
	ID     string  `json:"id" desc:"订单ID"`
	Amount float64 `json:"amount" desc:"金额"`
	Status string  `json:"status" desc:"状态"`
}

func main() {
	app := mod.New()

	// 用户管理分组 - 分组级别启用Mock
	app.Register(mod.Service{
		Name:        "get_user",
		DisplayName: "获取用户信息",
		Group:       "用户管理",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *GetUserRequest, resp *GetUserResponse) error {
			// 实际Handler（Mock启用时不会执行）
			resp.ID = req.UserID
			resp.Name = "实际用户名"
			resp.Email = "real@example.com"
			ctx.Info("执行实际Handler: get_user")
			return nil
		}),
	})

	app.Register(mod.Service{
		Name:        "list_users",
		DisplayName: "获取用户列表",
		Group:       "用户管理",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *ListUsersRequest, resp *ListUsersResponse) error {
			// 实际Handler（Mock启用时不会执行）
			resp.Users = []GetUserResponse{{ID: "1", Name: "实际用户", Email: "real@example.com"}}
			resp.Total = 1
			ctx.Info("执行实际Handler: list_users")
			return nil
		}),
	})

	// 订单管理分组 - 服务级别启用Mock（优先级高于分组配置）
	app.Register(mod.Service{
		Name:        "get_order",
		DisplayName: "获取订单信息",
		Group:       "订单管理",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *GetOrderRequest, resp *GetOrderResponse) error {
			// 实际Handler（Mock启用时不会执行）
			resp.ID = req.OrderID
			resp.Amount = 99.99
			resp.Status = "paid"
			ctx.Info("执行实际Handler: get_order")
			return nil
		}),
	})

	app.Info("Mock功能演示服务启动")
	app.Info("访问 http://localhost:8080/services/docs 查看API文档并测试Mock功能")
	app.Run()
}
