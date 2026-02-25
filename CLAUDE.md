# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MOD is a modern enterprise-grade web application framework built on Go Fiber (v2.52.9), focused on rapid development, security, and scalability. It provides a service-oriented architecture for building RESTful APIs.

- **Language**: Go 1.24.2
- **Web Framework**: Fiber v2
- **Module**: github.com/iamdanielyin/mod

## Common Commands

```bash
# Run examples
cd examples/basic-services && go run main.go
cd examples/jwt-auth && go run main.go

# Build any example
go build -o app main.go

# Standard Go commands
go mod tidy
go vet
```

## Architecture

### Service-Oriented Design

MOD uses a service registration pattern. Each business function is registered as a service with its handler, metadata, and configuration:

```go
app.Register(mod.Service{
    Name:        "get_user",        // snake_case recommended
    DisplayName: "获取用户信息",
    Description: "根据用户ID获取用户详细信息",
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *GetUserRequest, resp *GetUserResponse) error {
        // handler logic
        return nil
    }),
    Group: "用户管理",
    Sort:  1,
})
```

### Core Components

1. **App** (`app.go`) - Main application struct extending fiber.App, manages service registration and configuration
2. **Context** (`ctx.go`) - Enhanced fiber.Ctx with logging, user info, JWT claims access
3. **JWT System** (`jwt.go`, `jwt_middleware.go`) - Token generation, validation, refresh, revocation with multiple cache backends
4. **Encryption** (`encryption.go`, `encryption_middleware.go`) - AES256-GCM, ChaCha20-Poly1305 symmetric, RSA-OAEP asymmetric encryption
5. **Permission** (`permission.go`) - Rule-based access control with operators: eq, ne, gt, gte, lt, lte, in, not_in, contains, exists
6. **Mock** (`mock.go`) - Automatic mock data generation from response types

### Middleware Execution Order

All global middleware must be registered before services:

```go
app := mod.New()

app.UseEncryption()     // 1. Decrypt request / encrypt response
app.UseOptionalJWT()    // 2. Authenticate (or UseJWT() for mandatory)

// Register services...
app.Run(":8080")
```

### Request Flow

1. Request arrives at Fiber router
2. EncryptionMiddleware decrypts (if enabled)
3. JWTMiddleware validates authentication
4. Service handler executes with type-safe request/response
5. Permission check validates access
6. Response encrypted (if enabled) and returned

## Key Patterns

### Type-Safe Handlers

Use `MakeHandler[I, O]()` for type-safe service handlers where I is the request type and O is the response type:

```go
Handler: mod.MakeHandler(func(ctx *mod.Context, req *Request, resp *Response) error {
    userID := ctx.GetUserID()
    // ...
    return nil
})
```

### JWT Context Methods

```go
ctx.IsAuthenticated()  // Check auth status
ctx.GetUserID()         // Get user ID
ctx.GetUsername()       // Get username
ctx.GetUserRole()       // Get user role
ctx.GetUserEmail()      // Get email
ctx.GetJWTToken()      // Get raw JWT token
ctx.GetJWTClaims()     // Get JWT claims object
```

### Structured Logging

```go
ctx.Info("message")
ctx.WithFields(map[string]interface{}{
    "user_id": "123",
    "action":  "update",
}).Info("user updated")
```

### Service Configuration Options

- `SkipAuth`: Skip JWT authentication for this service
- `ReturnRaw`: Return raw data without wrapping in standard response format
- `Permission`: Configure permission rules for role-based access

## Configuration

Configuration is loaded from `mod.yml` (or path specified by `MOD_PATH` environment variable). Copy `mod.yml.example` to `mod.yml` to get started.

Key configuration sections:
- `app` - Application name, service base path, token keys
- `server` - Host, port, timeouts, CORS
- `token.jwt` - JWT secret, issuer, expire duration
- `encryption` - Global/group/service-level encryption config
- `cache` - BigCache, BadgerDB, or Redis for token caching
- `file_upload` - Local, S3, or OSS backend
- `logging` - Console, file, Loki, or SLS

## API Documentation

When running, API documentation is auto-generated and available at `/services/docs`.

## Examples

See `examples/` directory for complete working examples:
- `basic-services/` - Service registration, validation
- `jwt-auth/` - JWT authentication
- `encryption/` - Service encryption
- `file-upload/` - Multi-backend file upload
- `static-files/` - Static file serving
- `logging/` - Multi-backend logging
- `mock/` - Mock data generation

## Commit Convention

All commits must follow this format:

```
修改类型: 内容
```

**修改类型:**
- `feat`: 提交新功能
- `fix`: 修复了bug
- `docs`: 只修改了文档
- `style`: 调整代码格式，未修改代码逻辑（如修改空格、格式化、缺少分号等）
- `refactor`: 代码重构，既没修复bug也没有添加新功能
- `perf`: 性能优化，提高性能的代码更改
- `test`: 添加或修改代码测试
- `chore`: 对构建流程或辅助工具和依赖库的更改（如文档生成等）
