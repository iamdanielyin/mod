# Token 缓存测试指南

这个目录包含了用于测试 mod 框架三种 Token 缓存策略的完整示例代码。

## 文件说明

- **basic_demo.go**: 包含 Token 测试服务的主应用
- **test_mod.yml**: 测试配置文件，配置了三种缓存策略
- **test_token_cache.sh**: 自动化测试脚本
- **types/common.go**: 包含 Token 测试相关的类型定义

## 三种缓存策略

### 1. BigCache (内存缓存)
- **优点**: 性能最高，纯内存操作
- **缺点**: 重启后数据丢失，内存使用较多
- **适用场景**: 高并发、对性能要求极高的场景
- **配置**: `cache_strategy: "bigcache"`

### 2. BadgerDB (本地持久化)
- **优点**: 数据持久化，性能较好，无外部依赖
- **缺点**: 单机存储，不支持分布式
- **适用场景**: 单机部署，需要数据持久化的场景
- **配置**: `cache_strategy: "badger"`

### 3. Redis (远程缓存)
- **优点**: 支持分布式，可多实例共享，功能丰富
- **缺点**: 需要外部 Redis 服务，网络延迟
- **适用场景**: 集群部署，多实例共享 Token 的场景
- **配置**: `cache_strategy: "redis"`

## 测试服务

### 1. 基础登录服务 (`basic_login`)
- **功能**: 用户登录并生成 Token
- **特点**: 自动将 Token 存储到配置的缓存中
- **测试**: 验证 Token 创建和存储功能

### 2. Token 验证服务 (`token_verify_test`)
- **功能**: 验证 Token 的有效性
- **特点**: `SkipAuth=false`，需要有效 Token 才能访问
- **测试**: 验证 Token 验证逻辑

### 3. Token 查询服务 (`token_query_test`)
- **功能**: 查询指定 Token 的详细信息
- **特点**: 可以查看 Token 关联的用户数据
- **测试**: 验证 Token 数据检索功能

### 4. Token 登出服务 (`token_logout_test`)
- **功能**: 删除指定 Token，模拟用户登出
- **特点**: 测试 Token 失效机制
- **测试**: 验证 Token 删除功能

### 5. Token 批量测试服务 (`token_batch_test`)
- **功能**: 批量创建多个 Token
- **特点**: 用于性能测试和压力测试
- **测试**: 验证缓存的批量操作性能

## 快速开始

### 1. 启动应用

```bash
# 进入示例目录
cd examples

# 使用测试配置启动应用
MOD_PATH=./test_mod.yml go run basic_demo.go
```

### 2. 运行测试

```bash
# 运行完整测试套件
./test_token_cache.sh

# 或者运行特定测试
./test_token_cache.sh login testuser
./test_token_cache.sh batch 10
```

### 3. 查看文档

访问 http://localhost:3000/services/docs 查看所有 API 接口的详细文档。

## 测试不同缓存策略

### 测试 BigCache

```yaml
# 修改 test_mod.yml
token:
  validation:
    cache_strategy: "bigcache"
    bigcache:
      enabled: true
```

### 测试 BadgerDB

```yaml
# 修改 test_mod.yml
token:
  validation:
    cache_strategy: "badger"
    badger:
      enabled: true
      path: "./data/tokens"
      ttl: "24h"
```

### 测试 Redis

```yaml
# 修改 test_mod.yml
token:
  validation:
    cache_strategy: "redis"
redis:
  enabled: true  # 确保主 Redis 配置启用
token:
  validation:
    redis:
      enabled: true
```

## 手动测试示例

### 1. 用户登录

```bash
curl -X POST http://localhost:3000/services/basic_login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123"}'
```

### 2. Token 验证

```bash
curl -X POST http://localhost:3000/services/token_verify_test \
  -H "Content-Type: application/json" \
  -H "mod-token: YOUR_TOKEN_HERE" \
  -d '{"user_id":"test"}'
```

### 3. Token 查询

```bash
curl -X POST http://localhost:3000/services/token_query_test \
  -H "Content-Type: application/json" \
  -d '{"token":"YOUR_TOKEN_HERE"}'
```

### 4. Token 删除

```bash
curl -X POST http://localhost:3000/services/token_logout_test \
  -H "Content-Type: application/json" \
  -d '{"token":"YOUR_TOKEN_HERE"}'
```

### 5. 批量测试

```bash
curl -X POST http://localhost:3000/services/token_batch_test \
  -H "Content-Type: application/json" \
  -d '{"count":5}'
```

## 日志观察

应用启动后，可以观察日志输出了解缓存操作的详细信息：

- Token 存储日志
- Token 验证日志
- Token 删除日志
- 缓存初始化日志
- 错误处理日志

## 性能测试

使用批量测试服务可以测试不同缓存策略的性能：

```bash
# 测试创建 100 个 Token 的性能
./test_token_cache.sh batch 100

# 测试创建 1000 个 Token 的性能（最大限制）
./test_token_cache.sh batch 1000
```

## 故障排除

### 1. BadgerDB 权限问题
确保 `./data/tokens` 目录有写入权限。

### 2. Redis 连接失败
检查 Redis 服务是否启动，配置的地址和端口是否正确。

### 3. BigCache 内存不足
根据系统内存调整 `hard_max_cache_size` 配置。

### 4. Token 验证失败
检查 `cache_strategy` 配置是否与启用的缓存匹配。

## 总结

通过这套测试代码，你可以：

1. **验证功能**: 确保三种缓存策略都能正常工作
2. **性能对比**: 比较不同策略的性能表现
3. **压力测试**: 测试高并发场景下的表现
4. **故障模拟**: 测试各种异常情况的处理
5. **配置验证**: 验证不同配置参数的效果

选择合适的缓存策略需要根据具体的部署环境、性能要求和可用性需求来决定。