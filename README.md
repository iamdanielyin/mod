# MOD

> 基于Go Fiber的现代化企业级Web应用框架，专注于快速开发、安全性和可扩展性

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Fiber Version](https://img.shields.io/badge/Fiber-v2.x-green.svg)](https://gofiber.io)
[![License](https://img.shields.io/badge/License-Apache2.0-green.svg)](LICENSE)

---

## 📋 目录

- [核心特性](#-核心特性)
- [快速开始](#-快速开始)
- [核心架构](#-核心架构)
- [功能特性](#-功能特性)
  - [JWT认证系统](#jwt认证系统)
  - [服务加解密](#服务加解密)
  - [文件服务](#文件服务)
  - [日志系统](#日志系统)
  - [Mock功能](#mock功能)
  - [缓存系统](#缓存系统)
- [配置系统](#-配置系统)
- [完整示例](#-完整示例)
- [配置参考](#-配置参考)
- [获取帮助](#-获取帮助)

---

## ✨ 核心特性

### 🚀 开发效率
- **服务化架构** - 基于服务注册的模块化开发，推荐使用蛇形命名法（snake_case）
- **自动API文档** - 内置API文档生成和交互式Web界面
- **参数验证** - 集成go-playground/validator，支持复杂验证规则
- **统一响应** - 标准化的JSON响应格式和错误处理

### 🔒 安全特性
- **JWT认证** - 完整的JWT认证系统，支持角色权限控制和Token管理
- **服务加解密** - 多级别的加解密配置，支持对称和非对称加密
- **数字签名** - HMAC-SHA256签名验证，确保数据完整性
- **白名单机制** - 灵活的服务和分组级白名单配置

### 🛠 企业功能
- **多后端日志** - 控制台、文件、Loki、阿里云SLS多种日志输出
- **文件上传** - 本地、S3、阿里云OSS多后端文件存储
- **静态文件** - 高性能静态文件服务和目录浏览
- **缓存系统** - BigCache、BadgerDB、Redis多种缓存方案

### 🔧 开发工具
- **Mock功能** - 智能Mock数据生成，支持全局、分组、服务级配置
- **中间件系统** - 丰富的内置中间件和灵活的自定义扩展
- **CORS支持** - 完善的跨域资源共享配置
- **热重载** - 开发环境友好的配置热加载

---

## 🚀 快速开始

### 安装

```bash
go get github.com/iamdanielyin/mod
```

### Hello World

```go
package main

import "github.com/iamdanielyin/mod"

// 定义请求和响应结构
type GetUserRequest struct {
    ID string `json:"id" validate:"required" desc:"用户ID"`
}

type GetUserResponse struct {
    Name  string `json:"name" desc:"用户姓名"`
    Email string `json:"email" desc:"用户邮箱"`
}

func main() {
    app := mod.New()

    // 注册服务（推荐使用蛇形命名法）
    app.Register(mod.Service{
        Name:        "get_user",
        DisplayName: "获取用户信息",
        Description: "根据用户ID获取用户详细信息",
        Handler: mod.MakeHandler(func(ctx *mod.Context, req *GetUserRequest, resp *GetUserResponse) error {
            resp.Name = "张三"
            resp.Email = "zhangsan@example.com"
            return nil
        }),
        Group: "用户管理",
    })

    app.Run(":8080")
}
```

启动后访问 [http://127.0.0.1:8080/services/docs](http://127.0.0.1:8080/services/docs) 查看自动生成的API文档。

---

## 🏗 核心架构

### 服务化设计

MOD采用服务化架构，每个业务功能都注册为独立的服务。**推荐使用蛇形命名法（snake_case）来命名服务**：

```go
app.Register(mod.Service{
    Name:        "get_user",              // 服务名称（推荐蛇形命名）
    DisplayName: "获取用户信息",            // 显示名称
    Description: "根据用户ID获取详细信息",   // 服务描述
    Handler:     mod.MakeHandler(fn),     // 处理函数
    Group:       "用户管理",               // 服务分组
    Sort:        1,                       // 排序
    SkipAuth:    false,                   // 是否跳过认证
    ReturnRaw:   false,                   // 是否返回原始数据
})
```

### 中间件系统

MOD提供了丰富的内置中间件，**所有全局中间件必须在注册服务之前调用**。

#### 中间件概览

MOD支持以下全局中间件：

| 中间件 | 功能说明 | 配置要求 |
|--------|----------|----------|
| [JWT认证中间件](#jwt认证中间件) | 提供JWT令牌认证功能 | 需要配置 `token.jwt` 部分 |
| [加解密中间件](#加解密中间件) | 自动处理请求解密和响应加密 | 需要配置 `encryption` 部分 |

#### 中间件执行顺序

```go
func main() {
    app := mod.New()

    // 推荐的中间件调用顺序
    app.UseEncryption()     // 1. 先处理加解密
    app.UseOptionalJWT()    // 2. 再处理认证

    // 注册服务...
    app.Run(":8080")
}
```

**执行顺序说明：**
- 🔐 **加解密中间件** - 首先解密请求数据
- 🔑 **JWT认证中间件** - 然后验证用户身份
- 📋 **服务权限检查** - 最后在服务处理前检查权限

---

#### JWT认证中间件

JWT认证中间件提供完整的用户身份认证功能，支持令牌生成、验证、刷新和撤销。

##### 可用方法

MOD提供两种JWT认证模式：

**🔒 强制认证模式**
```go
app.UseJWT()  // 所有请求必须提供有效JWT令牌
```

**🔓 可选认证模式（推荐）**
```go
app.UseOptionalJWT()  // 验证JWT但允许无令牌访问
```

##### 模式对比

| 特性 | UseJWT() | UseOptionalJWT() |
|------|----------|------------------|
| **缺少令牌时** | 返回 `401` 错误 | 继续执行 |
| **令牌无效时** | 返回 `401` 错误 | 继续执行 |
| **黑名单令牌** | 返回 `401` 错误 | 返回 `401` 错误 |
| **适用场景** | 严格认证的API | 混合公开/私有接口 |
| **推荐用途** | 企业内部系统 | Web应用、移动APP |

##### 使用策略

**🎯 策略一：灵活控制（推荐）**

适用于需要混合公开/私有接口的应用

```go
app.UseOptionalJWT()  // 全局可选认证

// 完全公开的接口
app.Register(mod.Service{
    Name:     "login",
    SkipAuth: true,  // 跳过所有认证检查
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *LoginRequest, resp *LoginResponse) error {
        // 登录逻辑，无需认证
        return nil
    }),
})

// 需要认证的接口
app.Register(mod.Service{
    Name:     "user_info",
    SkipAuth: true,  // 由Handler内部控制
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *UserInfoRequest, resp *UserInfoResponse) error {
        // 手动检查认证状态
        if !ctx.IsAuthenticated() {
            return mod.Reply(401, "需要身份认证")
        }
        // 已认证用户的处理逻辑
        return nil
    }),
})

// 可选认证的接口（个性化功能）
app.Register(mod.Service{
    Name:     "get_articles",
    SkipAuth: true,
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *ArticlesRequest, resp *ArticlesResponse) error {
        if ctx.IsAuthenticated() {
            // 已登录用户看到个性化内容
            resp.Articles = getPersonalizedArticles(ctx.GetUserID())
        } else {
            // 未登录用户看到通用内容
            resp.Articles = getPublicArticles()
        }
        return nil
    }),
})
```

**优势：**
- ✅ 最大灵活性，每个接口精确控制认证逻辑
- ✅ 支持可选认证场景（个性化功能）
- ✅ 代码逻辑清晰，容易调试

**🔐 策略二：严格认证**

适用于大部分接口都需要认证的应用

```go
app.UseJWT()  // 全局强制认证

// 特殊跳过认证的接口
app.Register(mod.Service{
    Name:     "login",
    SkipAuth: true,  // 必须跳过，否则无法登录
    Handler:  mod.MakeHandler(handleLogin),
})

// 自动认证的接口（默认）
app.Register(mod.Service{
    Name:    "user_info",
    // 不设置SkipAuth，框架自动要求JWT认证
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *UserInfoRequest, resp *UserInfoResponse) error {
        // 执行到这里时用户已通过JWT认证
        userID := ctx.GetUserID()    // 保证有值
        username := ctx.GetUsername() // 保证有值
        return nil
    }),
})
```

**优势：**
- ✅ 安全性高，默认所有接口都需要认证
- ✅ 代码更简洁，减少重复的认证检查

##### 配置示例

```yaml
# mod.yml
token:
  jwt:
    enabled: true
    secret_key: "your-super-secret-jwt-key"
    issuer: "your-app-name"
    algorithm: "HS256"
    expire_duration: "24h"
    refresh_expire_duration: "168h"

  validation:
    enabled: true
    cache_strategy: "bigcache"
    cache_key_prefix: "jwt:"
```

##### 上下文方法

JWT中间件会自动解析令牌并将信息注入到上下文中：

```go
// 检查认证状态
if ctx.IsAuthenticated() {
    // 用户已认证
}

// 获取用户信息
userID := ctx.GetUserID()          // 用户ID
username := ctx.GetUsername()      // 用户名
email := ctx.GetUserEmail()        // 邮箱
role := ctx.GetUserRole()          // 角色

// 获取JWT相关信息
token := ctx.GetJWTToken()         // 原始JWT令牌
claims := ctx.GetJWTClaims()       // JWT声明对象
```

---

#### 加解密中间件

加解密中间件提供服务级别的数据加解密功能，支持多种加密算法和灵活的配置策略。

##### 启用方法

```go
app.UseEncryption()  // 启用全局加解密中间件
```

##### 工作原理

1. **请求处理**：自动解密客户端发送的加密数据
2. **签名验证**：验证请求数据的HMAC-SHA256签名
3. **响应加密**：将服务响应数据自动加密后返回

##### 多级配置

加解密中间件支持三级配置，优先级从高到低：

```yaml
# mod.yml
encryption:
  # 全局级别
  global:
    enabled: true
    algorithm: "AES256-GCM"
    mode: "symmetric"

  # 分组级别
  groups:
    "用户管理":
      enabled: true
      algorithm: "AES256-GCM"

  # 服务级别（优先级最高）
  services:
    "create-user":
      enabled: true
      algorithm: "AES256-GCM"

  # 白名单（跳过加解密）
  whitelist:
    groups:
      - "公开服务"
    services:
      - "login"
      - "register"
```

##### 使用示例

```go
app.UseEncryption()

// 启用加解密的服务
app.Register(mod.Service{
    Name:        "create_user",
    DisplayName: "创建用户",
    Description: "包含敏感信息，需要加密传输",
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *CreateUserRequest, resp *CreateUserResponse) error {
        // 请求数据已自动解密
        // 响应数据将自动加密
        return nil
    }),
    Group: "用户管理",
})

// 白名单服务（跳过加解密）
app.Register(mod.Service{
    Name:        "login",
    DisplayName: "用户登录",
    Description: "公开接口，无需加密",
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *LoginRequest, resp *LoginResponse) error {
        // 普通的JSON请求/响应
        return nil
    }),
    Group: "公开服务",
})
```

##### 支持的算法

| 算法 | 模式 | 安全性 | 性能 |
|------|------|--------|------|
| **AES256-GCM** | 对称加密 | 高 | 快 |
| **ChaCha20-Poly1305** | 对称加密 | 高 | 快 |
| **RSA-OAEP** | 非对称加密 | 很高 | 慢 |

##### 配置示例

```yaml
encryption:
  global:
    enabled: true
    algorithm: "AES256-GCM"
    mode: "symmetric"

  # 对称加密配置
  symmetric:
    algorithm: "AES256-GCM"
    key: "dGhpcy1pcy1hLXN1cGVyLXNlY3JldC1rZXktZm9yLWVuY3J5cHRpb24="

  # 签名验证配置
  signature:
    enabled: true
    algorithm: "HMAC-SHA256"
    key: "dGhpcy1pcy1hLXNpZ25hdHVyZS1rZXktZm9yLXZlcmlmaWNhdGlvbg=="
```

### 服务权限系统

MOD提供了基于Token缓存数据的灵活权限控制系统，支持细粒度的权限管理。

#### 权限配置

在服务注册时通过 `Permission` 字段配置权限规则：

```go
app.Register(mod.Service{
    Name:        "admin_data",
    DisplayName: "管理员数据",
    Handler:     mod.MakeHandler(handleAdminData),
    Permission: &mod.PermissionConfig{
        Rules: []mod.PermissionRule{
            {Field: "user.role", Operator: "eq", Value: "admin"},
        },
        Logic: "AND",
    },
})
```

#### 权限规则

**PermissionRule 结构**：
- `Field`: Token缓存数据中的字段路径，支持嵌套访问如 `"user.role"`, `"permissions.admin"`
- `Operator`: 操作符，支持 `eq`、`ne`、`in`、`not_in`、`gt`、`gte`、`lt`、`lte`、`contains`、`exists`
- `Value`: 期望值

**Logic 类型**：
- `"AND"`: 所有规则都必须满足（默认）
- `"OR"`: 任一规则满足即可

#### 使用示例

```go
// 管理员专用服务
app.Register(mod.Service{
    Name: "admin_users",
    Permission: &mod.PermissionConfig{
        Rules: []mod.PermissionRule{
            {Field: "user.role", Operator: "eq", Value: "admin"},
        },
    },
})

// VIP服务（需要VIP等级2以上）
app.Register(mod.Service{
    Name: "vip_service",
    Permission: &mod.PermissionConfig{
        Rules: []mod.PermissionRule{
            {Field: "user.vip_level", Operator: "gte", Value: 2},
            {Field: "user.status", Operator: "eq", Value: "active"},
        },
        Logic: "AND",
    },
})

// 多角色服务
app.Register(mod.Service{
    Name: "manager_data",
    Permission: &mod.PermissionConfig{
        Rules: []mod.PermissionRule{
            {Field: "user.role", Operator: "in", Value: []string{"admin", "manager"}},
        },
    },
})
```

#### Token缓存数据结构

登录时在Token缓存中存储权限相关数据：

```go
tokenData := map[string]interface{}{
    "user": map[string]interface{}{
        "id":        "123",
        "role":      "admin",
        "vip_level": 3,
        "status":    "active",
    },
    "permissions": map[string]interface{}{
        "user_management": true,
        "financial_data":  false,
    },
    "department": map[string]interface{}{
        "name":  "技术部",
        "level": 4,
    },
}

app.SetToken(accessToken, tokenData)
```

#### 权限检查流程

1. 服务请求时自动检查是否配置了 `Permission`
2. 如果配置了权限规则，从Token缓存获取用户数据
3. 根据规则逐一验证字段值
4. 按照 `Logic` 类型（AND/OR）综合判断
5. 权限不足时返回403错误

**优势**：
- **灵活性**：支持复杂的权限规则组合
- **实时性**：基于Token缓存，支持动态权限更新
- **无状态**：不依赖数据库查询
- **服务化**：完全集成到服务注册流程

### 上下文增强

提供强大的上下文功能：

```go
func handler(ctx *mod.Context, req *Request, resp *Response) error {
    // 获取用户信息
    userID := ctx.GetUserID()
    claims := ctx.GetJWTClaims()

    // 检查权限
    if !ctx.HasRole("admin") {
        return mod.Reply(403, "权限不足")
    }

    // 结构化日志
    ctx.WithFields(map[string]interface{}{
        "user_id": userID,
        "action":  "update_user",
    }).Info("用户更新操作")

    // 获取应用实例
    config := ctx.App().GetModConfig()

    return nil
}
```

---

## 🔧 功能特性

### JWT认证系统

完整的JWT认证系统，使用 `github.com/golang-jwt/jwt/v5` 库：

#### 核心功能
- ✅ Token生成和验证
- 🔐 角色权限控制
- 🔄 Token刷新机制
- ❌ Token撤销和黑名单
- 💾 多种存储后端支持（BigCache、BadgerDB、Redis）

#### 快速开始

**1. 配置JWT设置**
```yaml
# mod.yml
token:
  jwt:
    enabled: true
    secret_key: "your-super-secret-jwt-key"
    issuer: "your-app-name"
    algorithm: "HS256"
    expire_duration: "24h"
    refresh_expire_duration: "168h"  # 7天

  validation:
    enabled: true
    cache_strategy: "bigcache"
    cache_key_prefix: "jwt:"
```

**2. 启用JWT中间件**
```go
func main() {
    app := mod.New()

    // 选择认证模式（推荐可选模式）
    app.UseOptionalJWT()

    // 注册服务...
    app.Run(":8080")
}
```

#### 完整用法示例

```go
// 用户登录 - 生成JWT
app.Register(mod.Service{
    Name:     "login",
    SkipAuth: true,  // 登录接口跳过认证
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *LoginRequest, resp *LoginResponse) error {
        // 验证用户名密码...
        user := validateUser(req.Username, req.Password)

        // 生成JWT令牌
        tokens, err := app.GenerateJWT(
            user.ID,
            user.Username,
            user.Email,
            user.Role,
            map[string]interface{}{  // 自定义声明
                "login_time": time.Now().Unix(),
                "permissions": []string{"read", "write"},
            },
        )
        if err != nil {
            return mod.Reply(500, "生成令牌失败")
        }

        // 可选：存储令牌到缓存用于权限控制
        tokenData := map[string]interface{}{
            "user_id": user.ID,
            "role":    user.Role,
            "status":  "active",
        }
        app.SetToken(tokens.AccessToken, tokenData)

        resp.User = user
        resp.Token = tokens
        return nil
    }),
})

// 需要认证的接口
app.Register(mod.Service{
    Name:     "user_info",
    SkipAuth: true,  // 使用内部认证控制
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *UserInfoRequest, resp *UserInfoResponse) error {
        // 检查是否已认证
        if !ctx.IsAuthenticated() {
            return mod.Reply(401, "需要身份认证")
        }

        // 获取用户信息
        userID := ctx.GetUserID()
        username := ctx.GetUsername()
        role := ctx.GetUserRole()

        resp.User = User{
            ID:       userID,
            Username: username,
            Role:     role,
        }
        return nil
    }),
})

// 令牌刷新
app.Register(mod.Service{
    Name:     "refresh",
    SkipAuth: true,
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *RefreshRequest, resp *mod.TokenResponse) error {
        // 刷新令牌
        tokens, err := app.RefreshJWT(req.RefreshToken)
        if err != nil {
            return mod.Reply(401, "刷新令牌无效")
        }

        *resp = *tokens
        return nil
    }),
})

// 用户登出
app.Register(mod.Service{
    Name:     "logout",
    SkipAuth: true,
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *LogoutRequest, resp *LogoutResponse) error {
        token := ctx.GetJWTToken()
        if token == "" {
            return mod.Reply(401, "未提供令牌")
        }

        // 撤销令牌（加入黑名单）
        if err := app.RevokeJWT(token); err != nil {
            return mod.Reply(500, "登出失败")
        }

        // 从缓存移除令牌
        app.RemoveToken(token)

        resp.Message = "登出成功"
        return nil
    }),
})
```

#### 上下文方法

JWT中间件会自动解析令牌并将信息注入到上下文中：

```go
// 检查认证状态
if ctx.IsAuthenticated() {
    // 用户已认证
}

// 获取用户信息
userID := ctx.GetUserID()          // 用户ID
username := ctx.GetUsername()      // 用户名
email := ctx.GetUserEmail()        // 邮箱
role := ctx.GetUserRole()          // 角色

// 获取JWT相关信息
token := ctx.GetJWTToken()         // 原始JWT令牌
claims := ctx.GetJWTClaims()       // JWT声明对象

// 获取自定义声明
if claims != nil {
    loginTime := claims.ExtraClaims["login_time"]
}
```

#### 令牌格式支持

MOD支持两种Authorization头格式：

```bash
# Bearer格式（标准）
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# 直接格式
Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 错误处理

```go
// 令牌验证失败时的错误响应
{
  "code": 401,
  "message": "Invalid authentication token",
  "data": null,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 服务加解密

多级别的服务加解密系统，保护敏感数据传输：

#### 支持的加密算法
- **对称加密**: AES256-GCM, ChaCha20-Poly1305
- **非对称加密**: RSA-OAEP
- **数字签名**: HMAC-SHA256

#### 配置级别
- **全局级别**: 所有服务默认加密
- **分组级别**: 特定分组的服务加密
- **服务级别**: 特定服务的加密配置
- **白名单**: 跳过加密的服务和分组

#### 配置示例

```yaml
encryption:
  # 全局配置
  global:
    enabled: true
    algorithm: "AES256-GCM"
    mode: "symmetric"

  # 对称加密配置
  symmetric:
    algorithm: "AES256-GCM"
    key: "base64-encoded-key"
    key_file: "/path/to/key/file"

  # 非对称加密配置
  asymmetric:
    algorithm: "RSA-OAEP"
    public_key: "-----BEGIN PUBLIC KEY-----..."
    private_key: "-----BEGIN PRIVATE KEY-----..."
    key_size: 2048

  # 签名验证配置
  signature:
    enabled: true
    algorithm: "HMAC-SHA256"
    key: "signature-key"

  # 分组级别配置
  groups:
    "敏感数据":
      enabled: true

  # 服务级别配置
  services:
    "get_user_detail":
      enabled: true

  # 白名单配置
  whitelist:
    groups: ["公开数据"]
    services: ["health_check"]
```

#### 使用方式

```go
// 启用加解密中间件
app.UseEncryption()

// 手动加解密
encrypted, err := app.EncryptData(data, "symmetric")
decrypted, err := app.DecryptData(encrypted, "symmetric")

// 数字签名
signature, err := app.SignData(data)
err = app.VerifySignature(data, signature)
```

### 文件服务

#### 文件上传

支持多种存储后端的文件上传：

```yaml
file_upload:
  # 本地存储
  local:
    enabled: true
    upload_dir: "./uploads"
    max_size: "50MB"
    allowed_types: ["image/jpeg", "image/png", "application/pdf"]
    allowed_exts: [".jpg", ".png", ".pdf"]
    keep_original_name: false
    auto_create_dir: true
    date_sub_dir: true

  # AWS S3
  s3:
    enabled: true
    bucket: "my-bucket"
    region: "us-east-1"
    access_key: "your-access-key"
    secret_key: "your-secret-key"

  # 阿里云OSS
  oss:
    enabled: true
    bucket: "my-oss-bucket"
    endpoint: "oss-cn-shenzhen.aliyuncs.com"
    access_key_id: "your-access-key-id"
    access_key_secret: "your-access-key-secret"
```

#### 静态文件

高性能静态文件服务：

```yaml
static_mounts:
  - url_prefix: "/static"
    local_path: "./static"
    browseable: true
    index_file: "index.html"

  - url_prefix: "/docs"
    local_path: "./docs"
    browseable: false
    index_file: "README.html"
```

### 日志系统

#### 多后端日志支持

```yaml
logging:
  # 控制台日志
  console:
    enabled: true
    level: "info"

  # 文件日志（支持轮转）
  file:
    enabled: true
    path: "./logs/app.log"
    max_size: "100MB"
    max_backups: 10
    max_age: "30d"
    compress: true

  # Grafana Loki
  loki:
    enabled: true
    url: "http://127.0.0.1:3100/loki/api/v1/push"
    labels:
      service: "mod-app"
      environment: "production"
    batch_size: 100
    timeout: "10s"

  # 阿里云SLS
  sls:
    enabled: true
    endpoint: "cn-shenzhen.log.aliyuncs.com"
    project: "my-project"
    logstore: "my-logstore"
    access_key_id: "your-access-key-id"
    access_key_secret: "your-access-key-secret"
```

#### 结构化日志

```go
// 基础日志
ctx.Info("用户登录成功")

// 结构化日志
ctx.WithFields(map[string]interface{}{
    "user_id": "123",
    "action":  "login",
    "ip":      "192.168.1.1",
}).Info("用户登录成功")

// 获取Logger实例
logger := ctx.Logger()
logger.WithField("key", "value").Warn("警告信息")
```

### Mock功能

智能Mock数据生成，支持多级别配置：

```yaml
mock:
  # 全局Mock
  global:
    enabled: false

  # 分组Mock
  groups:
    "用户管理":
      enabled: true

  # 服务Mock
  services:
    "get_user":
      enabled: true
```

Mock功能会根据响应结构自动生成合理的测试数据，支持开发和测试阶段快速原型开发。

### 缓存系统

用于JWT Token验证的多种缓存方案：

```yaml
cache:
  # BigCache（内存缓存）
  bigcache:
    enabled: true
    shards: 1024
    life_window: "24h"
    clean_window: "1h"
    max_entries_in_window: 10000
    max_entry_size: 1024

  # BadgerDB（持久化缓存）
  badger:
    enabled: false
    path: "./data/tokens"
    in_memory: false
    sync_writes: false
    ttl: "24h"

  # Redis
  redis:
    enabled: false
    address: "127.0.0.1:6379"
    password: ""
    db: 0
    pool_size: 10
    min_idle_conns: 0
    ttl: "24h"
```

---

## ⚙️ 配置系统

MOD使用YAML配置文件 `mod.yml` 进行统一配置管理。配置文件支持环境变量替换和热重载。

### 完整配置示例

```yaml
# 应用配置
app:
  name: "MyApp"
  display_name: "我的应用"
  description: "应用描述"
  version: "1.0.0"
  service_base: "/services"
  token_keys: ["Authorization", "X-API-Key", "mod-token"]

# 服务器配置
server:
  host: "127.0.0.1"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  body_limit: "100MB"
  concurrency: 256

  # CORS配置
  cors:
    enabled: true
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["Origin", "Content-Type", "Accept", "Authorization"]
    allow_credentials: false
    max_age: "24h"
```

---

## 📚 完整示例

MOD提供了丰富的示例，涵盖所有核心功能：

```
examples/
├── basic-services/     # 基础服务注册和参数验证
├── jwt-auth/          # JWT认证和权限控制
├── encryption/        # 服务加解密和签名验证
├── file-upload/       # 多后端文件上传
├── static-files/      # 静态文件服务
├── logging/           # 多种日志输出方式
└── mock/              # 服务Mock功能
```

每个示例都可以独立运行：

```bash
cd examples/basic-services
go run main.go
```

---

## 📖 配置参考

### 应用配置 (app)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `name` | string | 应用名称 | "MOD" |
| `display_name` | string | 应用显示名称 | "MOD Application" |
| `description` | string | 应用描述 | "" |
| `version` | string | 应用版本 | "" |
| `service_base` | string | 服务基础路径 | "/services" |
| `token_keys` | []string | Token请求头名称 | ["Authorization", "X-API-Key", "mod-token"] |

### 服务器配置 (server)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `host` | string | 监听主机 | "" |
| `port` | int | 监听端口 | 8080 |
| `read_timeout` | string | 读取超时 | "30s" |
| `write_timeout` | string | 写入超时 | "30s" |
| `idle_timeout` | string | 空闲超时 | "120s" |
| `body_limit` | string | 请求体大小限制 | "100MB" |
| `concurrency` | int | 并发连接数 | 256 |

### CORS配置 (server.cors)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用CORS | false |
| `allow_origins` | []string | 允许的源 | ["*"] |
| `allow_methods` | []string | 允许的HTTP方法 | ["GET", "POST", "PUT", "DELETE", "OPTIONS"] |
| `allow_headers` | []string | 允许的请求头 | ["Origin", "Content-Type", "Accept", "Authorization"] |
| `allow_credentials` | bool | 是否允许携带凭证 | false |
| `max_age` | string | 预检请求缓存时间 | "24h" |

### JWT配置 (jwt)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用JWT | false |
| `secret_key` | string | JWT签名密钥 | "" |
| `issuer` | string | JWT签发者 | "" |
| `expire_duration` | string | Access Token过期时间 | "24h" |
| `refresh_expire_duration` | string | Refresh Token过期时间 | "168h" |
| `algorithm` | string | 签名算法 | "HS256" |

### 加解密配置 (encryption)

#### 全局配置 (encryption.global)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用全局加解密 | false |
| `algorithm` | string | 加密算法 | "AES256-GCM" |
| `mode` | string | 加密模式 (symmetric/asymmetric) | "symmetric" |

#### 对称加密配置 (encryption.symmetric)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `algorithm` | string | 对称加密算法 | "AES256-GCM" |
| `key` | string | 加密密钥 (base64编码) | "" |
| `key_file` | string | 密钥文件路径 | "" |

#### 非对称加密配置 (encryption.asymmetric)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `algorithm` | string | 非对称加密算法 | "RSA-OAEP" |
| `public_key` | string | 公钥内容 (PEM格式) | "" |
| `private_key` | string | 私钥内容 (PEM格式) | "" |
| `public_key_file` | string | 公钥文件路径 | "" |
| `private_key_file` | string | 私钥文件路径 | "" |
| `key_size` | int | RSA密钥长度 | 2048 |

#### 签名验证配置 (encryption.signature)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用签名验证 | false |
| `algorithm` | string | 签名算法 | "HMAC-SHA256" |
| `key` | string | 签名密钥 | "" |
| `key_file` | string | 签名密钥文件路径 | "" |

### 日志配置 (logging)

#### 控制台日志 (logging.console)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用控制台日志 | true |
| `level` | string | 日志级别 (debug/info/warn/error) | "info" |

#### 文件日志 (logging.file)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用文件日志 | false |
| `path` | string | 日志文件路径 | "" |
| `max_size` | string | 单个日志文件最大大小 | "100MB" |
| `max_backups` | int | 保留的历史日志文件数量 | 3 |
| `max_age` | string | 日志文件保留时间 | "30d" |
| `compress` | bool | 是否压缩历史日志文件 | false |

### 文件上传配置 (file_upload)

#### 本地上传配置 (file_upload.local)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用本地文件上传 | false |
| `upload_dir` | string | 上传目录路径 | "./uploads" |
| `max_size` | string | 单文件最大大小 | "10MB" |
| `allowed_types` | []string | 允许的文件MIME类型 | [] |
| `allowed_exts` | []string | 允许的文件扩展名 | [] |
| `keep_original_name` | bool | 是否保持原始文件名 | false |
| `auto_create_dir` | bool | 自动创建上传目录 | true |
| `date_sub_dir` | bool | 按日期创建子目录 | false |

### 缓存配置 (cache)

#### BigCache配置 (cache.bigcache)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用BigCache | false |
| `shards` | int | 分片数量 | 1024 |
| `life_window` | string | 生命周期窗口 | "24h" |
| `clean_window` | string | 清理窗口 | "1h" |
| `max_entries_in_window` | int | 窗口内最大条目数 | 10000 |
| `max_entry_size` | int | 最大条目大小 | 1024 |

---

## 🆘 获取帮助

- 📚 **API文档**: 运行任意示例后访问 http://127.0.0.1:8080/services/docs
- 💬 **问题反馈**: [GitHub Issues](https://github.com/iamdanielyin/mod/issues) - 报告bug、提出建议或寻求帮助

## 📄 许可证

本项目采用 [Apache 2.0](LICENSE) 许可证。

## 📖 文档说明

**这是MOD的唯一完整文档**，包含了所有功能特性的详细说明和配置参考。

---

**MOD** - 让Go Web开发更简单、更安全、更高效！