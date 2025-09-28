package main

import (
	"github.com/iamdanielyin/mod"
	"github.com/iamdanielyin/mod/examples/types"
)

func main() {
	app := mod.New()

	// 统一的结构体注册方式
	app.Register(mod.Service{
		Name:        "admin_login",
		DisplayName: "管理员登录",
		SkipAuth:    true,
		Description: "管理员登录接口",
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			c.Info("Admin login:", args.Username)
			reply.Uid = "admin123"
			reply.Token = "admin_token"
			return nil
		}),
	})

	app.Register(mod.Service{
		Name:        "user_login",
		DisplayName: "用户登录",
		SkipAuth:    false,
		ReturnRaw:   true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			c.Info("User login:", args.Username)
			if args.Username == "" {
				return mod.ReplyWithDetail(400, "用户名不能为空", "Username field is required")
			}
			reply.Uid = "user123"
			reply.Token = "user_token"
			return nil
		}),
	})

	app.Register(mod.Service{
		Name:        "user_profile",
		DisplayName: "用户资料",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.UserArgs, reply *types.UserReply) error {
			c.Info("Get user profile:", args.UserID)
			reply.ID = args.UserID
			reply.Name = args.Name
			reply.Role = "user"
			return nil
		}),
	})

	// 批量注册（如果需要的话，可以使用循环）
	adminServices := []mod.Service{
		{
			Name:        "admin_users",
			DisplayName: "管理用户",
			SkipAuth:    false,
			Handler: mod.MakeHandler(func(c *mod.Context, args *types.UserArgs, reply *types.UserReply) error {
				reply.ID = args.UserID
				reply.Name = "Admin " + args.Name
				reply.Role = "admin"
				return nil
			}),
		},
		{
			Name:        "admin_settings",
			DisplayName: "管理设置",
			SkipAuth:    false,
			ReturnRaw:   true,
			Handler: mod.MakeHandler(func(c *mod.Context, args *types.UserArgs, reply *types.UserReply) error {
				reply.ID = args.UserID
				reply.Name = "Settings"
				reply.Role = "admin"
				return nil
			}),
		},
	}

	for _, svc := range adminServices {
		app.Register(svc)
	}

	app.Run(":8080")
}
