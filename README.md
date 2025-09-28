# MOD Framework

一个基于Go语言和Fiber的高性能微服务框架，提供完整的API开发、文件管理、日志记录和配置管理解决方案。

## 🚀 核心特性

### ✅ 已实现功能

- **🔧 参数解析与验证**: 基于结构体标签的自动参数解析和验证
- **📚 API文档生成**: 自动生成交互式API文档
- **🔐 JWT认证**: 完整的Token签发、验证和刷新机制
- **📁 静态文件服务**: 多路径静态文件挂载和目录浏览
- **📤 文件上传**: 支持本地存储、S3和阿里云OSS的多后端文件上传
- **📋 日志系统**: 基于logrus的结构化日志，支持文件轮转
- **💾 多级缓存**: BigCache内存缓存、BadgerDB本地存储、Redis远程缓存
- **🌐 CORS支持**: 跨域资源共享配置
- **⚙️ 配置管理**: 基于YAML的灵活配置系统
- **🔍 请求跟踪**: 请求ID生成和分布式链路追踪

### 🔄 开发中功能

- **🔒 配置加解密**: RSA密钥对加解密配置
- **🎯 参数加解密**: 敏感参数自动加解密
- **🔄 类型转换**: 参数类型智能转换
- **🎭 接口Mock**: API接口模拟和测试

## 📦 安装

```bash
go get github.com/iamdanielyin/mod
```

## 🚀 快速开始

### 基础示例

```go
package main

import (
    "github.com/iamdanielyin/mod"
)

// 定义请求结构
type LoginRequest struct {
    Username string `json:"username" validate:"required" desc:"用户名"`
    Password string `json:"password" validate:"required,min=6" desc:"密码"`
}

// 定义响应结构
type LoginResponse struct {
    Token string `json:"token" desc:"访问令牌"`
    UID   string `json:"uid" desc:"用户ID"`
}

func main() {
    // 创建应用实例
    app := mod.New()

    // 注册服务
    app.Register(mod.Service{
        Name:        "login",
        DisplayName: "用户登录",
        Description: "用户登录验证",
        SkipAuth:    true,
        Handler: mod.MakeHandler(func(ctx *mod.Context, req *LoginRequest, resp *LoginResponse) error {
            // 业务逻辑处理
            if req.Username == "admin" && req.Password == "123456" {
                resp.Token = "your-jwt-token"
                resp.UID = "user-123"
                return nil
            }
            return mod.ReplyWithDetail(400, "登录失败", "用户名或密码错误")
        }),
    })

    // 启动服务
    app.Run()
}
```

访问 `http://localhost:8080/services/docs` 查看自动生成的API文档。

## 📋 配置文件

创建 `mod.yml` 配置文件：

```yaml
# 应用配置
app:
  name: "my-app"
  display_name: "我的应用"
  description: "这是一个示例应用"
  version: "1.0.0"
  port: 8080

# 文件上传配置
file_upload:
  # 本地存储
  local:
    enabled: true
    upload_dir: "./uploads"
    max_size: "10MB"
    allowed_types: ["image/jpeg", "image/png", "application/pdf"]
    allowed_exts: [".jpg", ".png", ".pdf"]

  # 阿里云OSS
  oss:
    enabled: false
    bucket: "my-bucket"
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    access_key_id: "your-access-key"
    access_key_secret: "your-secret-key"

  # Amazon S3
  s3:
    enabled: false
    bucket: "my-s3-bucket"
    region: "us-west-2"
    access_key: "your-access-key"
    secret_key: "your-secret-key"

# 日志配置
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

# 静态文件挂载
static_mounts:
  - url_prefix: "/static"
    local_path: "./public"
    browseable: false
    index_file: "index.html"
  - url_prefix: "/uploads"
    local_path: "./uploads"
    browseable: true

# 缓存配置
cache:
  bigcache:
    enabled: true
    life_window: "10m"
    clean_window: "5m"
  badger:
    enabled: true
    path: "./data/badger"
  redis:
    enabled: false
    address: "localhost:6379"

# JWT Token配置
token:
  jwt:
    enabled: true
    secret_key: "your-secret-key"
    expire_duration: "24h"
    refresh_expire_duration: "168h"
```

## 🔧 主要功能

### 1. 服务注册

```go
app.Register(mod.Service{
    Name:        "user-info",           // 服务名称
    DisplayName: "获取用户信息",          // 显示名称
    Description: "根据用户ID获取详细信息", // 服务描述
    SkipAuth:    false,                 // 是否跳过认证
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *UserInfoRequest, resp *UserInfoResponse) error {
        // 处理逻辑
        return nil
    }),
})
```

### 2. 文件上传

```go
// 单文件上传
POST /upload
Content-Type: multipart/form-data
参数: file (文件)

// 批量上传
POST /upload/batch
Content-Type: multipart/form-data
参数: files (多个文件)
```

### 3. 静态文件服务

配置文件中设置 `static_mounts` 后，可直接通过URL访问静态文件：

- `http://localhost:8080/static/css/style.css`
- `http://localhost:8080/uploads/image.jpg`

### 4. 日志记录

```go
func handleRequest(ctx *mod.Context, req *Request, resp *Response) error {
    logger := ctx.GetLogger()

    // 结构化日志
    logger.WithFields(map[string]interface{}{
        "request_id": ctx.GetRequestID(),
        "user_id":    req.UserID,
        "action":     "process_request",
    }).Info("处理用户请求")

    return nil
}
```

### 5. 缓存使用

```go
// 设置缓存
err := app.SetCache("key", "value", time.Hour)

// 获取缓存
value, err := app.GetCache("key")

// 删除缓存
err := app.DeleteCache("key")
```

### 6. JWT认证

```go
// 生成Token
token, err := app.GenerateToken(userID, claims)

// 验证Token
claims, err := app.VerifyToken(tokenString)

// 刷新Token
newToken, err := app.RefreshToken(oldToken)
```

## 📁 项目结构

```
your-project/
├── main.go                 # 应用入口
├── mod.yml                 # 配置文件
├── uploads/                # 文件上传目录
├── logs/                   # 日志文件目录
├── public/                 # 静态文件目录
└── data/                   # 数据存储目录
    └── badger/             # BadgerDB数据库
```

## 📚 示例代码

查看 `examples/` 目录获取更多示例：

- `basic_demo.go` - 基础服务示例
- `complex_services_demo.go` - 复杂服务示例
- `token_demo.go` - JWT认证示例
- `upload_example.go` - 文件上传示例
- `static_example.go` - 静态文件服务示例
- `cors_example.go` - CORS跨域示例
- `file-logging/` - 文件日志完整示例
- `s3-upload/` - S3文件上传示例

### 完整应用示例

```go
package main

import (
    "time"
    "github.com/iamdanielyin/mod"
)

// 用户注册请求
type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=20" desc:"用户名，3-20个字符"`
    Email    string `json:"email" validate:"required,email" desc:"邮箱地址"`
    Password string `json:"password" validate:"required,min=6" desc:"密码，至少6位"`
}

// 用户注册响应
type RegisterResponse struct {
    UserID string `json:"user_id" desc:"用户ID"`
    Token  string `json:"token" desc:"访问令牌"`
}

// 获取用户列表请求
type UserListRequest struct {
    Page     int    `json:"page" validate:"min=1" desc:"页码，从1开始"`
    PageSize int    `json:"page_size" validate:"min=1,max=100" desc:"每页数量，1-100"`
    Keyword  string `json:"keyword" desc:"搜索关键词"`
}

// 用户信息
type UserInfo struct {
    ID       string    `json:"id" desc:"用户ID"`
    Username string    `json:"username" desc:"用户名"`
    Email    string    `json:"email" desc:"邮箱"`
    Created  time.Time `json:"created" desc:"创建时间"`
}

// 获取用户列表响应
type UserListResponse struct {
    Users []UserInfo `json:"users" desc:"用户列表"`
    Total int        `json:"total" desc:"总数量"`
    Page  int        `json:"page" desc:"当前页码"`
}

func main() {
    app := mod.New()

    // 用户注册服务
    app.Register(mod.Service{
        Name:        "register",
        DisplayName: "用户注册",
        Description: "创建新用户账户",
        SkipAuth:    true,
        Handler: mod.MakeHandler(func(ctx *mod.Context, req *RegisterRequest, resp *RegisterResponse) error {
            logger := ctx.GetLogger()

            // 记录请求日志
            logger.WithFields(map[string]interface{}{
                "username":   req.Username,
                "email":      req.Email,
                "request_id": ctx.GetRequestID(),
                "ip":         ctx.IP(),
            }).Info("用户注册请求")

            // 检查用户名是否已存在（这里是示例逻辑）
            if req.Username == "admin" {
                return mod.ReplyWithDetail(400, "注册失败", "用户名已存在")
            }

            // 模拟创建用户
            userID := "user_" + ctx.GetRequestID()

            // 生成JWT Token
            token, err := generateUserToken(userID)
            if err != nil {
                logger.WithError(err).Error("生成Token失败")
                return mod.ReplyWithDetail(500, "系统错误", "Token生成失败")
            }

            resp.UserID = userID
            resp.Token = token

            logger.WithField("user_id", userID).Info("用户注册成功")
            return nil
        }),
    })

    // 获取用户列表服务
    app.Register(mod.Service{
        Name:        "user-list",
        DisplayName: "获取用户列表",
        Description: "分页获取用户列表，支持关键词搜索",
        SkipAuth:    false, // 需要认证
        Handler: mod.MakeHandler(func(ctx *mod.Context, req *UserListRequest, resp *UserListResponse) error {
            logger := ctx.GetLogger()

            // 从缓存获取数据
            cacheKey := fmt.Sprintf("user_list_%d_%d_%s", req.Page, req.PageSize, req.Keyword)
            if cached, err := app.GetCache(cacheKey); err == nil {
                logger.Info("从缓存获取用户列表")
                return json.Unmarshal([]byte(cached), resp)
            }

            // 模拟数据库查询
            users := []UserInfo{
                {
                    ID:       "user_001",
                    Username: "alice",
                    Email:    "alice@example.com",
                    Created:  time.Now().Add(-24 * time.Hour),
                },
                {
                    ID:       "user_002",
                    Username: "bob",
                    Email:    "bob@example.com",
                    Created:  time.Now().Add(-12 * time.Hour),
                },
            }

            // 过滤搜索结果
            if req.Keyword != "" {
                filtered := []UserInfo{}
                for _, user := range users {
                    if strings.Contains(user.Username, req.Keyword) ||
                       strings.Contains(user.Email, req.Keyword) {
                        filtered = append(filtered, user)
                    }
                }
                users = filtered
            }

            // 分页处理
            start := (req.Page - 1) * req.PageSize
            end := start + req.PageSize
            if start > len(users) {
                users = []UserInfo{}
            } else if end > len(users) {
                users = users[start:]
            } else {
                users = users[start:end]
            }

            resp.Users = users
            resp.Total = len(users)
            resp.Page = req.Page

            // 缓存结果
            if data, err := json.Marshal(resp); err == nil {
                app.SetCache(cacheKey, string(data), 5*time.Minute)
            }

            logger.WithFields(map[string]interface{}{
                "page":      req.Page,
                "page_size": req.PageSize,
                "keyword":   req.Keyword,
                "count":     len(users),
            }).Info("获取用户列表成功")

            return nil
        }),
    })

    // 文件上传处理
    app.Post("/api/upload", func(c *fiber.Ctx) error {
        file, err := c.FormFile("file")
        if err != nil {
            return c.Status(400).JSON(fiber.Map{
                "error": "文件上传失败",
                "detail": err.Error(),
            })
        }

        // 这里可以添加自定义文件处理逻辑
        return c.JSON(fiber.Map{
            "message": "文件上传成功",
            "filename": file.Filename,
            "size": file.Size,
        })
    })

    // 健康检查
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "ok",
            "timestamp": time.Now(),
            "version": "1.0.0",
        })
    })

    // 启动服务
    app.Run()
}

// 生成用户Token（示例实现）
func generateUserToken(userID string) (string, error) {
    // 这里应该使用JWT库生成真实的Token
    return "jwt_token_" + userID, nil
}
```

### 生产环境配置示例

```yaml
# production.yml
app:
  name: "production-app"
  display_name: "生产环境应用"
  host: "0.0.0.0"
  port: 8080
  body_limit: "50MB"
  read_timeout: "30s"
  write_timeout: "30s"

# 生产环境文件上传配置
file_upload:
  # 优先使用云存储
  s3:
    enabled: true
    bucket: "my-production-bucket"
    region: "us-west-2"
    access_key: "${S3_ACCESS_KEY}"      # 使用环境变量
    secret_key: "${S3_SECRET_KEY}"

  # OSS作为备选
  oss:
    enabled: true
    bucket: "backup-bucket"
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    access_key_id: "${OSS_ACCESS_KEY}"
    access_key_secret: "${OSS_SECRET_KEY}"

  # 本地存储作为最后备选
  local:
    enabled: true
    upload_dir: "/var/uploads"
    max_size: "100MB"

# 生产环境日志配置
logging:
  console:
    enabled: false  # 生产环境关闭控制台日志
    level: "warn"

  file:
    enabled: true
    path: "/var/log/app/app.log"
    max_size: "500MB"
    max_backups: 30
    max_age: "90d"
    compress: true

  # 日志收集服务
  loki:
    enabled: true
    url: "http://loki:3100/loki/api/v1/push"
    labels:
      environment: "production"
      service: "api-server"

# 生产环境缓存配置
cache:
  # Redis集群
  redis:
    enabled: true
    address: "redis-cluster:6379"
    password: "${REDIS_PASSWORD}"
    pool_size: 20
    min_idle_conns: 5

  # 本地缓存作为L1缓存
  bigcache:
    enabled: true
    hard_max_cache_size: 1024  # 1GB
    life_window: "5m"

  # 持久化缓存
  badger:
    enabled: true
    path: "/var/data/badger"

# JWT配置
token:
  jwt:
    enabled: true
    secret_key: "${JWT_SECRET_KEY}"     # 使用环境变量
    expire_duration: "2h"               # 生产环境缩短过期时间
    refresh_expire_duration: "24h"
    algorithm: "HS256"
```

## 🚀 高级特性

### 中间件支持

```go
app.Use(func(c *fiber.Ctx) error {
    // 自定义中间件逻辑
    return c.Next()
})
```

### 自定义路由

```go
app.Get("/health", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok"})
})

app.Post("/webhook", func(c *fiber.Ctx) error {
    // Webhook处理逻辑
    return c.SendStatus(200)
})
```

### 错误处理

```go
func handleService(ctx *mod.Context, req *Request, resp *Response) error {
    // 业务错误
    if req.ID == 0 {
        return mod.ReplyWithDetail(400, "参数错误", "ID不能为空")
    }

    // 系统错误
    if err := someOperation(); err != nil {
        ctx.GetLogger().WithError(err).Error("操作失败")
        return mod.ReplyWithDetail(500, "系统错误", err.Error())
    }

    return nil
}
```

## 🛠️ 开发工具

### API文档

启动应用后访问 `/services/docs` 查看自动生成的API文档，包含：

- 服务列表和描述
- 请求参数结构
- 响应数据格式
- 在线测试界面

### 健康检查

- `GET /health` - 基础健康检查
- `GET /services/ping` - 服务可用性检查

## 🔧 性能优化

### 缓存策略

框架支持三层缓存架构：

1. **BigCache** - 内存缓存，毫秒级响应
2. **BadgerDB** - 本地持久化，适合单机部署
3. **Redis** - 分布式缓存，适合集群部署

### 文件上传优化

- 支持大文件分块上传
- 自动文件类型验证
- 智能存储后端选择（S3 > OSS > Local）
- 文件去重和压缩

### 日志性能

- 异步日志写入
- 自动日志轮转
- 结构化JSON格式（文件）
- 彩色文本格式（控制台）

## 📖 API文档

### 内置端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/services/docs` | API文档 |
| GET | `/services/ping` | 服务检查 |
| POST | `/upload` | 单文件上传 |
| POST | `/upload/batch` | 批量文件上传 |
| POST | `/services/{service_name}` | 自定义服务 |

### 响应格式

```json
{
  "code": 0,
  "msg": "success",
  "data": {},
  "rid": "req_1234567890"
}
```

## 🤝 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目基于 MIT 许可证开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

- 📖 文档: [GitHub README](https://github.com/iamdanielyin/mod/blob/main/README.md)
- 🐛 问题反馈: [GitHub Issues](https://github.com/iamdanielyin/mod/issues)

## 🙏 致谢

感谢以下开源项目：

- [Fiber](https://github.com/gofiber/fiber) - HTTP Web框架
- [Logrus](https://github.com/sirupsen/logrus) - 结构化日志
- [BadgerDB](https://github.com/dgraph-io/badger) - 嵌入式数据库
- [BigCache](https://github.com/allegro/bigcache) - 内存缓存
- [Validator](https://github.com/go-playground/validator) - 参数验证
- [MinIO Go Client](https://github.com/minio/minio-go) - S3兼容存储
- [Alibaba Cloud OSS SDK](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2) - 阿里云对象存储

---

⭐ 如果这个项目对你有帮助，请给我们一个星标！