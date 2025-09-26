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
	createUploadDirectories()

	// 注册文件管理API服务
	app.Register(mod.Service{
		Name:        "file-manager",
		DisplayName: "File Manager API",
		Description: "Manage uploaded files",
		Handler: mod.Handler{
			Func: func(ctx *mod.Context, in interface{}, out interface{}) error {
				return ctx.JSON(map[string]interface{}{
					"message": "文件上传服务已启用",
					"endpoints": map[string]string{
						"POST /upload":       "单文件上传",
						"POST /upload/batch": "批量文件上传",
						"GET /uploads/":      "浏览已上传文件",
					},
					"features": []string{
						"文件类型验证（MIME类型和扩展名）",
						"文件大小限制",
						"路径安全验证",
						"随机文件名生成",
						"日期子目录组织",
						"批量上传支持",
					},
					"upload_config": map[string]interface{}{
						"max_size":           "20MB",
						"allowed_types":      []string{"image/", "text/", "application/json", "application/pdf"},
						"auto_create_dir":    true,
						"date_sub_dir":       false,
						"keep_original_name": true,
					},
				})
			},
		},
	})

	// 注册文件信息查询服务
	app.Register(mod.Service{
		Name:        "file-info",
		DisplayName: "File Info API",
		Description: "Get file information",
		Handler: mod.Handler{
			Func: func(ctx *mod.Context, in interface{}, out interface{}) error {
				// 这里可以实现文件信息查询逻辑
				return ctx.JSON(map[string]interface{}{
					"message": "文件信息查询接口",
					"usage": map[string]string{
						"method":  "GET",
						"params":  "filename (可选)",
						"example": "/services/file-info?filename=example.jpg",
					},
				})
			},
		},
	})

	// 启动服务器（配置通过static_test.yml文件加载）
	fmt.Println("🚀 文件上传服务示例启动中...")
	fmt.Println("使用配置文件 examples/static_test.yml")
	fmt.Println()
	fmt.Println("📁 可用的端点：")
	fmt.Println("- POST http://localhost:3000/upload           - 单文件上传")
	fmt.Println("- POST http://localhost:3000/upload/batch     - 批量文件上传")
	fmt.Println("- GET  http://localhost:3000/uploads/         - 浏览上传的文件")
	fmt.Println("- GET  http://localhost:3000/services/file-manager - 文件管理API")
	fmt.Println("- GET  http://localhost:3000/services/file-info    - 文件信息API")
	fmt.Println()
	fmt.Println("📋 上传测试命令：")
	fmt.Println("# 单文件上传")
	fmt.Println("curl -X POST -F 'file=@example.jpg' http://localhost:3000/upload")
	fmt.Println()
	fmt.Println("# 批量文件上传")
	fmt.Println("curl -X POST -F 'files=@file1.jpg' -F 'files=@file2.png' http://localhost:3000/upload/batch")
	fmt.Println()

	app.Run()
}

// createUploadDirectories 创建上传相关的目录和测试文件
func createUploadDirectories() {
	directories := []string{
		"./uploads",
		"./test/mock",
	}

	// 创建目录
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create directory %s: %v\n", dir, err)
		}
	}

	// 创建上传测试说明文件
	uploadReadme := `# 文件上传目录

这个目录用于存储通过API上传的文件。

## 上传配置

- **最大文件大小**: 20MB
- **允许的文件类型**:
  - 图片文件 (image/*)
  - 文本文件 (text/*)
  - JSON文件 (application/json)
  - PDF文件 (application/pdf)
- **允许的扩展名**: .jpg, .jpeg, .png, .gif, .bmp, .webp, .txt, .md, .json, .pdf, .zip
- **文件命名**: 保持原始文件名
- **目录组织**: 不按日期分类（测试环境配置）

## 上传方式

### 单文件上传
` + "```bash" + `
curl -X POST -F 'file=@your-file.jpg' http://localhost:3000/upload
` + "```" + `

### 批量文件上传
` + "```bash" + `
curl -X POST \
  -F 'files=@file1.jpg' \
  -F 'files=@file2.png' \
  -F 'files=@file3.txt' \
  http://localhost:3000/upload/batch
` + "```" + `

## 文件访问

上传成功后，可以通过以下方式访问文件：
- 浏览器访问: http://localhost:3000/uploads/
- 直接访问: http://localhost:3000/uploads/文件名

## 安全特性

- ✅ 文件类型验证（MIME类型检测）
- ✅ 文件扩展名验证
- ✅ 文件大小限制
- ✅ 路径安全验证（防止路径遍历攻击）
- ✅ 自动目录创建
- ✅ 重名文件处理（自动添加时间戳）
`

	// 写入README文件
	readmePath := "./uploads/README.md"
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		if err := os.WriteFile(readmePath, []byte(uploadReadme), 0644); err != nil {
			fmt.Printf("Warning: Failed to create %s: %v\n", readmePath, err)
		}
	}

	// 创建一个示例测试文件
	testFilePath := "./test-upload.txt"
	testContent := `这是一个测试上传文件。

你可以使用以下命令上传这个文件：

curl -X POST -F 'file=@test-upload.txt' http://localhost:3000/upload

上传成功后，可以在 http://localhost:3000/uploads/ 查看上传的文件。
`
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
			fmt.Printf("Warning: Failed to create %s: %v\n", testFilePath, err)
		}
	}

	fmt.Println("✅ Upload directories and test files created successfully")
	fmt.Println("📄 测试文件已创建: test-upload.txt")
}
