# 静态文件服务文档

这个目录包含了文档文件，通过 `/docs` 路径访问。

## 功能特性

- 支持多种文件类型
- 可配置目录浏览
- 支持自定义索引文件
- 高性能的静态文件服务

## 配置说明

在 `mod.yml` 中配置静态文件挂载点：

```yaml
static_mounts:
  - url_prefix: "/docs"
    local_path: "./docs"
    browseable: false
    index_file: "README.md"
```

访问地址: `/docs/README.md`