package main

import (
	"fmt"
	"github.com/iamdanielyin/mod"
	"github.com/iamdanielyin/mod/examples/types"
	"time"
)

func main() {
	app := mod.New() // 使用默认配置

	// 展示基本的服务注册
	app.Register(mod.Service{
		Name:        "basic_login",
		DisplayName: "基础登录",
		SkipAuth:    true,
		Description: "演示基础的登录功能",
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			reply.Uid = "basic_user"
			reply.Token = "basic_token_" + fmt.Sprintf("%d", time.Now().Unix())

			c.Info("用户登录成功", map[string]interface{}{
				"username": args.Username,
				"uid":      reply.Uid,
			})

			return nil
		}),
	})

	// 展示错误处理
	app.Register(mod.Service{
		Name:        "error_demo",
		DisplayName: "错误处理演示",
		SkipAuth:    true,
		Description: "演示API错误处理机制",
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			if args.Username == "error" {
				return mod.ReplyWithDetail(400, "演示错误", "This is a demo error")
			}
			reply.Uid = "demo_user"
			reply.Token = "demo_token"
			return nil
		}),
	})

	// 展示不同的配置选项
	app.Register(mod.Service{
		Name:        "config_demo",
		DisplayName: "配置演示",
		SkipAuth:    false,
		ReturnRaw:   true,
		Description: "展示不同配置选项的服务",
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.UserArgs, reply *types.UserReply) error {
			reply.ID = args.UserID
			reply.Name = "Config Demo"
			reply.Role = "demo"
			return nil
		}),
	})

	// 展示简单的数据验证
	app.Register(mod.Service{
		Name:        "validate_demo",
		DisplayName: "数据验证演示",
		SkipAuth:    true,
		Description: "演示简单的输入数据验证",
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			// 验证用户名长度
			if len(args.Username) < 3 {
				return mod.ReplyWithDetail(400, "用户名太短", "用户名至少需要3个字符")
			}
			if len(args.Username) > 20 {
				return mod.ReplyWithDetail(400, "用户名太长", "用户名不能超过20个字符")
			}

			reply.Uid = "validated_user"
			reply.Token = "validated_token"

			c.Info("数据验证通过", map[string]interface{}{
				"username": args.Username,
				"length":   len(args.Username),
			})

			return nil
		}),
	})

	app.Run(":3000")
}
