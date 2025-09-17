# Examples

这个目录包含了 mod 框架的示例代码。

## 文件结构

```
examples/
├── types/
│   └── common.go      # 共享的类型定义
├── basic_demo.go      # 基础功能演示
└── simple_demo.go     # 完整的服务注册演示
```

## 运行示例

### 基础演示
```bash
go run examples/basic_demo.go
```
服务将在 :3000 端口启动

### 完整演示
```bash
go run examples/simple_demo.go
```
服务将在 :8080 端口启动

## 测试服务

### 基础登录
```bash
curl -X POST http://localhost:3000/services/basic_login \
  -H "Content-Type: application/json" \
  -d '{"password":"test123"}'
```

### 错误处理演示
```bash
curl -X POST http://localhost:3000/services/error_demo \
  -H "Content-Type: application/json" \
  -d '{"username":"error","password":"test123"}'
```

### 管理员登录
```bash
curl -X POST http://localhost:8080/services/admin_login \
  -H "Content-Type: application/json" \
  -d '{"password":"admin123"}'
```

### 用户资料
```bash
curl -X POST http://localhost:8080/services/user_profile \
  -H "Content-Type: application/json" \
  -d '{"userID":"user456","name":"Test User"}'
```

## 注意事项

- 所有结构体类型定义都在 `types/common.go` 中，避免重复定义
- 使用统一的 `mod.Service{}` 结构体进行服务注册
- 使用 `mod.MakeHandler()` 创建类型安全的处理函数