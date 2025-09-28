package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// 简单的静态文件服务测试
func main() {
	fmt.Println("🧪 静态文件挂载功能测试")
	fmt.Println("请确保先运行 examples/static_example.go 并使用 examples/static_test.yml 配置")
	fmt.Println()

	// 等待服务器启动
	fmt.Println("等待服务器启动...")
	time.Sleep(2 * time.Second)

	baseURL := "http://localhost:3000"

	// 测试用例
	testCases := []struct {
		name        string
		url         string
		expectCode  int
		description string
	}{
		{
			name:        "静态首页",
			url:         baseURL + "/static/",
			expectCode:  200,
			description: "访问静态资源目录的默认index.html文件",
		},
		{
			name:        "上传目录浏览",
			url:         baseURL + "/uploads/",
			expectCode:  200,
			description: "访问可浏览的uploads目录",
		},
		{
			name:        "文档首页",
			url:         baseURL + "/docs/",
			expectCode:  200,
			description: "访问文档目录的默认index.html文件",
		},
		{
			name:        "开发资源目录",
			url:         baseURL + "/dev/",
			expectCode:  200,
			description: "访问开发资源目录（使用README.md作为默认文件）",
		},
		{
			name:        "Mock数据",
			url:         baseURL + "/mock/",
			expectCode:  200,
			description: "访问mock目录的index.json文件",
		},
		{
			name:        "API端点",
			url:         baseURL + "/services/file-info",
			expectCode:  200,
			description: "访问file-info API端点",
		},
		{
			name:        "不存在的路径",
			url:         baseURL + "/nonexistent/",
			expectCode:  404,
			description: "访问不存在的路径应该返回404",
		},
	}

	// 执行测试
	passed := 0
	total := len(testCases)

	for i, tc := range testCases {
		fmt.Printf("[%d/%d] 测试: %s\n", i+1, total, tc.name)
		fmt.Printf("       URL: %s\n", tc.url)
		fmt.Printf("       描述: %s\n", tc.description)

		resp, err := http.Get(tc.url)
		if err != nil {
			fmt.Printf("       ❌ 失败: %v\n", err)
			fmt.Println()
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == tc.expectCode {
			fmt.Printf("       ✅ 成功: HTTP %d\n", resp.StatusCode)
			passed++
		} else {
			fmt.Printf("       ❌ 失败: 期望 HTTP %d, 得到 HTTP %d\n", tc.expectCode, resp.StatusCode)
		}

		// 显示内容类型
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			fmt.Printf("       Content-Type: %s\n", contentType)
		}

		fmt.Println()
	}

	// 显示测试结果
	fmt.Println("=" * 50)
	fmt.Printf("测试完成: %d/%d 通过\n", passed, total)
	if passed == total {
		fmt.Println("🎉 所有测试通过！静态文件挂载功能正常工作")
	} else {
		fmt.Printf("⚠️  %d 个测试失败，请检查配置和服务器状态\n", total-passed)
		os.Exit(1)
	}

	// 提供一些使用提示
	fmt.Println()
	fmt.Println("💡 使用提示:")
	fmt.Println("1. 在浏览器中访问 http://localhost:3000/static/ 查看静态首页")
	fmt.Println("2. 访问 http://localhost:3000/uploads/ 体验目录浏览功能")
	fmt.Println("3. 访问 http://localhost:3000/services/file-info 查看API响应")
	fmt.Println("4. 尝试上传文件到 uploads/ 目录，然后通过浏览器访问")
}
