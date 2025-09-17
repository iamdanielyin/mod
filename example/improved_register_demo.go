package main

import (
	"github.com/iamdanielyin/mod"
	"github.com/sirupsen/logrus"
)

type LoginArgs struct {
	Username string `validate:""`
	Password string `validate:"required"`
	Token    string `mod:"from=query"`
}

type LoginReply struct {
	Uid   string
	Token string
}

type UserArgs struct {
	UserID string `validate:"required"`
	Name   string
}

type UserReply struct {
	ID   string
	Name string
	Role string
}

func main() {
	// 设置 logrus 为更友好的格式
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	app := mod.New()

	// 方式1: 原始方式（保持兼容）
	svc1 := mod.NewService("original_login", "原始方式登录", func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
		reply.Uid = "user1"
		reply.Token = "token1"
		return nil
	})
	svc1.SkipAuth = true
	app.Register(svc1)

	// 方式2: 选项模式
	mod.RegisterService(app, "option_login", "选项模式登录", func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
		reply.Uid = "user2"
		reply.Token = "token2"
		return nil
	}, mod.WithSkipAuth(), mod.WithDescription("使用选项模式的登录接口"))

	// 方式3: 链式调用（传统）
	svc3 := mod.NewService("chain_login", "链式调用登录", func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
		reply.Uid = "user3"
		reply.Token = "token3"
		return nil
	})
	app.Register(*svc3.WithSkipAuth().WithDescription("使用链式调用的登录接口"))

	// 方式4: 构建器模式
	mod.AddService(app, "builder_login", "构建器模式登录", func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
		reply.Uid = "user4"
		reply.Token = "token4"
		return nil
	}).SkipAuth().ReturnRaw().Description("使用构建器模式的登录接口").Register()

	// 方式5a: 数组批量注册（包含完整的handler）
	loginServices := []mod.ServiceInfo[LoginArgs, LoginReply]{
		{
			Name:        "batch_admin_login",
			DisplayName: "批量管理员登录",
			Handler: func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
				logrus.Info("Admin login:", args.Username)
				reply.Uid = "admin123"
				reply.Token = "admin_token"
				return nil
			},
			SkipAuth:    true,
			Description: "管理员登录接口",
		},
		{
			Name:        "batch_user_login",
			DisplayName: "批量用户登录",
			Handler: func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
				logrus.Info("User login:", args.Username)
				reply.Uid = "user123"
				reply.Token = "user_token"
				return nil
			},
			SkipAuth:  false,
			ReturnRaw: true,
		},
		{
			Name:        "batch_guest_login",
			DisplayName: "批量访客登录",
			Handler: func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
				logrus.Info("Guest login")
				reply.Uid = "guest123"
				reply.Token = "guest_token"
				return nil
			},
			SkipAuth:    true,
			ReturnRaw:   true,
			Description: "访客登录接口",
		},
	}
	mod.RegisterServices(app, loginServices)

	// 方式5b: 链式批量注册（包含完整的handler）
	mod.BatchRegister[UserArgs, UserReply](app).
		Add("user_profile", "用户资料", func(c *mod.Context, args *UserArgs, reply *UserReply) error {
			reply.ID = args.UserID
			reply.Name = args.Name
			reply.Role = "user"
			return nil
		}).SkipAuth().Description("获取用户资料").
		Add("user_settings", "用户设置", func(c *mod.Context, args *UserArgs, reply *UserReply) error {
			reply.ID = args.UserID
			reply.Name = "Settings for " + args.Name
			reply.Role = "settings"
			return nil
		}).ReturnRaw().
		Add("user_permissions", "用户权限", func(c *mod.Context, args *UserArgs, reply *UserReply) error {
			reply.ID = args.UserID
			reply.Name = args.Name
			reply.Role = "admin"
			return nil
		}).SkipAuth().ReturnRaw().Description("获取用户权限").
		Register()

	logrus.Info("Starting server on :9000")
	app.Run(":9000")
}
