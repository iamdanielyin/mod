package main

import (
	"github.com/iamdanielyin/mod"
)

func main() {
	// 创建应用实例 - 默认CORS关闭
	app := mod.New()

	// 如果需要启用CORS，可以在mod.yml中配置：
	/*
		app:
		  cors:
		    enabled: true
		    allow_origins:
		      - "http://localhost:3000"
		      - "https://yourdomain.com"
		    allow_credentials: true
	*/

	// 注册一个测试服务
	app.Register(mod.Service{
		Name:        "hello",
		DisplayName: "Hello World",
		Description: "A simple hello world service",
		Handler: mod.Handler{
			Func: func(ctx *mod.Context, in interface{}, out interface{}) error {
				return ctx.JSON(map[string]string{
					"message": "Hello, World!",
					"cors":    "This endpoint supports CORS if enabled in config",
				})
			},
		},
	})

	// 启动服务器
	app.Run()
}
