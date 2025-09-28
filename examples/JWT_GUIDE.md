# JWT 认证示例

这个示例演示了如何在 MOD 框架中使用 JWT (JSON Web Token) 进行用户认证和授权。

## 功能特性

- ✅ JWT 令牌生成和验证
- ✅ 用户登录/登出
- ✅ 令牌刷新
- ✅ 基于角色的访问控制 (RBAC)
- ✅ 令牌黑名单/撤销
- ✅ 多种缓存策略支持 (BigCache, BadgerDB, Redis)
- ✅ 灵活的中间件配置
- ✅ 完整的 API 文档

## 快速开始

### 1. 启动服务器

```bash
cd /Users/danielyin/Projects/github.com/mod/examples
MOD_PATH=jwt_mod.yml go run jwt_example.go
```

服务器将在 `http://localhost:8080` 启动。

### 2. 访问 API 文档

打开浏览器访问：`http://localhost:8080/services/docs`

### 3. 运行测试脚本

```bash
# 确保服务器正在运行，然后执行：
./test_jwt.sh
```

## API 端点

### 认证相关

| 端点 | 方法 | 描述 | 需要认证 |
|------|------|------|----------|
| `/services/login` | POST | 用户登录 | ❌ |
| `/services/logout` | POST | 用户登出 | ✅ |
| `/services/refresh` | POST | 刷新令牌 | ❌ |

### 用户数据

| 端点 | 方法 | 描述 | 需要认证 | 需要角色 |
|------|------|------|----------|----------|
| `/services/userinfo` | POST | 获取用户信息 | ✅ | 任意 |
| `/services/protected-data` | POST | 受保护的数据 | ✅ | 任意 |
| `/admin/data` | POST | 管理员数据 | ✅ | admin |

## 使用示例

### 1. 用户登录

```bash
curl -X POST http://localhost:8080/services/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "user": {
      "id": "1",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin"
    },
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "access_token_expires_in": 86400,
      "refresh_token_expires_in": 604800,
      "token_type": "Bearer"
    }
  },
  "msg": "success",
  "rid": "1234567890"
}
```

### 2. 访问受保护的资源

```bash
curl -X POST http://localhost:8080/services/userinfo \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_access_token>"
```

### 3. 刷新令牌

```bash
curl -X POST http://localhost:8080/services/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<your_refresh_token>"
  }'
```

### 4. 用户登出

```bash
curl -X POST http://localhost:8080/services/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_access_token>"
```

## 测试用户

示例中包含两个测试用户：

| 用户名 | 密码 | 角色 | 描述 |
|--------|------|------|------|
| `admin` | `admin123` | `admin` | 管理员用户，可访问所有端点 |
| `user` | `user123` | `user` | 普通用户，无法访问管理员端点 |

## 配置说明

JWT 配置在 `jwt_mod.yml` 文件中：

### JWT 基本配置

```yaml
token:
  jwt:
    enabled: true                              # 启用JWT功能
    secret_key: "your-super-secret-jwt-key"    # JWT签名密钥
    issuer: "jwt-example-app"                  # JWT发行者
    algorithm: "HS256"                         # 签名算法
    expire_duration: "24h"                     # 访问令牌过期时间
    refresh_expire_duration: "168h"            # 刷新令牌过期时间
```

### 令牌验证配置

```yaml
token:
  validation:
    enabled: true                              # 启用token验证
    skip_expired_check: false                  # 是否跳过过期检查
    cache_strategy: "bigcache"                 # 缓存策略
    cache_key_prefix: "jwt:"                   # 缓存键前缀
```

### 缓存配置

支持三种缓存策略：

1. **BigCache** (默认) - 内存缓存，高性能
2. **BadgerDB** - 嵌入式键值数据库，持久化存储
3. **Redis** - 分布式缓存，支持集群

## 中间件使用

### 全局 JWT 中间件

```go
// 要求所有路由都需要 JWT 认证
app.UseJWT()

// 可选 JWT 中间件 - 如果有 JWT 则验证，但不强制要求
app.UseOptionalJWT()
```

### 基于角色的访问控制

```go
// 只有 admin 角色可以访问
app.Post("/admin/data",
    mod.JWTMiddleware(app),
    mod.RoleMiddleware("admin"),
    handlerFunc)

// 多个角色都可以访问
app.Post("/manager/data",
    mod.JWTMiddleware(app),
    mod.RoleMiddleware("admin", "manager"),
    handlerFunc)
```

## 在服务中使用 JWT

### 检查认证状态

```go
func handleProtectedData(ctx *mod.Context, req *struct{}, resp *Response) error {
    // 检查用户是否已认证
    if !ctx.IsAuthenticated() {
        return mod.Reply(401, "Authentication required")
    }

    // 获取用户信息
    userID := ctx.GetUserID()
    username := ctx.GetUsername()
    role := ctx.GetUserRole()

    // 检查角色权限
    if !ctx.HasRole("admin") {
        return mod.Reply(403, "Admin access required")
    }

    return nil
}
```

### 令牌管理

```go
// 生成令牌
tokens, err := app.GenerateJWT(userID, username, email, role, extraData)

// 验证令牌
claims, err := app.ValidateJWT(tokenString)

// 刷新令牌
newTokens, err := app.RefreshJWT(refreshToken)

// 撤销令牌
err := app.RevokeJWT(tokenString)
```

## 安全考虑

1. **密钥管理**: 生产环境中请使用强密钥并定期轮换
2. **HTTPS**: 生产环境中务必使用 HTTPS 传输令牌
3. **令牌存储**: 客户端应安全存储令牌，避免 XSS 攻击
4. **令牌撤销**: 实现令牌黑名单机制处理用户登出
5. **过期时间**: 合理设置令牌过期时间，平衡安全性和用户体验

## 错误处理

常见错误码：

- `401`: 未认证或令牌无效
- `403`: 权限不足
- `400`: 请求参数错误
- `500`: 服务器内部错误

## 调试

启用调试日志查看 JWT 处理详情：

```yaml
logging:
  console:
    enabled: true
    level: "debug"  # 设置为 debug 级别
```

## 生产部署

1. 修改默认密钥和用户凭据
2. 使用环境变量管理敏感配置
3. 配置适当的缓存策略
4. 设置合理的令牌过期时间
5. 启用 HTTPS
6. 配置日志记录和监控

## 相关文档

- [MOD 框架文档](../README.md)
- [JWT 官方文档](https://jwt.io/)
- [golang-jwt 库文档](https://github.com/golang-jwt/jwt)