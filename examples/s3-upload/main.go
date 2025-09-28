package main

import (
	"fmt"
	"reflect"

	"github.com/danielyin/mod"
	"github.com/gofiber/fiber/v2"
)

// 演示如何使用MOD框架的S3文件上传功能
func main() {
	// 1. 创建MOD应用实例
	// 配置文件mod.yml必须存在并配置S3参数
	app := mod.New()

	// 2. 注册一个测试服务，用于演示文件上传后的处理
	testService := mod.Service{
		Name:        "upload-test",
		DisplayName: "S3文件上传测试",
		Description: "测试S3文件上传功能",
		SkipAuth:    true, // 跳过认证，方便测试
		Handler: mod.Handler{
			Func: handleUploadTest,
		},
	}

	err := app.Register(testService)
	if err != nil {
		panic(fmt.Sprintf("注册服务失败: %v", err))
	}

	// 3. 注册一个获取文件列表的服务
	listService := mod.Service{
		Name:        "list-files",
		DisplayName: "文件列表",
		Description: "获取已上传文件列表",
		SkipAuth:    true,
		Handler: mod.Handler{
			Func:       handleListFiles,
			OutputType: reflect.TypeOf(FileListResponse{}),
		},
	}

	err = app.Register(listService)
	if err != nil {
		panic(fmt.Sprintf("注册文件列表服务失败: %v", err))
	}

	// 4. 启动Web界面路由
	app.Get("/", func(c *fiber.Ctx) error {
		return handleIndexPage(c)
	})
	app.Get("/test", func(c *fiber.Ctx) error {
		return handleTestPage(c)
	})

	fmt.Println("S3文件上传测试服务器启动...")
	fmt.Println("访问 http://localhost:8080 查看测试界面")
	fmt.Println("访问 http://localhost:8080/test 进行上传测试")
	fmt.Println("访问 http://localhost:8080/services/docs 查看API文档")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("- POST /upload         - 单文件上传")
	fmt.Println("- POST /upload/batch   - 批量文件上传")
	fmt.Println("- POST /services/upload-test - 上传测试服务")
	fmt.Println("- POST /services/list-files  - 文件列表服务")

	// 5. 启动服务器
	app.Run()
}

// FileListResponse 文件列表响应结构
type FileListResponse struct {
	Files []FileInfo `json:"files" desc:"文件列表"`
	Total int        `json:"total" desc:"文件总数"`
}

// FileInfo 文件信息结构
type FileInfo struct {
	Name   string `json:"name" desc:"文件名"`
	URL    string `json:"url" desc:"访问URL"`
	Size   int64  `json:"size" desc:"文件大小"`
	Bucket string `json:"bucket" desc:"存储桶"`
	Region string `json:"region" desc:"区域"`
}

// handleUploadTest 处理上传测试
func handleUploadTest(ctx *mod.Context, in interface{}, out interface{}) error {
	// 这里可以添加上传后的业务逻辑
	// 比如记录上传日志、更新数据库等
	fmt.Println("S3文件上传测试服务被调用")
	return nil
}

// handleListFiles 处理文件列表请求
func handleListFiles(ctx *mod.Context, in interface{}, out interface{}) error {
	// 这是一个模拟的文件列表
	// 实际项目中应该从S3或数据库获取文件列表
	response := out.(*FileListResponse)

	response.Files = []FileInfo{
		{
			Name:   "example1.jpg",
			URL:    "https://my-s3-bucket.s3.us-west-2.amazonaws.com/2024/01/15/abcd1234.jpg",
			Size:   1024000,
			Bucket: "my-s3-bucket",
			Region: "us-west-2",
		},
		{
			Name:   "example2.pdf",
			URL:    "https://my-s3-bucket.s3.us-west-2.amazonaws.com/2024/01/15/efgh5678.pdf",
			Size:   2048000,
			Bucket: "my-s3-bucket",
			Region: "us-west-2",
		},
	}
	response.Total = len(response.Files)

	return nil
}

// handleIndexPage 首页
func handleIndexPage(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MOD Framework S3上传测试</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; margin-bottom: 30px; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px; background-color: #fafafa; }
        .section h3 { margin-top: 0; color: #666; }
        .links { display: flex; flex-wrap: wrap; gap: 15px; }
        .link { background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; transition: background 0.3s; }
        .link:hover { background: #0056b3; }
        .api-link { background: #28a745; }
        .api-link:hover { background: #1e7e34; }
        .doc-link { background: #6f42c1; }
        .doc-link:hover { background: #5a2d91; }
        .code { background: #f8f9fa; padding: 15px; border-radius: 5px; border-left: 4px solid #007bff; margin: 10px 0; font-family: monospace; }
        .note { background: #fff3cd; padding: 15px; border-radius: 5px; border-left: 4px solid #ffc107; margin: 10px 0; }
        .aws-note { background: #e6f3ff; padding: 15px; border-radius: 5px; border-left: 4px solid #007bff; margin: 10px 0; }
        .minio-note { background: #f0f8ff; padding: 15px; border-radius: 5px; border-left: 4px solid #4caf50; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 MOD Framework S3上传测试</h1>

        <div class="section">
            <h3>📁 S3文件上传功能</h3>
            <p>测试S3兼容存储文件上传功能，支持AWS S3和MinIO等S3兼容存储服务。</p>
            <div class="links">
                <a href="/test" class="link">📤 文件上传测试</a>
                <a href="/upload" class="link api-link">📡 单文件上传API</a>
                <a href="/upload/batch" class="link api-link">📦 批量上传API</a>
            </div>
        </div>

        <div class="section">
            <h3>📚 API文档</h3>
            <p>查看完整的API接口文档，了解所有可用的服务和参数。</p>
            <div class="links">
                <a href="/services/docs" class="link doc-link">📖 API文档</a>
            </div>
        </div>

        <div class="section">
            <h3>🔧 配置要求</h3>
            <div class="aws-note">
                <strong>AWS S3配置：</strong>使用AWS S3存储时的配置示例：
                <div class="code">
file_upload:
  s3:
    enabled: true
    bucket: "my-s3-bucket"
    region: "us-west-2"
    access_key: "AKIAIOSFODNN7EXAMPLE"
    secret_key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    endpoint: ""  # 留空使用AWS S3
                </div>
            </div>

            <div class="minio-note">
                <strong>MinIO配置：</strong>使用MinIO等S3兼容存储时的配置示例：
                <div class="code">
file_upload:
  s3:
    enabled: true
    bucket: "my-minio-bucket"
    region: "us-east-1"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    endpoint: "http://localhost:9000"  # MinIO服务端点
                </div>
            </div>
        </div>

        <div class="section">
            <h3>🧪 测试指南</h3>
            <p><strong>1. 单文件上传测试：</strong></p>
            <div class="code">
curl -X POST http://localhost:8080/upload \\
  -F "file=@/path/to/your/file.jpg"
            </div>

            <p><strong>2. 批量上传测试：</strong></p>
            <div class="code">
curl -X POST http://localhost:8080/upload/batch \\
  -F "files=@/path/to/file1.jpg" \\
  -F "files=@/path/to/file2.pdf"
            </div>

            <p><strong>3. 自定义服务测试：</strong></p>
            <div class="code">
curl -X POST http://localhost:8080/services/upload-test
curl -X POST http://localhost:8080/services/list-files
            </div>
        </div>

        <div class="section">
            <h3>⚡ 特性说明</h3>
            <ul>
                <li><strong>兼容性</strong>：支持AWS S3和MinIO等S3兼容存储</li>
                <li><strong>优先级</strong>：S3存储拥有最高优先级（S3 > OSS > 本地存储）</li>
                <li><strong>安全性</strong>：支持SSL/TLS加密传输</li>
                <li><strong>灵活性</strong>：可配置自定义端点和区域</li>
                <li><strong>自动化</strong>：自动检测MIME类型和生成访问URL</li>
            </ul>
        </div>
    </div>
</body>
</html>`
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

// handleTestPage 上传测试页面
func handleTestPage(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>S3文件上传测试</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .form-group { margin: 20px 0; }
        label { display: block; margin-bottom: 5px; font-weight: bold; color: #555; }
        input[type="file"] { width: 100%; padding: 10px; border: 2px dashed #ccc; border-radius: 5px; background: #fafafa; }
        button { background: #007bff; color: white; padding: 12px 30px; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; margin: 10px 5px; }
        button:hover { background: #0056b3; }
        .batch-btn { background: #28a745; }
        .batch-btn:hover { background: #1e7e34; }
        .result { margin: 20px 0; padding: 15px; border-radius: 5px; display: none; }
        .success { background: #d4edda; border: 1px solid #c3e6cb; color: #155724; }
        .error { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; }
        .loading { text-align: center; color: #666; }
        pre { background: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; }
        .back-link { display: inline-block; margin-bottom: 20px; color: #007bff; text-decoration: none; }
        .back-link:hover { text-decoration: underline; }
        .info-box { background: #e6f3ff; padding: 15px; border-radius: 5px; border-left: 4px solid #007bff; margin-bottom: 20px; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <a href="/" class="back-link">← 返回首页</a>
        <h1>📤 S3文件上传测试</h1>

        <div class="info-box">
            <strong>🔧 提示：</strong>确保mod.yml中已正确配置S3参数。上传成功后文件将存储到配置的S3存储桶中。
        </div>

        <div class="form-group">
            <label for="singleFile">单文件上传：</label>
            <input type="file" id="singleFile" accept="image/*,.pdf,.txt">
            <button onclick="uploadSingle()">上传到S3</button>
        </div>

        <div class="form-group">
            <label for="multipleFiles">批量文件上传：</label>
            <input type="file" id="multipleFiles" multiple accept="image/*,.pdf,.txt">
            <button onclick="uploadBatch()" class="batch-btn">批量上传到S3</button>
        </div>

        <div id="result" class="result"></div>
    </div>

    <script>
        function showResult(content, isSuccess = true) {
            const resultDiv = document.getElementById('result');
            resultDiv.className = 'result ' + (isSuccess ? 'success' : 'error');
            resultDiv.innerHTML = content;
            resultDiv.style.display = 'block';
        }

        function showLoading() {
            showResult('<div class="loading">⏳ 正在上传到S3中...</div>', true);
        }

        async function uploadSingle() {
            const fileInput = document.getElementById('singleFile');
            const file = fileInput.files[0];

            if (!file) {
                showResult('❌ 请选择要上传的文件', false);
                return;
            }

            showLoading();

            const formData = new FormData();
            formData.append('file', file);

            try {
                const response = await fetch('/upload', {
                    method: 'POST',
                    body: formData
                });

                const result = await response.json();

                if (response.ok && result.success) {
                    const data = result.data;
                    showResult(
                        '✅ 上传到S3成功！<br>' +
                        '<strong>存储后端：</strong> ' + result.backend + '<br>' +
                        '<strong>文件名：</strong> ' + data.filename + '<br>' +
                        '<strong>存储桶：</strong> ' + data.bucket + '<br>' +
                        '<strong>区域：</strong> ' + data.region + '<br>' +
                        '<strong>对象键：</strong> ' + data.object_key + '<br>' +
                        '<strong>访问URL：</strong><br>' +
                        '<a href="' + data.url + '" target="_blank">' + data.url + '</a><br>' +
                        '<strong>文件大小：</strong> ' + data.size + ' bytes',
                        true
                    );
                } else {
                    showResult(
                        '❌ 上传失败：<br>' +
                        '<pre>' + JSON.stringify(result, null, 2) + '</pre>',
                        false
                    );
                }
            } catch (error) {
                showResult('❌ 网络错误：' + error.message, false);
            }
        }

        async function uploadBatch() {
            const fileInput = document.getElementById('multipleFiles');
            const files = fileInput.files;

            if (files.length === 0) {
                showResult('❌ 请选择要上传的文件', false);
                return;
            }

            showLoading();

            const formData = new FormData();
            for (let i = 0; i < files.length; i++) {
                formData.append('files', files[i]);
            }

            try {
                const response = await fetch('/upload/batch', {
                    method: 'POST',
                    body: formData
                });

                const result = await response.json();

                if (response.ok && result.success) {
                    showResult(
                        '✅ 批量上传到S3完成！<br>' +
                        '<strong>成功：</strong> ' + result.success_count + ' 个<br>' +
                        '<strong>失败：</strong> ' + result.failed_count + ' 个<br>' +
                        '<strong>存储后端：</strong> ' + result.backend + '<br>' +
                        '<strong>详细结果：</strong><br>' +
                        '<pre>' + JSON.stringify(result.results, null, 2) + '</pre>',
                        true
                    );
                } else {
                    showResult(
                        '❌ 批量上传失败：<br>' +
                        '<pre>' + JSON.stringify(result, null, 2) + '</pre>',
                        false
                    );
                }
            } catch (error) {
                showResult('❌ 网络错误：' + error.message, false);
            }
        }
    </script>
</body>
</html>`
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}
