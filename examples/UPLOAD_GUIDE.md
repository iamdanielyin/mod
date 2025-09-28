# 文件上传功能使用指南

MOD Framework 提供了强大的本地文件上传功能，支持单文件和批量文件上传，具备完善的安全验证和灵活的配置选项。

## 功能特性

- ✅ **单文件上传**: 支持通过 `POST /upload` 上传单个文件
- ✅ **批量上传**: 支持通过 `POST /upload/batch` 批量上传多个文件
- ✅ **文件类型验证**: 支持MIME类型和文件扩展名双重验证
- ✅ **大小限制**: 可配置单文件最大大小限制
- ✅ **安全验证**: 内置路径安全验证，防止路径遍历攻击
- ✅ **灵活命名**: 支持保持原文件名或随机生成文件名
- ✅ **目录管理**: 支持自动创建目录和按日期组织文件
- ✅ **重名处理**: 自动处理文件名冲突

## 配置方式

在 `mod.yml` 配置文件中添加 `file_upload.local` 部分：

```yaml
file_upload:
  local:
    enabled: true                  # 启用本地文件上传
    upload_dir: "./uploads"        # 上传目录路径
    max_size: "10MB"               # 单文件最大大小
    allowed_types:                 # 允许的MIME类型
      - "image/jpeg"
      - "image/png"
      - "application/pdf"
    allowed_exts:                  # 允许的文件扩展名
      - ".jpg"
      - ".jpeg"
      - ".png"
      - ".pdf"
    keep_original_name: false      # 是否保持原始文件名
    auto_create_dir: true          # 自动创建上传目录
    date_sub_dir: true             # 按日期创建子目录
```

## 配置参数详解

| 参数 | 类型 | 必填 | 说明 | 默认值 |
|------|------|------|------|--------|
| `enabled` | bool | ✅ | 是否启用文件上传功能 | false |
| `upload_dir` | string | ✅ | 文件上传目录路径 | - |
| `max_size` | string | ❌ | 单文件最大大小 | "10MB" |
| `allowed_types` | []string | ❌ | 允许的MIME类型列表 | - |
| `allowed_exts` | []string | ❌ | 允许的文件扩展名列表 | - |
| `keep_original_name` | bool | ❌ | 是否保持原始文件名 | false |
| `auto_create_dir` | bool | ❌ | 自动创建上传目录 | false |
| `date_sub_dir` | bool | ❌ | 按日期创建子目录 | false |

### 大小格式说明

`max_size` 支持以下格式：
- `"10MB"` - 10兆字节
- `"1GB"` - 1吉字节
- `"500KB"` - 500千字节
- `"1048576"` - 纯数字（字节）

### 文件类型控制

可以通过以下方式控制允许的文件类型：

1. **MIME类型前缀匹配**：
   ```yaml
   allowed_types:
     - "image/"          # 所有图片类型
     - "text/"           # 所有文本类型
   ```

2. **具体MIME类型**：
   ```yaml
   allowed_types:
     - "image/jpeg"
     - "application/pdf"
   ```

3. **文件扩展名**：
   ```yaml
   allowed_exts:
     - ".jpg"
     - ".png"
     - ".pdf"
   ```

## API端点

### 单文件上传

**端点**: `POST /upload`

**请求格式**: `multipart/form-data`

**字段名**: `file`

**示例**:
```bash
curl -X POST -F 'file=@example.jpg' http://localhost:8080/upload
```

**成功响应**:
```json
{
  "success": true,
  "message": "文件上传成功",
  "filename": "randomname123.jpg",
  "path": "./uploads/2024/01/15/randomname123.jpg",
  "size": 245760
}
```

### 批量文件上传

**端点**: `POST /upload/batch`

**请求格式**: `multipart/form-data`

**字段名**: `files` (可多个)

**示例**:
```bash
curl -X POST \
  -F 'files=@file1.jpg' \
  -F 'files=@file2.png' \
  http://localhost:8080/upload/batch
```

**成功响应**:
```json
{
  "success": true,
  "message": "批量上传完成，成功: 2, 总数: 2",
  "total": 2,
  "success_count": 2,
  "failed_count": 0,
  "results": [
    {
      "filename": "file1.jpg",
      "size": 245760,
      "success": true,
      "path": "./uploads/randomname1.jpg",
      "saved_filename": "randomname1.jpg"
    },
    {
      "filename": "file2.png",
      "size": 189324,
      "success": true,
      "path": "./uploads/randomname2.png",
      "saved_filename": "randomname2.png"
    }
  ]
}
```

## 使用示例

### 1. 基本配置

```yaml
# 最简配置
file_upload:
  local:
    enabled: true
    upload_dir: "./uploads"
    max_size: "10MB"
    auto_create_dir: true
```

### 2. 图片上传配置

```yaml
# 专门用于图片上传
file_upload:
  local:
    enabled: true
    upload_dir: "./images"
    max_size: "5MB"
    allowed_types: ["image/"]
    allowed_exts: [".jpg", ".jpeg", ".png", ".gif"]
    keep_original_name: true
    date_sub_dir: true
```

### 3. 文档上传配置

```yaml
# 用于文档文件上传
file_upload:
  local:
    enabled: true
    upload_dir: "./documents"
    max_size: "50MB"
    allowed_types:
      - "application/pdf"
      - "application/msword"
      - "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    allowed_exts: [".pdf", ".doc", ".docx"]
    keep_original_name: false
    date_sub_dir: true
```

## 安全特性

### 1. 路径安全验证
- 自动检测和阻止路径遍历攻击（`..` 检测）
- 验证上传目录的有效性
- 警告访问工作目录外的路径

### 2. 文件验证
- **双重类型检查**: 同时验证MIME类型和文件扩展名
- **内容检测**: 读取文件头进行MIME类型检测，防止扩展名伪造
- **大小限制**: 严格控制文件大小，防止存储空间滥用

### 3. 文件名安全
- **随机命名**: 默认生成随机文件名，防止文件名冲突和猜测
- **重名处理**: 保持原文件名时自动处理重名冲突
- **扩展名保留**: 确保文件扩展名正确保留

## 目录组织

### 按日期组织 (`date_sub_dir: true`)

```
uploads/
├── 2024/
│   ├── 01/
│   │   ├── 15/
│   │   │   ├── file1.jpg
│   │   │   └── file2.png
│   │   └── 16/
│   │       └── file3.pdf
│   └── 02/
│       └── 01/
│           └── file4.docx
```

### 平铺组织 (`date_sub_dir: false`)

```
uploads/
├── file1.jpg
├── file2.png
├── file3.pdf
└── file4.docx
```

## 错误处理

### 常见错误响应

**文件大小超限**:
```json
{
  "error": "File validation failed",
  "message": "文件大小 15728640 超过限制 10485760"
}
```

**文件类型不允许**:
```json
{
  "error": "File validation failed",
  "message": "文件类型 application/x-msdownload 不被允许"
}
```

**未选择文件**:
```json
{
  "error": "No file provided",
  "message": "请选择要上传的文件"
}
```

## 与静态文件服务集成

配合静态文件挂载功能，可以实现完整的文件上传和访问流程：

```yaml
# 文件上传配置
file_upload:
  local:
    enabled: true
    upload_dir: "./uploads"

# 静态文件访问配置
static_mounts:
  - url_prefix: "/uploads"
    local_path: "./uploads"
    browseable: true  # 允许浏览上传的文件
```

这样配置后：
1. 通过 `POST /upload` 上传文件
2. 通过 `GET /uploads/文件名` 访问文件
3. 通过 `GET /uploads/` 浏览所有文件

## 性能考虑

1. **并发处理**: 支持同时处理多个文件上传请求
2. **内存优化**: 使用流式处理，避免大文件占用过多内存
3. **存储优化**: 按日期组织文件，便于管理和备份
4. **验证效率**: 优先进行文件大小检查，减少不必要的处理

## 生产环境建议

1. **安全配置**:
   - 严格限制 `allowed_types` 和 `allowed_exts`
   - 设置合理的 `max_size` 限制
   - 使用 `keep_original_name: false` 避免文件名冲突

2. **存储管理**:
   - 启用 `date_sub_dir: true` 便于文件管理
   - 定期清理旧文件或迁移到云存储
   - 配置备份策略

3. **监控告警**:
   - 监控上传目录的磁盘使用情况
   - 记录上传失败的文件和原因
   - 设置上传频率限制（可在反向代理层实现）