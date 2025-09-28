# MOD Framework Examples

这个目录包含了MOD框架各个功能特性的完整示例，每个示例都在独立的文件夹中，包含完整的代码和配置文件。

## 目录结构

```
examples/
├── basic-services/     # 基础服务示例
├── jwt-auth/          # JWT认证示例
├── encryption/        # 加解密示例
├── file-upload/       # 文件上传示例
├── static-files/      # 静态文件服务示例
├── logging/           # 日志功能示例
├── mock/              # Mock功能示例
└── README.md          # 本文件
```

## 示例说明

### 1. 基础服务 (basic-services)
- **功能**: 演示基本的服务注册、参数验证、错误处理
- **包含**: 用户管理服务、参数验证、响应格式
- **运行**: `cd basic-services && go run main.go`
- **API文档**: http://localhost:8080/services/docs

### 2. JWT认证 (jwt-auth)
- **功能**: 完整的JWT认证系统，包括登录、登出、权限控制
- **包含**: Token生成、验证、刷新、撤销、角色权限
- **运行**: `cd jwt-auth && go run main.go`
- **特性**: 支持BigCache、BadgerDB、Redis作为token存储

### 3. 加解密 (encryption)
- **功能**: 服务级别的加解密和签名验证
- **包含**: AES256-GCM对称加密、RSA非对称加密、HMAC-SHA256签名
- **运行**: `cd encryption && go run main.go`
- **配置**: 支持全局、分组、服务级别的加密配置

### 4. 文件上传 (file-upload)
- **功能**: 多后端文件上传支持
- **包含**: 本地存储、S3、阿里云OSS
- **运行**: `cd file-upload && go run main.go`
- **特性**: 文件类型验证、大小限制、批量上传

### 5. 静态文件 (static-files)
- **功能**: 静态文件服务和目录浏览
- **包含**: 多挂载点、目录浏览、自定义索引文件
- **运行**: `cd static-files && go run main.go`
- **访问**: http://localhost:8080/static/index.html

### 6. 日志功能 (logging)
- **功能**: 多种日志输出方式
- **包含**: 控制台、文件、Loki、阿里云SLS
- **运行**: `cd logging && go run main.go`
- **特性**: 日志轮转、结构化日志、多级别日志

### 7. Mock功能 (mock)
- **功能**: 服务Mock功能，用于开发和测试
- **包含**: 全局Mock、分组Mock、服务级Mock
- **运行**: `cd mock && go run main.go`
- **特性**: 智能Mock数据生成

## 快速开始

1. **选择感兴趣的示例**:
   ```bash
   cd basic-services  # 或其他示例目录
   ```

2. **查看配置文件**:
   ```bash
   cat mod.yml  # 查看示例配置
   ```

3. **运行示例**:
   ```bash
   go run main.go
   ```

4. **访问API文档**:
   浏览器打开 http://localhost:8080/services/docs

## 配置文件说明

每个示例都包含一个 `mod.yml` 配置文件，展示了相关功能的配置方法：

- **app**: 应用基础信息配置
- **server**: 服务器相关配置
- **logging**: 日志输出配置
- **file_upload**: 文件上传配置
- **static_mounts**: 静态文件挂载配置
- **jwt**: JWT认证配置
- **encryption**: 加解密配置
- **mock**: Mock功能配置

## 开发建议

1. **先运行基础服务示例**了解MOD框架的基本用法
2. **根据项目需求**选择相应的功能示例进行学习
3. **参考配置文件**了解各项功能的配置方法
4. **查看API文档**了解服务接口的详细信息
5. **结合多个示例**构建完整的应用程序

## 注意事项

- 所有示例都使用端口8080，请确保端口未被占用
- 部分示例需要外部服务（如Redis、Loki等），请根据实际情况配置
- 生产环境请修改相关的密钥和凭证信息
- 文件上传示例会在当前目录创建uploads文件夹

## 技术支持

如有问题，请参考：
1. API文档: http://localhost:8080/services/docs
2. 配置文件注释
3. 源代码注释