# S3文件上传示例

这个示例演示了如何使用MOD框架的S3文件上传功能，支持AWS S3和MinIO等S3兼容存储服务。

## 功能特性

- 🚀 **S3兼容存储**: 支持AWS S3和MinIO等S3兼容存储服务
- 📤 **多种上传方式**: 单文件上传和批量文件上传
- 🎯 **优先级系统**: S3 > OSS > 本地存储的自动选择
- 🔒 **安全验证**: 文件类型、大小和MIME类型验证
- 🌐 **Web界面**: 提供友好的文件上传测试界面
- 📊 **API文档**: 自动生成的API接口文档

## 快速开始

### 1. 环境要求

- Go 1.24.2 或更高版本
- AWS S3账户或MinIO服务器

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置文件

编辑 `mod.yml` 文件，配置S3参数：

#### AWS S3配置
```yaml
file_upload:
  s3:
    enabled: true
    bucket: "your-s3-bucket"
    region: "us-west-2"
    access_key: "AKIAIOSFODNN7EXAMPLE"
    secret_key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    endpoint: ""  # 留空使用AWS S3
```

#### MinIO配置
```yaml
file_upload:
  s3:
    enabled: true
    bucket: "my-minio-bucket"
    region: "us-east-1"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    endpoint: "http://localhost:9000"  # MinIO服务端点
```

### 4. 运行应用

```bash
go run main.go
```

应用将在 `http://localhost:8080` 启动。

## 使用指南

### Web界面

- **首页**: `http://localhost:8080` - 查看功能介绍和配置指南
- **上传测试**: `http://localhost:8080/test` - 测试文件上传功能
- **API文档**: `http://localhost:8080/services/docs` - 查看完整API文档

### API端点

#### 单文件上传
```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@/path/to/your/file.jpg"
```

#### 批量文件上传
```bash
curl -X POST http://localhost:8080/upload/batch \
  -F "files=@/path/to/file1.jpg" \
  -F "files=@/path/to/file2.pdf"
```

#### 自定义服务
```bash
# 上传测试服务
curl -X POST http://localhost:8080/services/upload-test

# 文件列表服务
curl -X POST http://localhost:8080/services/list-files
```

## 配置说明

### S3配置参数

| 参数 | 说明 | 必填 | 示例 |
|------|------|------|------|
| `enabled` | 是否启用S3上传 | 是 | `true` |
| `bucket` | S3存储桶名称 | 是 | `"my-s3-bucket"` |
| `region` | AWS区域 | 是 | `"us-west-2"` |
| `access_key` | AWS访问密钥ID | 是 | `"AKIAIOSFODNN7EXAMPLE"` |
| `secret_key` | AWS秘密访问密钥 | 是 | `"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"` |
| `endpoint` | 自定义端点（MinIO等） | 否 | `"http://localhost:9000"` |

### 存储后端优先级

MOD框架支持多种存储后端，按以下优先级自动选择：

1. **S3存储** (最高优先级)
2. **OSS存储** (中等优先级)
3. **本地存储** (最低优先级)

### 文件验证

- **大小限制**: 默认10MB，可通过`local.max_size`配置
- **类型限制**: 支持图片、PDF、文本文件
- **扩展名验证**: 白名单机制，确保安全

## 技术架构

### S3集成

使用 `github.com/minio/minio-go/v7` 库实现S3兼容存储：

- **连接测试**: 启动时自动验证S3配置
- **对象键生成**: 按日期组织文件路径 `YYYY/MM/DD/随机名.ext`
- **URL生成**: 自动生成正确的访问URL
- **MIME检测**: 自动检测和设置文件MIME类型

### 安全特性

- **路径安全**: 防止路径遍历攻击
- **文件验证**: 多层文件类型和大小验证
- **配置验证**: 启动时验证所有存储配置

## 故障排除

### 常见问题

1. **S3连接失败**
   - 检查网络连接
   - 验证访问密钥和密钥正确性
   - 确认存储桶存在且有权限

2. **文件上传失败**
   - 检查文件大小是否超限
   - 验证文件类型是否被允许
   - 确认存储后端配置正确

3. **MinIO连接问题**
   - 确认MinIO服务正在运行
   - 检查端点URL格式
   - 验证SSL/TLS配置

### 日志调试

应用日志保存在 `./logs/app.log`，包含详细的错误信息和调试数据。

## 开发扩展

### 添加新的存储后端

1. 在 `app.go` 中实现新的配置方法
2. 在 `determineUploadBackend()` 中添加优先级
3. 在 `saveUploadFile()` 中添加处理逻辑
4. 更新配置文件示例

### 自定义文件处理

可以通过注册自定义服务来扩展文件上传后的处理逻辑：

```go
customService := mod.Service{
    Name: "custom-upload",
    Handler: mod.Handler{
        Func: func(ctx *mod.Context, in interface{}, out interface{}) error {
            // 自定义文件处理逻辑
            return nil
        },
    },
}
app.Register(customService)
```

## 许可证

MIT License