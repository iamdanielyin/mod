package main

import (
	"github.com/iamdanielyin/mod"
	"github.com/iamdanielyin/mod/examples/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// 设置 logrus 为更友好的格式
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	app := mod.New()

	// 展示基本的服务注册
	app.Register(mod.Service{
		Name:        "basic_login",
		DisplayName: "基础登录",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			reply.Uid = "basic_user"
			reply.Token = "basic_token"
			return nil
		}),
	})

	// 展示错误处理
	app.Register(mod.Service{
		Name:        "error_demo",
		DisplayName: "错误处理演示",
		SkipAuth:    true,
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

	logrus.Info("Starting demo server on :3000")
	app.Run(":3000")
}