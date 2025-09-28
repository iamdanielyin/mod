# Mock功能演示

这个示例演示了MOD框架的Mock功能，支持全局、分组和服务级别的Mock配置。

## 功能特性

- ✅ **三级Mock配置**: 支持全局、分组、服务三个级别的Mock设置
- ✅ **智能数据生成**: 根据结构体字段自动生成合理的Mock数据
- ✅ **优先级控制**: 服务级 > 分组级 > 全局级配置优先级
- ✅ **类型智能推断**: 根据字段名和标签生成特定类型的Mock数据
- ✅ **完整示例界面**: 提供Web界面测试Mock功能
- ✅ **实时Mock状态**: 可查看各服务的Mock状态和配置

## 快速开始

### 1. 启动示例

```bash
cd examples/mock-demo
go mod tidy
go run main.go
```

### 2. 访问测试

打开浏览器访问：
- http://localhost:8080 - Mock功能演示界面
- http://localhost:8080/mock-status - Mock状态查看
- http://localhost:8080/services/docs - API文档

## Mock配置说明

### 配置文件结构

```yaml
# mod.yml
mock:
  # 全局Mock设置
  global:
    enabled: false

  # 分组级别Mock设置
  groups:
    "用户管理":
      enabled: true
    "订单管理":
      enabled: false

  # 服务级别Mock设置
  services:
    "order-info":
      enabled: true
```

### 配置优先级

Mock配置的优先级从高到低为：

1. **服务级配置** - `mock.services.{service_name}.enabled`
2. **分组级配置** - `mock.groups.{group_name}.enabled`
3. **全局级配置** - `mock.global.enabled`

### 本示例的配置

- **全局Mock**: 关闭
- **用户管理分组**: 启用Mock（所有用户相关服务使用Mock数据）
- **订单管理分组**: 关闭Mock
- **order-info服务**: 单独启用Mock（服务级配置覆盖分组配置）
- **消息服务分组**: 关闭Mock（使用实际Handler）

## API测试端点

### 用户管理组 (Mock启用)

#### 获取用户信息
```bash
curl -X POST http://localhost:8080/services/user-info \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test_123"}'
```

#### 获取用户列表
```bash
curl -X POST http://localhost:8080/services/user-list \
  -H "Content-Type: application/json" \
  -d '{"page": 1, "page_size": 10, "keyword": "test"}'
```

### 订单管理组

#### 获取订单信息 (服务级Mock启用)
```bash
curl -X POST http://localhost:8080/services/order-info \
  -H "Content-Type: application/json" \
  -d '{"order_id": "order_456"}'
```

### 消息服务组 (Mock关闭)

#### 发送消息 (使用实际Handler)
```bash
curl -X POST http://localhost:8080/services/send-message \
  -H "Content-Type: application/json" \
  -d '{"to_user_id": "user_789", "message": "测试消息", "type": "text"}'
```

## Mock数据生成规则

### 字段名智能识别

Mock生成器会根据字段名自动生成相应类型的数据：

- `id`, `uid` → `mock_id_12345`
- `name` → `Alice`, `Bob`, `Charlie`...
- `email` → `user123@example.com`
- `phone` → `13812345678`
- `url`, `link` → `https://example.com/mock/123`
- `token` → `mock_token_abc123`
- `address` → `北京市朝阳区`, `上海市浦东新区`...
- `message`, `msg` → `这是一条Mock消息`
- `status` → `active`, `inactive`, `pending`...

### 数据类型支持

- **基础类型**: bool, int, float, string
- **复合类型**: slice, array, map, struct
- **嵌套结构**: 支持任意层级的结构体嵌套
- **指针类型**: 自动处理指针类型字段
- **时间类型**: 自动生成当前时间

### 示例Mock数据

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "mock_id_67890",
    "name": "Alice",
    "email": "user456@example.com",
    "phone": "13823456789",
    "address": "北京市朝阳区",
    "status": "active",
    "create_at": "2024-01-15T10:30:00Z"
  },
  "rid": "req_1234567890"
}
```

## 使用场景

### 1. 前端开发

前端开发者可以在后端接口未完成时使用Mock数据进行开发：

```yaml
mock:
  global:
    enabled: true  # 启用全局Mock
```

### 2. 接口测试

为特定接口启用Mock以便测试：

```yaml
mock:
  services:
    "user-info":
      enabled: true  # 只为user-info接口启用Mock
```

### 3. 分组测试

为整个业务模块启用Mock：

```yaml
mock:
  groups:
    "用户管理":
      enabled: true  # 用户管理相关接口全部使用Mock
```

### 4. 演示环境

在演示环境中使用Mock数据：

```yaml
mock:
  global:
    enabled: true
  services:
    "send-message":
      enabled: false  # 发送消息仍使用真实功能
```

## 开发指南

### 1. 添加新的Mock服务

```go
app.Register(mod.Service{
    Name:        "my-service",
    DisplayName: "我的服务",
    Group:       "我的分组",
    Handler: mod.MakeHandler(func(ctx *mod.Context, req *MyRequest, resp *MyResponse) error {
        // 实际业务逻辑
        // Mock模式下不会执行这里的代码
        return nil
    }),
})
```

### 2. 配置Mock

在 `mod.yml` 中添加相应配置：

```yaml
mock:
  groups:
    "我的分组":
      enabled: true
```

### 3. 自定义Mock数据生成

可以通过字段标签和命名约定影响Mock数据生成：

```go
type MyResponse struct {
    UserID   string    `json:"user_id" desc:"用户ID"`    // 会生成 mock_id_xxx
    UserName string    `json:"name" desc:"用户名"`       // 会生成随机姓名
    Email    string    `json:"email" desc:"邮箱"`       // 会生成邮箱格式
    CreateAt time.Time `json:"create_at" desc:"创建时间"` // 会生成当前时间
}
```

## 注意事项

1. **Mock模式下参数验证仍会执行**，只是跳过Handler函数调用
2. **身份验证检查正常进行**，Mock不会跳过认证流程
3. **日志记录完整**，会记录Mock模式的调用日志
4. **输出格式一致**，Mock数据使用相同的响应格式包装

## 故障排除

### 1. Mock不生效

检查配置文件中的服务名称和分组名称是否准确匹配。

### 2. Mock数据格式不正确

检查输出结构体的字段类型和标签定义。

### 3. 配置优先级问题

记住配置优先级：服务级 > 分组级 > 全局级。

## 扩展功能

### 1. 自定义Mock生成器

可以扩展 `MockGenerator` 来支持更多的数据类型和生成规则。

### 2. Mock数据模板

可以在配置中定义特定的Mock数据模板。

### 3. 条件Mock

可以根据请求参数动态决定是否使用Mock。