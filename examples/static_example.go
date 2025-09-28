package main

import (
	"fmt"
	"github.com/iamdanielyin/mod"
	"os"
)

func main() {
	// 创建应用实例
	app := mod.New()

	// 创建示例目录结构（如果不存在）
	createExampleDirectories()

	// 注册一个API服务来演示静态文件和API的结合使用
	app.Register(mod.Service{
		Name:        "file-info",
		DisplayName: "File Info API",
		Description: "Get information about static files",
		Handler: mod.Handler{
			Func: func(ctx *mod.Context, in interface{}, out interface{}) error {
				return ctx.JSON(map[string]interface{}{
					"message": "Static file server is running",
					"endpoints": map[string]string{
						"/static":  "Static assets (CSS, JS, images)",
						"/uploads": "Uploaded files (browseable)",
						"/docs":    "Documentation files",
					},
					"features": []string{
						"Directory browsing for uploads",
						"Index file serving",
						"Compression support",
						"Security path validation",
					},
				})
			},
		},
	})

	// 启动服务器（配置通过mod.yml文件加载）
	fmt.Println("Static file server example starting...")
	fmt.Println("使用配置文件 examples/static_test.yml 来配置静态文件挂载")
	fmt.Println("访问以下URL来测试静态文件功能：")
	fmt.Println("- http://localhost:3000/static/       - 静态资源")
	fmt.Println("- http://localhost:3000/uploads/      - 上传文件（可浏览）")
	fmt.Println("- http://localhost:3000/docs/         - 文档文件")
	fmt.Println("- http://localhost:3000/services/file-info - API接口")

	app.Run()
}

// createExampleDirectories 创建示例目录结构和文件
func createExampleDirectories() {
	directories := []string{
		"./public",
		"./uploads",
		"./docs",
		"./dev-assets",
		"./test/mock",
	}

	// 创建目录
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create directory %s: %v\n", dir, err)
		}
	}

	// 创建示例文件
	exampleFiles := map[string]string{
		"./public/index.html": `<!DOCTYPE html>
<html>
<head>
    <title>MOD Framework Static Example</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>MOD Framework 静态文件服务</h1>
        <p>这是一个通过MOD框架static_mounts功能提供的静态HTML文件。</p>

        <h2>可用的端点：</h2>
        <div class="endpoint">
            <strong>GET /static/</strong> - 静态资源文件
        </div>
        <div class="endpoint">
            <strong>GET /uploads/</strong> - 上传文件目录（可浏览）
        </div>
        <div class="endpoint">
            <strong>GET /docs/</strong> - 文档文件目录
        </div>
        <div class="endpoint">
            <strong>GET /services/file-info</strong> - API接口
        </div>

        <h2>功能特性：</h2>
        <ul>
            <li>✅ 支持目录浏览（uploads目录）</li>
            <li>✅ 默认首页文件（index.html）</li>
            <li>✅ 文件压缩支持</li>
            <li>✅ 安全路径验证</li>
            <li>✅ 范围请求支持</li>
        </ul>
    </div>
</body>
</html>`,

		"./uploads/README.txt": `这是上传文件目录。

在生产环境中，这个目录通常用于存储用户上传的文件。
在测试配置中，此目录设置为可浏览模式，方便开发和调试。

安全提示：
- 生产环境应该关闭目录浏览功能
- 建议对上传文件进行类型和大小限制
- 考虑使用CDN来提供静态文件服务`,

		"./docs/index.html": `<!DOCTYPE html>
<html>
<head>
    <title>API 文档</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f9f9f9; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    </style>
</head>
<body>
    <div class="container">
        <h1>📚 API 文档</h1>
        <p>欢迎使用 MOD Framework API 文档页面。</p>

        <h2>可用的API端点：</h2>
        <ul>
            <li><code>GET /services/file-info</code> - 获取文件服务信息</li>
        </ul>

        <h2>静态文件端点：</h2>
        <ul>
            <li><code>GET /static/</code> - 静态资源</li>
            <li><code>GET /uploads/</code> - 上传文件</li>
            <li><code>GET /docs/</code> - 文档文件</li>
        </ul>
    </div>
</body>
</html>`,

		"./dev-assets/README.md": `# 开发资源目录

这个目录包含开发环境专用的资源文件。

## 特性

- **目录浏览**: 启用
- **默认文件**: README.md
- **用途**: 开发和调试

## 使用方法

访问 http://localhost:3000/dev/ 来浏览此目录的内容。`,

		"./test/mock/index.json": `{
  "message": "这是测试和Mock数据目录",
  "features": [
    "JSON格式的默认文件",
    "测试数据存储",
    "Mock API响应"
  ],
  "endpoints": {
    "static": "/static/",
    "uploads": "/uploads/",
    "docs": "/docs/",
    "dev": "/dev/",
    "mock": "/mock/"
  }
}`,
	}

	// 创建示例文件
	for filePath, content := range exampleFiles {
		// 检查文件是否已存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				fmt.Printf("Warning: Failed to create file %s: %v\n", filePath, err)
			}
		}
	}

	fmt.Println("✅ Example directories and files created successfully")
}
