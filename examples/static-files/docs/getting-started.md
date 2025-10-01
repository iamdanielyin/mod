# 快速开始

## 安装

```bash
go get github.com/iamdanielyin/mod
```

## 创建服务

```go
package main

import "github.com/iamdanielyin/mod"

func main() {
    app := mod.New()

    app.Register(mod.Service{
        Name:        "hello",
        DisplayName: "Hello World",
        Handler: mod.MakeHandler(func(ctx *mod.Context, req *struct{}, resp *struct{
            Message string `json:"message"`
        }) error {
            resp.Message = "Hello, World!"
            return nil
        }),
    })

    app.Run()
}
```

## 更多文档

- [API 参考](api-reference.md)
- [配置指南](configuration.md)
