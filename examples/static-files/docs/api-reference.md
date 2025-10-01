# API 参考

## 核心 API

### mod.New()
```go
app := mod.New()
```

### app.Register()
```go
app.Register(mod.Service{
    Name:        "service_name",
    DisplayName: "服务显示名称",
    Handler:     mod.MakeHandler(handlerFunc),
})
```

### app.Run()
```go
app.Run()  // 使用默认端口
app.Run(":3000")  // 指定端口
```

## 上下文 API

### 获取用户信息
```go
userID := ctx.GetUserID()
```

### 日志记录
```go
ctx.Info("操作成功")
ctx.Error("操作失败")
```

## JWT API

```go
// 生成令牌
tokens, err := app.GenerateJWT(userID, username, email, role, extraClaims)

// 验证令牌
claims, err := app.ValidateJWT(tokenString)

// 刷新令牌
newTokens, err := app.RefreshJWT(refreshToken)
```
