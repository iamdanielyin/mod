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

	// 方式5: 批量注册示例
	services := []struct {
		name        string
		displayName string
		skipAuth    bool
		returnRaw   bool
	}{
		{"batch_login1", "批量登录1", true, false},
		{"batch_login2", "批量登录2", false, true},
		{"batch_login3", "批量登录3", true, true},
	}

	for _, s := range services {
		opts := []mod.ServiceOption{}
		if s.skipAuth {
			opts = append(opts, mod.WithSkipAuth())
		}
		if s.returnRaw {
			opts = append(opts, mod.WithReturnRaw())
		}

		mod.RegisterService(app, s.name, s.displayName, func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
			reply.Uid = "batch_user"
			reply.Token = "batch_token"
			return nil
		}, opts...)
	}

	logrus.Info("Starting server on :8080")
	app.Run(":8080")
}
