# 文件日志示例

这个示例演示了如何使用MOD框架的文件日志功能，包括日志轮转、多级别日志记录和结构化日志输出。

## 功能特性

- 📝 **多级别日志**: 支持 Debug、Info、Warn、Error 四个级别
- 🔄 **自动轮转**: 基于文件大小和时间的自动日志轮转
- 📦 **日志压缩**: 自动压缩历史日志文件节省空间
- 🎯 **双输出**: 同时输出到控制台和文件
- 🏷️ **结构化记录**: 支持结构化字段和上下文信息
- ⚡ **高性能**: 基于 logrus 和 lumberjack 的高性能日志系统

## 快速开始

### 1. 环境要求

- Go 1.24.2 或更高版本

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置文件

编辑 `mod.yml` 文件，配置日志参数：

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
```

### 4. 运行应用

```bash
go run main.go
```

应用将在 `http://localhost:8080` 启动。

## 使用指南

### Web界面

- **首页**: `http://localhost:8080` - 查看功能介绍和配置指南
- **日志测试**: `http://localhost:8080/test` - 测试文件日志功能
- **API文档**: `http://localhost:8080/services/docs` - 查看完整API文档

### API端点

#### 日志测试服务
```bash
curl -X POST http://localhost:8080/services/log-test \
  -H "Content-Type: application/json" \
  -d '{
    "message": "这是一条测试日志",
    "level": "info"
  }'
```

#### 错误测试服务
```bash
curl -X POST http://localhost:8080/services/error-test \
  -H "Content-Type: application/json" \
  -d '{
    "error_type": "business",
    "message": "测试业务错误"
  }'
```

### 查看日志文件

```bash
# 实时查看日志
tail -f ./logs/app.log

# 查看完整日志
cat ./logs/app.log

# 查看压缩的历史日志
ls -la ./logs/
```

## 配置说明

### 日志配置参数

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `console.enabled` | 是否启用控制台输出 | `true` | `true` |
| `console.level` | 控制台日志级别 | `"info"` | `"debug"` |
| `file.enabled` | 是否启用文件输出 | `false` | `true` |
| `file.path` | 日志文件路径 | - | `"./logs/app.log"` |
| `file.max_size` | 单文件最大大小 | `"100MB"` | `"50MB"` |
| `file.max_backups` | 历史文件保留数量 | `10` | `5` |
| `file.max_age` | 文件保留天数 | `"30d"` | `"7d"` |
| `file.compress` | 是否压缩历史文件 | `true` | `false` |

### 日志级别

- **Debug**: 详细的调试信息，开发阶段使用
- **Info**: 一般信息，记录程序运行状态
- **Warn**: 警告信息，潜在问题但不影响运行
- **Error**: 错误信息，影响功能但程序可继续运行

### 日志格式

#### 控制台输出（文本格式）
```
INFO[2024-01-15T10:30:45+08:00] Log test request processed  request_id=abc123 level=info message="测试消息" service=log-test
```

#### 文件输出（JSON格式）
```json
{
  "level": "info",
  "msg": "Log test request processed",
  "request_id": "abc123",
  "level": "info",
  "message": "测试消息",
  "service": "log-test",
  "time": "2024-01-15T10:30:45+08:00"
}
```

## 技术架构

### 日志库集成

- **logrus**: 主要日志库，提供结构化日志和多级别支持
- **lumberjack**: 日志轮转库，提供文件大小和时间基础的轮转
- **io.MultiWriter**: 实现同时输出到多个目标

### 轮转机制

1. **大小轮转**: 当文件达到 `max_size` 时自动创建新文件
2. **数量限制**: 保留最近 `max_backups` 个历史文件
3. **时间清理**: 自动删除超过 `max_age` 的旧文件
4. **压缩存储**: 历史文件自动压缩为 `.gz` 格式

### 性能优化

- **异步写入**: 日志写入不阻塞主业务逻辑
- **缓冲输出**: 内置缓冲机制提高写入效率
- **结构化字段**: 避免字符串拼接，提高性能

## 最佳实践

### 1. 日志级别使用

```go
// Debug: 详细调试信息
logger.Debug("Processing request", "user_id", userID)

// Info: 业务流程信息
logger.Info("User login successful", "user_id", userID)

// Warn: 潜在问题警告
logger.Warn("High memory usage detected", "usage", memUsage)

// Error: 错误和异常
logger.Error("Database connection failed", "error", err)
```

### 2. 结构化字段

```go
// 推荐：使用结构化字段
logger.WithFields(logrus.Fields{
    "user_id":    userID,
    "request_id": requestID,
    "action":     "upload_file",
    "file_size":  fileSize,
}).Info("File upload completed")

// 不推荐：字符串拼接
logger.Info(fmt.Sprintf("User %s uploaded file %s", userID, fileName))
```

### 3. 上下文信息

```go
// 在请求处理中包含上下文
logger.WithFields(logrus.Fields{
    "request_id": ctx.GetRequestID(),
    "user_agent": ctx.Get("User-Agent"),
    "ip":         ctx.IP(),
    "method":     ctx.Method(),
    "path":       ctx.Path(),
}).Info("Request processed")
```

## 监控和分析

### 日志文件分析

```bash
# 统计各级别日志数量
grep -c '"level":"error"' ./logs/app.log
grep -c '"level":"warn"' ./logs/app.log
grep -c '"level":"info"' ./logs/app.log

# 查找特定错误
grep '"level":"error"' ./logs/app.log | jq '.msg'

# 分析请求频率
grep 'request_id' ./logs/app.log | jq -r '.time' | sort | uniq -c
```

### 日志轮转监控

```bash
# 查看日志文件大小
du -h ./logs/app.log*

# 查看轮转文件列表
ls -la ./logs/app.log*

# 检查压缩效果
ls -lah ./logs/app.log*.gz
```

## 故障排除

### 常见问题

1. **日志文件创建失败**
   - 检查目录权限
   - 确认磁盘空间充足
   - 验证文件路径正确

2. **日志轮转不工作**
   - 检查 `max_size` 配置
   - 确认 lumberjack 版本正确
   - 查看控制台错误信息

3. **性能问题**
   - 调整日志级别，减少 Debug 日志
   - 增加缓冲区大小
   - 考虑异步日志写入

### 调试技巧

```bash
# 检查日志配置
grep -A 10 "logging:" mod.yml

# 测试日志写入权限
touch ./logs/test.log && rm ./logs/test.log

# 监控日志文件变化
watch -n 1 'ls -la ./logs/'
```

## 许可证

MIT License