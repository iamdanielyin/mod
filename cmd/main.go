package main

import (
	"github.com/iamdanielyin/mod"
	"github.com/sirupsen/logrus"
)

type LoginArgs struct {
	Username string `validate:"required"`
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

	// 注册一个服务
	svc := mod.NewService("admin_login", "管理员登录", func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
		logrus.WithFields(logrus.Fields{
			"username": args.Username,
			"password": args.Password,
			"token":    args.Token,
		}).Info("Received login request")

		// 这里简化验证逻辑，只要有用户名就通过
		if args.Username == "" {
			logrus.Error("Username is empty")
			return mod.Reply(400, "用户名不能为空")
		}

		reply.Uid = "user123"
		reply.Token = "token456"
		return nil
	})
	svc.SkipAuth = true

	app.Register(svc)

	// 注册另一个服务
	svc2 := mod.NewService("user_register", "用户注册", func(c *mod.Context, args *LoginArgs, reply *LoginReply) error {
		reply.Uid = "newuser123"
		reply.Token = "newtoken456"
		return nil
	})

	app.Register(svc2)

	logrus.Info("Starting server on :8080")
	app.Run(":8080")
}
