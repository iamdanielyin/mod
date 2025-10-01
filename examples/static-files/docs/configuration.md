# 配置指南

## 基础配置

```yaml
app:
  name: "MyApp"
  display_name: "我的应用"

server:
  port: 8080
```

## 静态文件配置

```yaml
static_mounts:
  - url_prefix: "/static"
    local_path: "./static"
    browseable: true
    index_file: "index.html"

  - url_prefix: "/docs"
    local_path: "./docs"
    browseable: false
    index_file: "README.md"
```

## JWT 配置

```yaml
token:
  jwt:
    enabled: true
    secret_key: "your-secret-key"
    expire_duration: "24h"
```

## 日志配置

```yaml
logging:
  console:
    enabled: true
    level: "info"
```

## 环境变量

```yaml
token:
  jwt:
    secret_key: ${JWT_SECRET}
```
