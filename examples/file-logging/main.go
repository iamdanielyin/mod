package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/danielyin/mod"
	"github.com/gofiber/fiber/v2"
)

// 演示如何使用MOD框架的文件日志功能
func main() {
	// 1. 创建MOD应用实例
	// 配置文件mod.yml必须存在并配置日志参数
	app := mod.New()

	// 2. 注册一个测试服务，用于演示日志记录
	testService := mod.Service{
		Name:        "log-test",
		DisplayName: "日志测试",
		Description: "测试文件日志记录功能",
		SkipAuth:    true, // 跳过认证，方便测试
		Handler: mod.Handler{
			Func:       handleLogTest,
			InputType:  reflect.TypeOf(LogTestRequest{}),
			OutputType: reflect.TypeOf(LogTestResponse{}),
		},
	}

	err := app.Register(testService)
	if err != nil {
		panic(fmt.Sprintf("注册服务失败: %v", err))
	}

	// 3. 注册一个错误测试服务
	errorService := mod.Service{
		Name:        "error-test",
		DisplayName: "错误测试",
		Description: "测试错误日志记录",
		SkipAuth:    true,
		Handler: mod.Handler{
			Func:       handleErrorTest,
			InputType:  reflect.TypeOf(ErrorTestRequest{}),
			OutputType: reflect.TypeOf(ErrorTestResponse{}),
		},
	}

	err = app.Register(errorService)
	if err != nil {
		panic(fmt.Sprintf("注册错误测试服务失败: %v", err))
	}

	// 4. 启动Web界面路由
	app.Get("/", func(c *fiber.Ctx) error {
		return handleIndexPage(c)
	})
	app.Get("/test", func(c *fiber.Ctx) error {
		return handleTestPage(c)
	})

	fmt.Println("文件日志测试服务器启动...")
	fmt.Println("访问 http://localhost:8081 查看测试界面")
	fmt.Println("访问 http://localhost:8081/test 进行日志测试")
	fmt.Println("访问 http://localhost:8081/services/docs 查看API文档")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("- POST /services/log-test    - 日志测试服务")
	fmt.Println("- POST /services/error-test  - 错误测试服务")
	fmt.Println()
	fmt.Println("日志文件位置: ./logs/app.log")

	// 5. 启动服务器
	app.Run()
}

// LogTestRequest 日志测试请求结构
type LogTestRequest struct {
	Message string `json:"message" validate:"required" desc:"要记录的日志消息"`
	Level   string `json:"level" validate:"required,oneof=debug info warn error" desc:"日志级别: debug, info, warn, error"`
}

// LogTestResponse 日志测试响应结构
type LogTestResponse struct {
	Status    string    `json:"status" desc:"处理状态"`
	Timestamp time.Time `json:"timestamp" desc:"处理时间"`
	Message   string    `json:"message" desc:"响应消息"`
}

// ErrorTestRequest 错误测试请求结构
type ErrorTestRequest struct {
	ErrorType string `json:"error_type" validate:"required,oneof=panic runtime business" desc:"错误类型: panic, runtime, business"`
	Message   string `json:"message" desc:"错误消息"`
}

// ErrorTestResponse 错误测试响应结构
type ErrorTestResponse struct {
	ErrorHandled bool   `json:"error_handled" desc:"错误是否被处理"`
	Message      string `json:"message" desc:"响应消息"`
}

// handleLogTest 处理日志测试
func handleLogTest(ctx *mod.Context, in interface{}, out interface{}) error {
	req := in.(*LogTestRequest)
	resp := out.(*LogTestResponse)

	// 根据请求的日志级别记录日志
	logger := ctx.GetLogger()
	switch req.Level {
	case "debug":
		logger.WithField("service", "log-test").Debug(req.Message)
	case "info":
		logger.WithField("service", "log-test").Info(req.Message)
	case "warn":
		logger.WithField("service", "log-test").Warn(req.Message)
	case "error":
		logger.WithField("service", "log-test").Error(req.Message)
	}

	// 记录请求信息
	logger.WithFields(map[string]interface{}{
		"request_id": ctx.GetRequestID(),
		"level":      req.Level,
		"message":    req.Message,
		"user_agent": ctx.Get("User-Agent"),
		"ip":         ctx.IP(),
	}).Info("Log test request processed")

	resp.Status = "success"
	resp.Timestamp = time.Now()
	resp.Message = fmt.Sprintf("已记录 %s 级别日志: %s", req.Level, req.Message)

	return nil
}

// handleErrorTest 处理错误测试
func handleErrorTest(ctx *mod.Context, in interface{}, out interface{}) error {
	req := in.(*ErrorTestRequest)
	resp := out.(*ErrorTestResponse)

	logger := ctx.GetLogger()

	switch req.ErrorType {
	case "panic":
		// 记录即将发生的panic
		logger.WithFields(map[string]interface{}{
			"error_type": req.ErrorType,
			"message":    req.Message,
			"request_id": ctx.GetRequestID(),
		}).Error("About to trigger panic for testing")

		// 这会导致panic，但应该被框架捕获
		panic(fmt.Sprintf("测试panic: %s", req.Message))

	case "runtime":
		// 模拟运行时错误
		logger.WithFields(map[string]interface{}{
			"error_type": req.ErrorType,
			"message":    req.Message,
			"request_id": ctx.GetRequestID(),
		}).Error("Runtime error occurred")

		return fmt.Errorf("运行时错误: %s", req.Message)

	case "business":
		// 模拟业务逻辑错误
		logger.WithFields(map[string]interface{}{
			"error_type": req.ErrorType,
			"message":    req.Message,
			"request_id": ctx.GetRequestID(),
		}).Warn("Business logic error")

		return mod.ReplyWithDetail(400, "业务逻辑错误", req.Message)
	}

	resp.ErrorHandled = true
	resp.Message = fmt.Sprintf("错误类型 %s 已处理", req.ErrorType)

	return nil
}

// handleIndexPage 首页
func handleIndexPage(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">
    <title>MOD Framework 文件日志测试</title>
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
        .log-info { background: #e6f3ff; padding: 15px; border-radius: 5px; border-left: 4px solid #007bff; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>📋 MOD Framework 文件日志测试</h1>

        <div class="section">
            <h3>📝 文件日志功能</h3>
            <p>测试MOD框架的文件日志记录功能，包括不同级别的日志记录和日志轮转。</p>
            <div class="links">
                <a href="/test" class="link">🧪 日志测试</a>
                <a href="/services/log-test" class="link api-link">📊 日志API</a>
                <a href="/services/error-test" class="link api-link">❌ 错误测试API</a>
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
            <div class="note">
                <strong>注意：</strong>使用前请确保 mod.yml 配置文件中的日志配置正确：
                <div class="code">
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
                </div>
            </div>
        </div>

        <div class="section">
            <h3>📁 日志文件信息</h3>
            <div class="log-info">
                <strong>日志文件位置：</strong><br>
                <div class="code">./logs/app.log</div>

                <strong>特性：</strong><br>
                • 🔄 自动日志轮转（大小限制）<br>
                • 📦 历史日志压缩<br>
                • 🗓️ 按时间清理旧日志<br>
                • 📊 JSON格式（文件）+ 文本格式（控制台）<br>
                • 🏷️ 结构化字段记录
            </div>
        </div>

        <div class="section">
            <h3>🧪 测试指南</h3>
            <p><strong>1. 日志级别测试：</strong></p>
            <div class="code">
curl -X POST http://localhost:8080/services/log-test \\
  -H "Content-Type: application/json" \\
  -d '{"message": "测试信息日志", "level": "info"}'
            </div>

            <p><strong>2. 错误测试：</strong></p>
            <div class="code">
curl -X POST http://localhost:8080/services/error-test \\
  -H "Content-Type: application/json" \\
  -d '{"error_type": "business", "message": "测试业务错误"}'
            </div>

            <p><strong>3. 查看日志文件：</strong></p>
            <div class="code">
tail -f ./logs/app.log
            </div>
        </div>
    </div>
</body>
</html>`
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

// handleTestPage 测试页面
func handleTestPage(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>文件日志测试</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .form-group { margin: 20px 0; }
        label { display: block; margin-bottom: 5px; font-weight: bold; color: #555; }
        input, select { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px; }
        button { background: #007bff; color: white; padding: 12px 30px; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; margin: 10px 5px; }
        button:hover { background: #0056b3; }
        .error-btn { background: #dc3545; }
        .error-btn:hover { background: #c82333; }
        .result { margin: 20px 0; padding: 15px; border-radius: 5px; display: none; }
        .success { background: #d4edda; border: 1px solid #c3e6cb; color: #155724; }
        .error { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; }
        .loading { text-align: center; color: #666; }
        pre { background: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; }
        .back-link { display: inline-block; margin-bottom: 20px; color: #007bff; text-decoration: none; }
        .back-link:hover { text-decoration: underline; }
        .log-note { background: #e6f3ff; padding: 15px; border-radius: 5px; border-left: 4px solid #007bff; margin: 10px 0; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <a href="/" class="back-link">← 返回首页</a>
        <h1>📋 文件日志测试</h1>

        <div class="log-note">
            <strong>💡 提示：</strong>所有测试操作都会同时记录到控制台和文件 <code>./logs/app.log</code> 中
        </div>

        <div class="form-group">
            <label for="logMessage">日志消息：</label>
            <input type="text" id="logMessage" placeholder="输入要记录的日志消息" value="这是一条测试日志消息">
        </div>

        <div class="form-group">
            <label for="logLevel">日志级别：</label>
            <select id="logLevel">
                <option value="debug">Debug - 调试信息</option>
                <option value="info" selected>Info - 一般信息</option>
                <option value="warn">Warn - 警告信息</option>
                <option value="error">Error - 错误信息</option>
            </select>
            <button onclick="testLog()">记录日志</button>
        </div>

        <div class="form-group">
            <label for="errorType">错误类型：</label>
            <select id="errorType">
                <option value="business" selected>Business - 业务逻辑错误</option>
                <option value="runtime">Runtime - 运行时错误</option>
                <option value="panic">Panic - 系统级错误</option>
            </select>
            <input type="text" id="errorMessage" placeholder="错误消息" value="这是一条测试错误消息" style="margin-top: 10px;">
            <button onclick="testError()" class="error-btn">触发错误</button>
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
            showResult('<div class="loading">⏳ 处理中...</div>', true);
        }

        async function testLog() {
            const message = document.getElementById('logMessage').value;
            const level = document.getElementById('logLevel').value;

            if (!message.trim()) {
                showResult('❌ 请输入日志消息', false);
                return;
            }

            showLoading();

            try {
                const response = await fetch('/services/log-test', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        message: message,
                        level: level
                    })
                });

                const result = await response.json();

                if (response.ok && result.code === 0) {
                    showResult(
                        '✅ 日志记录成功！<br>' +
                        '<strong>状态：</strong> ' + result.data.status + '<br>' +
                        '<strong>时间：</strong> ' + result.data.timestamp + '<br>' +
                        '<strong>消息：</strong> ' + result.data.message + '<br>' +
                        '<strong>💾 检查文件：</strong> ./logs/app.log',
                        true
                    );
                } else {
                    showResult(
                        '❌ 日志记录失败：<br>' +
                        '<pre>' + JSON.stringify(result, null, 2) + '</pre>',
                        false
                    );
                }
            } catch (error) {
                showResult('❌ 网络错误：' + error.message, false);
            }
        }

        async function testError() {
            const errorType = document.getElementById('errorType').value;
            const message = document.getElementById('errorMessage').value;

            showLoading();

            try {
                const response = await fetch('/services/error-test', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        error_type: errorType,
                        message: message
                    })
                });

                const result = await response.json();

                if (response.ok && result.code === 0) {
                    showResult(
                        '✅ 错误测试完成！<br>' +
                        '<strong>错误已处理：</strong> ' + result.data.error_handled + '<br>' +
                        '<strong>消息：</strong> ' + result.data.message + '<br>' +
                        '<strong>💾 检查错误日志：</strong> ./logs/app.log',
                        true
                    );
                } else {
                    showResult(
                        '❌ 错误测试结果：<br>' +
                        '<strong>类型：</strong> ' + errorType + '<br>' +
                        '<strong>响应：</strong><br>' +
                        '<pre>' + JSON.stringify(result, null, 2) + '</pre>' +
                        '<br><strong>💾 检查错误日志：</strong> ./logs/app.log',
                        false
                    );
                }
            } catch (error) {
                showResult('❌ 网络错误：' + error.message + '<br><strong>💾 检查错误日志：</strong> ./logs/app.log', false);
            }
        }
    </script>
</body>
</html>`
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}
