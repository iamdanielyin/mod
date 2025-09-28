# MOD

一个基于Go Fiber的现代化企业级Web应用框架，专注于快速开发、安全性和可扩展性。

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Fiber Version](https://img.shields.io/badge/Fiber-v2.x-green.svg)](https://gofiber.io)
[![License](https://img.shields.io/badge/License-Apache2.0-green.svg)](LICENSE)

## ✨ 核心特性

### 🚀 开发效率
- **服务化架构**: 基于服务注册的模块化开发模式
- **自动API文档**: 内置API文档生成和Web界面
- **参数验证**: 集成go-playground/validator，支持复杂验证规则
- **统一响应**: 标准化的响应格式和错误处理

### 🔒 安全特性
- **JWT认证**: 完整的JWT认证系统，支持角色权限控制
- **服务加解密**: 多级别的加解密配置，保护敏感数据传输
- **数字签名**: HMAC/RSA签名验证，确保数据完整性
- **Token管理**: 支持Token黑名单和多种存储后端

### 🛠 企业功能
- **多日志后端**: 控制台、文件、Loki、阿里云SLS
- **文件上传**: 本地、S3、阿里云OSS多后端支持
- **静态文件**: 高性能静态文件服务和目录浏览
- **缓存系统**: BigCache、BadgerDB、Redis多种缓存方案

### 🔧 开发工具
- **Mock功能**: 智能Mock数据生成，支持多级别配置
- **热重载**: 开发环境友好的配置热加载
- **CORS支持**: 灵活的跨域配置
- **中间件**: 丰富的内置中间件和自定义扩展

## 🚀 快速开始

### 安装

```bash
go get github.com/iamdanielyin/mod
```

### 基础使用

```go
package main

import "github.com/iamdanielyin/mod"

// 定义请求和响应结构
type GetUserRequest struct {
    ID string `json:"id" validate:"required" desc:"用户ID"`
}

type GetUserResponse struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    app := mod.New()

    // 注册服务
    app.Register(mod.Service{
        Name:        "get-user",
        DisplayName: "获取用户信息",
        Description: "根据用户ID获取用户详细信息",
        Handler:     mod.MakeHandler(handleGetUser),
        Group:       "用户管理",
    })

    app.Run(":8080")
}

// 服务处理函数
func handleGetUser(ctx *mod.Context, req *GetUserRequest, resp *GetUserResponse) error {
    resp.Name = "张三"
    resp.Email = "zhangsan@example.com"
    return nil
}
```

访问 http://localhost:8080/services/docs 查看自动生成的API文档。

## 📚 完整示例

MOD框架提供了丰富的示例，涵盖所有核心功能：

```bash
cd examples/
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

## 🔧 配置系统

MOD框架使用YAML配置文件 `mod.yml` 进行统一配置管理：

```yaml
app:
  name: "MyApp"
  display_name: "我的应用"
  description: "应用描述"
  version: "1.0.0"

server:
  host: "localhost"
  port: 8080
  cors:
    enabled: true
    allow_origins: ["*"]

logging:
  console:
    enabled: true
    level: "info"
  file:
    enabled: true
    path: "./logs/app.log"

jwt:
  enabled: true
  secret_key: "your-secret-key"
  expire_duration: "24h"

encryption:
  global:
    enabled: true
    algorithm: "AES256-GCM"
    mode: "symmetric"
```

## 🏗 架构特点

### 服务化设计
MOD框架采用服务化架构，每个业务功能都注册为独立的服务：

```go
app.Register(mod.Service{
    Name:        "service-name",        // 服务名称
    DisplayName: "服务显示名",            // 显示名称
    Description: "服务描述",              // 服务描述
    Handler:     mod.MakeHandler(fn),   // 处理函数
    Group:       "服务分组",              // 服务分组
    Sort:        1,                     // 排序
    SkipAuth:    false,                 // 是否跳过认证
    ReturnRaw:   false,                 // 是否返回原始数据
})
```

### 中间件系统
支持灵活的中间件配置：

```go
// JWT认证中间件
app.UseJWT()

// 可选JWT中间件
app.UseOptionalJWT()

// 角色权限中间件
app.Use(mod.RoleMiddleware("admin"))

// 加解密中间件
app.UseEncryption()
```

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

    return nil
}
```

## 🔐 安全特性

### JWT认证
完整的JWT认证系统：

```go
// 生成Token
tokenResp, err := app.GenerateJWT("user123", "张三", "zhangsan@example.com", "admin", nil)

// 验证Token
claims, err := app.ValidateJWT(tokenString)

// 刷新Token
newTokenResp, err := app.RefreshJWT(refreshToken)

// 撤销Token
err = app.RevokeJWT(tokenString)
```

### 服务加解密
支持多级别的加解密配置：

```yaml
encryption:
  global:
    enabled: true                    # 全局启用
    algorithm: "AES256-GCM"         # 加密算法
    mode: "symmetric"               # 加密模式

  services:
    "sensitive-service":            # 特定服务配置
      enabled: true

  whitelist:
    services:
      - "public-service"            # 白名单服务
```

### 数字签名
确保数据完整性：

```go
// 创建签名
signature, err := app.SignData(data)

// 验证签名
err = app.VerifySignature(data, signature)
```

## 📁 文件服务

### 文件上传
支持多种存储后端：

```yaml
file_upload:
  local:
    enabled: true
    upload_dir: "./uploads"
    max_size: "50MB"
    allowed_types: ["image/jpeg", "image/png"]

  s3:
    enabled: true
    bucket: "my-bucket"
    region: "us-east-1"
    access_key: "your-access-key"
    secret_key: "your-secret-key"

  oss:
    enabled: true
    bucket: "my-oss-bucket"
    endpoint: "oss-cn-shenzhen.aliyuncs.com"
    access_key_id: "your-access-key-id"
    access_key_secret: "your-access-key-secret"
```

### 静态文件服务
灵活的静态文件挂载：

```yaml
static_mounts:
  - url_prefix: "/static"
    local_path: "./static"
    browseable: true
    index_file: "index.html"

  - url_prefix: "/docs"
    local_path: "./docs"
    browseable: false
```

## 📊 日志系统

### 多后端日志
支持多种日志输出方式：

```yaml
logging:
  console:
    enabled: true
    level: "info"

  file:
    enabled: true
    path: "./logs/app.log"
    max_size: "100MB"
    max_backups: 10
    max_age: "30d"
    compress: true

  loki:
    enabled: true
    url: "http://localhost:3100/loki/api/v1/push"
    labels:
      service: "my-app"
      environment: "production"

  sls:
    enabled: true
    endpoint: "cn-shenzhen.log.aliyuncs.com"
    project: "my-project"
    logstore: "my-logstore"
```

### 结构化日志
支持结构化日志记录：

```go
ctx.WithFields(map[string]interface{}{
    "user_id": "123",
    "action":  "login",
    "ip":      "192.168.1.1",
}).Info("用户登录成功")
```

## 💾 缓存系统

支持多种缓存后端用于Token验证：

```yaml
cache:
  bigcache:
    enabled: true
    shards: 1024
    life_window: "24h"
    clean_window: "1h"

  badger:
    enabled: true
    path: "./data/tokens"
    ttl: "24h"

  redis:
    enabled: true
    address: "localhost:6379"
    password: ""
    db: 0
    ttl: "24h"
```

## 🧪 开发工具

### Mock功能
智能Mock数据生成：

```yaml
mock:
  global:
    enabled: true                   # 全局Mock

  services:
    "user-service":                # 特定服务Mock
      enabled: true
```

### API文档
自动生成的交互式API文档：
- 访问 `/services/docs` 查看完整API文档
- 支持参数说明、类型信息、示例数据
- 提供在线测试功能

## 📋 完整配置参考

### 应用配置 (app)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `name` | string | 应用名称 | "MOD" |
| `display_name` | string | 应用显示名称 | "MOD Application" |
| `description` | string | 应用描述 | "" |
| `version` | string | 应用版本 | "" |
| `service_path_prefix` | string | 服务路径前缀 | "/services" |
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

#### CORS配置 (server.cors)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用CORS | false |
| `allow_origins` | []string | 允许的源 | ["*"] |
| `allow_methods` | []string | 允许的HTTP方法 | ["GET", "POST", "PUT", "DELETE", "OPTIONS"] |
| `allow_headers` | []string | 允许的请求头 | ["Origin", "Content-Type", "Accept", "Authorization"] |
| `allow_credentials` | bool | 是否允许携带凭证 | false |
| `max_age` | string | 预检请求缓存时间 | "24h" |

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

#### Loki日志 (logging.loki)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用Loki日志 | false |
| `url` | string | Loki推送URL | "" |
| `labels` | map[string]string | 日志标签 | {} |
| `batch_size` | int | 批量发送大小 | 100 |
| `timeout` | string | 发送超时时间 | "10s" |

#### 阿里云SLS日志 (logging.sls)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用SLS日志 | false |
| `endpoint` | string | SLS服务端点 | "" |
| `project` | string | SLS项目名 | "" |
| `logstore` | string | SLS日志库名 | "" |
| `access_key_id` | string | 访问密钥ID | "" |
| `access_key_secret` | string | 访问密钥Secret | "" |

### JWT配置 (jwt)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用JWT | false |
| `secret_key` | string | JWT签名密钥 | "" |
| `issuer` | string | JWT签发者 | "" |
| `expire_duration` | string | Access Token过期时间 | "24h" |
| `refresh_expire_duration` | string | Refresh Token过期时间 | "168h" |
| `algorithm` | string | 签名算法 | "HS256" |

#### Token验证配置 (token.validation)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用Token验证 | false |
| `skip_expired_check` | bool | 是否跳过过期检查 | false |
| `cache_strategy` | string | 缓存策略 (bigcache/badger/redis) | "" |
| `cache_key_prefix` | string | 缓存键前缀 | "token:" |

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
| `verbose` | bool | 是否启用详细日志 | false |
| `hard_max_cache_size` | int | 硬性最大缓存大小 | 0 |

#### BadgerDB配置 (cache.badger)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用BadgerDB | false |
| `path` | string | 数据库路径 | "./data/tokens" |
| `in_memory` | bool | 是否使用内存模式 | false |
| `sync_writes` | bool | 是否同步写入 | false |
| `ttl` | string | 数据过期时间 | "24h" |

#### Redis配置 (cache.redis)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用Redis | false |
| `address` | string | Redis地址 | "localhost:6379" |
| `password` | string | Redis密码 | "" |
| `db` | int | Redis数据库 | 0 |
| `pool_size` | int | 连接池大小 | 10 |
| `min_idle_conns` | int | 最小空闲连接数 | 0 |
| `ttl` | string | 数据过期时间 | "24h" |

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

#### 分组级别配置 (encryption.groups)

每个分组可以有独立的加解密配置：

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用该分组的加解密 | false |
| `algorithm` | string | 覆盖全局算法设置 | "" |
| `mode` | string | 覆盖全局模式设置 | "" |

#### 服务级别配置 (encryption.services)

每个服务可以有独立的加解密配置：

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用该服务的加解密 | false |
| `algorithm` | string | 覆盖全局算法设置 | "" |
| `mode` | string | 覆盖全局模式设置 | "" |

#### 白名单配置 (encryption.whitelist)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `groups` | []string | 白名单分组列表 | [] |
| `services` | []string | 白名单服务列表 | [] |

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

#### S3上传配置 (file_upload.s3)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用S3上传 | false |
| `bucket` | string | S3存储桶名称 | "" |
| `region` | string | S3区域 | "" |
| `access_key` | string | 访问密钥 | "" |
| `secret_key` | string | 密钥 | "" |
| `endpoint` | string | 自定义端点 (可选) | "" |

#### 阿里云OSS配置 (file_upload.oss)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用OSS上传 | false |
| `bucket` | string | OSS存储桶名称 | "" |
| `endpoint` | string | OSS端点 | "" |
| `access_key_id` | string | 访问密钥ID | "" |
| `access_key_secret` | string | 访问密钥Secret | "" |

### 静态文件配置 (static_mounts)

静态文件挂载配置是一个数组，每个挂载点包含：

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `url_prefix` | string | URL前缀 | "" |
| `local_path` | string | 本地路径 | "" |
| `browseable` | bool | 是否允许目录浏览 | false |
| `index_file` | string | 默认索引文件 | "index.html" |

### Mock配置 (mock)

#### 全局Mock配置 (mock.global)

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用全局Mock | false |

#### 分组Mock配置 (mock.groups)

每个分组可以有独立的Mock配置：

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用该分组的Mock | false |

#### 服务Mock配置 (mock.services)

每个服务可以有独立的Mock配置：

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `enabled` | bool | 是否启用该服务的Mock | false |

## 🤝 贡献指南

欢迎提交Issue和Pull Request来帮助改进MOD框架！

### 开发环境设置

```bash
git clone https://github.com/iamdanielyin/mod.git
cd mod
go mod tidy
```

### 运行测试

```bash
go test ./...
```

### 运行示例

```bash
cd examples/basic-services
go run main.go
```

## 📄 许可证

本项目采用 [Apache 2.0](LICENSE) 许可证。

## 📖 文档说明

**这是MOD的唯一完整文档**，包含了所有功能特性的详细说明和配置参考。

## 🆘 获取帮助

- 📚 **API文档**: 运行任意示例后访问 http://localhost:8080/services/docs
- 💬 **问题反馈**: [GitHub Issues](https://github.com/iamdanielyin/mod/issues) - 报告bug、提出建议或寻求帮助

---

**MOD** - 让Go Web开发更简单、更安全、更高效！