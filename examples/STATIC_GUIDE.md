# 静态文件挂载功能使用指南

MOD Framework 提供了强大的静态文件挂载功能，支持通过配置文件轻松管理静态资源服务。

## 功能特性

- ✅ **多路径挂载**: 支持将多个本地目录挂载到不同的URL路径
- ✅ **目录浏览**: 可选择性启用目录浏览功能
- ✅ **默认文件**: 支持设置目录的默认首页文件
- ✅ **安全验证**: 内置路径安全验证，防止路径遍历攻击
- ✅ **性能优化**: 自动启用文件压缩和范围请求支持
- ✅ **错误处理**: 完善的错误日志和异常处理

## 配置方式

在 `mod.yml` 配置文件中添加 `static_mounts` 部分：

```yaml
static_mounts:
  - url_prefix: "/static"          # URL访问路径
    local_path: "./public"         # 本地文件目录
    browseable: false              # 是否允许目录浏览
    index_file: "index.html"       # 默认首页文件

  - url_prefix: "/uploads"
    local_path: "./uploads"
    browseable: true               # 允许浏览上传文件
    index_file: ""                 # 不设置默认文件
```

## 配置参数说明

| 参数 | 类型 | 必填 | 说明 | 默认值 |
|------|------|------|------|--------|
| `url_prefix` | string | ✅ | URL访问路径前缀 | - |
| `local_path` | string | ✅ | 本地文件系统路径 | - |
| `browseable` | bool | ❌ | 是否允许目录浏览 | false |
| `index_file` | string | ❌ | 目录默认文件名 | "index.html" |

## 使用示例

### 1. 运行示例程序

```bash
# 使用静态文件配置运行示例
cd examples
go run static_example.go
```

程序会自动：
- 创建示例目录结构
- 生成测试文件
- 启动带有静态文件服务的HTTP服务器

### 2. 访问静态文件

启动后可以访问以下URL：

- `http://localhost:3000/static/` - 静态资源首页
- `http://localhost:3000/uploads/` - 文件上传目录（可浏览）
- `http://localhost:3000/docs/` - 文档目录
- `http://localhost:3000/dev/` - 开发资源（使用README.md作为首页）
- `http://localhost:3000/mock/` - Mock数据（使用index.json作为首页）

### 3. 测试功能

运行测试脚本验证功能是否正常：

```bash
# 先启动服务器
go run static_example.go

# 在另一个终端运行测试
go run test_static.go
```

## 安全特性

### 路径安全验证

- **路径遍历保护**: 自动检测和阻止 `..` 路径遍历尝试
- **绝对路径解析**: 将相对路径转换为绝对路径进行验证
- **工作目录检查**: 警告访问工作目录外的路径（不严格禁止）

### 错误处理

- **路径不存在**: 自动跳过不存在的本地路径，记录警告日志
- **配置错误**: 验证必填参数，跳过无效配置
- **权限问题**: 处理文件系统权限错误

## 生产环境建议

1. **目录浏览**: 生产环境建议关闭 `browseable` 功能
2. **路径限制**: 确保 `local_path` 指向安全的目录
3. **文件类型**: 考虑在反向代理层面限制可访问的文件类型
4. **CDN集成**: 大规模部署时建议使用CDN来提供静态文件服务

## 与其他功能集成

### 配合CORS使用

```yaml
app:
  cors:
    enabled: true
    allow_origins: ["https://yourdomain.com"]

static_mounts:
  - url_prefix: "/assets"
    local_path: "./assets"
```

### 配合文件上传使用

```yaml
file_upload:
  local:
    enabled: true
    upload_dir: "./uploads"

static_mounts:
  - url_prefix: "/uploads"
    local_path: "./uploads"
    browseable: false  # 生产环境建议关闭
```

## 性能优化

MOD Framework 的静态文件服务自动启用以下优化：

- **Gzip压缩**: 自动压缩文本文件
- **范围请求**: 支持断点续传和流媒体播放
- **ETag缓存**: 客户端缓存优化（如果在应用配置中启用）

这些优化无需额外配置，框架会自动处理。