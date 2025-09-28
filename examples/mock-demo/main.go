package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// 用户信息请求
type UserInfoRequest struct {
	UserID string `json:"user_id" validate:"required" desc:"用户ID"`
}

// 用户信息响应
type UserInfoResponse struct {
	ID       string    `json:"id" desc:"用户ID"`
	Name     string    `json:"name" desc:"用户名"`
	Email    string    `json:"email" desc:"邮箱地址"`
	Phone    string    `json:"phone" desc:"手机号码"`
	Address  string    `json:"address" desc:"地址"`
	Status   string    `json:"status" desc:"用户状态"`
	CreateAt time.Time `json:"create_at" desc:"创建时间"`
}

// 用户列表请求
type UserListRequest struct {
	Page     int    `json:"page" validate:"min=1" desc:"页码"`
	PageSize int    `json:"page_size" validate:"min=1,max=100" desc:"每页数量"`
	Keyword  string `json:"keyword" desc:"搜索关键词"`
}

// 用户列表响应
type UserListResponse struct {
	Users []UserInfoResponse `json:"users" desc:"用户列表"`
	Total int                `json:"total" desc:"总数量"`
	Page  int                `json:"page" desc:"当前页码"`
}

// 订单信息请求
type OrderInfoRequest struct {
	OrderID string `json:"order_id" validate:"required" desc:"订单ID"`
}

// 订单信息响应
type OrderInfoResponse struct {
	ID          string                 `json:"id" desc:"订单ID"`
	UserID      string                 `json:"user_id" desc:"用户ID"`
	ProductName string                 `json:"product_name" desc:"商品名称"`
	Amount      float64                `json:"amount" desc:"订单金额"`
	Status      string                 `json:"status" desc:"订单状态"`
	Items       []OrderItem            `json:"items" desc:"订单项"`
	Metadata    map[string]interface{} `json:"metadata" desc:"元数据"`
	CreateAt    time.Time              `json:"create_at" desc:"创建时间"`
}

// 订单项
type OrderItem struct {
	ProductID string  `json:"product_id" desc:"商品ID"`
	Name      string  `json:"name" desc:"商品名称"`
	Price     float64 `json:"price" desc:"单价"`
	Quantity  int     `json:"quantity" desc:"数量"`
}

// 消息发送请求
type SendMessageRequest struct {
	ToUserID string `json:"to_user_id" validate:"required" desc:"接收用户ID"`
	Message  string `json:"message" validate:"required" desc:"消息内容"`
	Type     string `json:"type" desc:"消息类型"`
}

// 消息发送响应
type SendMessageResponse struct {
	MessageID string    `json:"message_id" desc:"消息ID"`
	Status    string    `json:"status" desc:"发送状态"`
	SentAt    time.Time `json:"sent_at" desc:"发送时间"`
}

func main() {
	// 创建MOD应用实例
	app := mod.New()

	// 注册用户相关服务（用户管理分组）
	registerUserServices(app)

	// 注册订单相关服务（订单管理分组）
	registerOrderServices(app)

	// 注册消息相关服务（消息服务分组）
	registerMessageServices(app)

	// 添加首页路由
	app.Get("/", func(c *fiber.Ctx) error {
		return handleIndexPage(c)
	})

	// 添加Mock状态查看路由
	app.Get("/mock-status", func(c *fiber.Ctx) error {
		return handleMockStatusPage(c)
	})

	fmt.Println("Mock功能演示服务器启动...")
	fmt.Println("访问 http://localhost:8080 查看演示界面")
	fmt.Println("访问 http://localhost:8080/mock-status 查看Mock状态")
	fmt.Println("访问 http://localhost:8080/services/docs 查看API文档")
	fmt.Println()
	fmt.Println("服务说明:")
	fmt.Println("- 用户管理组: 启用分组级Mock（所有用户相关服务）")
	fmt.Println("- 订单管理组: 部分服务启用Mock")
	fmt.Println("- 消息服务组: 不启用Mock（使用实际Handler）")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("- POST /services/user-info     - 获取用户信息（Mock）")
	fmt.Println("- POST /services/user-list     - 获取用户列表（Mock）")
	fmt.Println("- POST /services/order-info    - 获取订单信息（Mock）")
	fmt.Println("- POST /services/send-message  - 发送消息（实际）")

	// 启动服务器
	app.Run()
}

// 注册用户相关服务
func registerUserServices(app *mod.App) {
	// 获取用户信息服务
	app.Register(mod.Service{
		Name:        "user-info",
		DisplayName: "获取用户信息",
		Description: "根据用户ID获取用户详细信息",
		Group:       "用户管理",
		Sort:        1,
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *UserInfoRequest, resp *UserInfoResponse) error {
			// 实际业务逻辑（Mock模式下不会执行）
			resp.ID = req.UserID
			resp.Name = "实际用户名"
			resp.Email = "real@example.com"
			resp.Phone = "13800138000"
			resp.Address = "实际地址"
			resp.Status = "active"
			resp.CreateAt = time.Now()

			ctx.GetLogger().WithFields(map[string]interface{}{
				"user_id": req.UserID,
			}).Info("实际获取用户信息")

			return nil
		}),
	})

	// 获取用户列表服务
	app.Register(mod.Service{
		Name:        "user-list",
		DisplayName: "获取用户列表",
		Description: "分页获取用户列表",
		Group:       "用户管理",
		Sort:        2,
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *UserListRequest, resp *UserListResponse) error {
			// 实际业务逻辑（Mock模式下不会执行）
			resp.Users = []UserInfoResponse{
				{
					ID:       "real_user_1",
					Name:     "实际用户1",
					Email:    "user1@real.com",
					Phone:    "13800138001",
					Address:  "实际地址1",
					Status:   "active",
					CreateAt: time.Now(),
				},
			}
			resp.Total = 1
			resp.Page = req.Page

			ctx.GetLogger().WithFields(map[string]interface{}{
				"page":      req.Page,
				"page_size": req.PageSize,
			}).Info("实际获取用户列表")

			return nil
		}),
	})
}

// 注册订单相关服务
func registerOrderServices(app *mod.App) {
	// 获取订单信息服务（单独启用Mock）
	app.Register(mod.Service{
		Name:        "order-info",
		DisplayName: "获取订单信息",
		Description: "根据订单ID获取订单详细信息",
		Group:       "订单管理",
		Sort:        1,
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *OrderInfoRequest, resp *OrderInfoResponse) error {
			// 实际业务逻辑（Mock模式下不会执行）
			resp.ID = req.OrderID
			resp.UserID = "real_user_123"
			resp.ProductName = "实际商品"
			resp.Amount = 99.99
			resp.Status = "paid"
			resp.Items = []OrderItem{
				{
					ProductID: "prod_real_1",
					Name:      "实际商品1",
					Price:     49.99,
					Quantity:  2,
				},
			}
			resp.Metadata = map[string]interface{}{
				"payment_method": "credit_card",
				"shipping":       "express",
			}
			resp.CreateAt = time.Now()

			ctx.GetLogger().WithFields(map[string]interface{}{
				"order_id": req.OrderID,
			}).Info("实际获取订单信息")

			return nil
		}),
	})
}

// 注册消息相关服务
func registerMessageServices(app *mod.App) {
	// 发送消息服务（不启用Mock）
	app.Register(mod.Service{
		Name:        "send-message",
		DisplayName: "发送消息",
		Description: "向指定用户发送消息",
		Group:       "消息服务",
		Sort:        1,
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *SendMessageRequest, resp *SendMessageResponse) error {
			// 实际业务逻辑
			resp.MessageID = fmt.Sprintf("msg_%d", time.Now().Unix())
			resp.Status = "sent"
			resp.SentAt = time.Now()

			ctx.GetLogger().WithFields(map[string]interface{}{
				"to_user_id": req.ToUserID,
				"message":    req.Message,
				"type":       req.Type,
				"message_id": resp.MessageID,
			}).Info("消息发送成功")

			return nil
		}),
	})
}

// 处理首页
func handleIndexPage(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MOD Framework Mock功能演示</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 1000px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; margin-bottom: 30px; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px; background-color: #fafafa; }
        .section h3 { margin-top: 0; color: #666; }
        .api-group { margin: 15px 0; }
        .api-item { margin: 10px 0; padding: 10px; background: white; border-radius: 5px; border-left: 4px solid #007bff; }
        .mock-enabled { border-left-color: #28a745; background-color: #f8fff9; }
        .mock-disabled { border-left-color: #dc3545; background-color: #fff8f8; }
        .test-button { background: #007bff; color: white; border: none; padding: 8px 15px; border-radius: 3px; cursor: pointer; margin: 5px; }
        .test-button:hover { background: #0056b3; }
        .response { margin: 10px 0; padding: 10px; background: #f8f9fa; border-radius: 5px; display: none; }
        .note { background: #fff3cd; padding: 15px; border-radius: 5px; border-left: 4px solid #ffc107; margin: 15px 0; }
        pre { background: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎭 MOD Framework Mock功能演示</h1>

        <div class="section">
            <h3>📋 Mock配置说明</h3>
            <div class="note">
                本演示展示了MOD框架的三级Mock配置：<br>
                <strong>1. 全局级别</strong>: 关闭（可在mod.yml中配置）<br>
                <strong>2. 分组级别</strong>: "用户管理"分组启用Mock<br>
                <strong>3. 服务级别</strong>: "order-info"服务单独启用Mock
            </div>
        </div>

        <div class="section">
            <h3>🧪 API测试</h3>

            <div class="api-group">
                <h4>👥 用户管理组 (分组级Mock: 启用)</h4>

                <div class="api-item mock-enabled">
                    <strong>获取用户信息</strong> - <span style="color: #28a745;">Mock启用</span><br>
                    <small>POST /services/user-info</small><br>
                    <button class="test-button" onclick="testAPI('user-info', {user_id: 'test_123'})">测试API</button>
                    <div id="response-user-info" class="response"></div>
                </div>

                <div class="api-item mock-enabled">
                    <strong>获取用户列表</strong> - <span style="color: #28a745;">Mock启用</span><br>
                    <small>POST /services/user-list</small><br>
                    <button class="test-button" onclick="testAPI('user-list', {page: 1, page_size: 10, keyword: 'test'})">测试API</button>
                    <div id="response-user-list" class="response"></div>
                </div>
            </div>

            <div class="api-group">
                <h4>📦 订单管理组 (分组级Mock: 关闭，服务级Mock: 部分启用)</h4>

                <div class="api-item mock-enabled">
                    <strong>获取订单信息</strong> - <span style="color: #28a745;">Mock启用</span> (服务级配置)<br>
                    <small>POST /services/order-info</small><br>
                    <button class="test-button" onclick="testAPI('order-info', {order_id: 'order_456'})">测试API</button>
                    <div id="response-order-info" class="response"></div>
                </div>
            </div>

            <div class="api-group">
                <h4>💬 消息服务组 (所有Mock: 关闭)</h4>

                <div class="api-item mock-disabled">
                    <strong>发送消息</strong> - <span style="color: #dc3545;">Mock关闭</span> (使用实际Handler)<br>
                    <small>POST /services/send-message</small><br>
                    <button class="test-button" onclick="testAPI('send-message', {to_user_id: 'user_789', message: '测试消息', type: 'text'})">测试API</button>
                    <div id="response-send-message" class="response"></div>
                </div>
            </div>
        </div>

        <div class="section">
            <h3>🔗 相关链接</h3>
            <p>
                <a href="/mock-status" target="_blank">🔍 查看Mock状态</a> |
                <a href="/services/docs" target="_blank">📖 API文档</a>
            </p>
        </div>
    </div>

    <script>
        async function testAPI(serviceName, params) {
            const responseDiv = document.getElementById('response-' + serviceName);
            responseDiv.style.display = 'block';
            responseDiv.innerHTML = '<p>⏳ 请求中...</p>';

            try {
                const response = await fetch('/services/' + serviceName, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(params)
                });

                const result = await response.json();

                let statusColor = response.ok ? '#28a745' : '#dc3545';
                responseDiv.innerHTML =
                    '<p><strong>状态:</strong> <span style="color: ' + statusColor + '">' + response.status + '</span></p>' +
                    '<p><strong>响应:</strong></p>' +
                    '<pre>' + JSON.stringify(result, null, 2) + '</pre>';

            } catch (error) {
                responseDiv.innerHTML = '<p style="color: #dc3545;">❌ 请求失败: ' + error.message + '</p>';
            }
        }
    </script>
</body>
</html>`

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

// 处理Mock状态页面
func handleMockStatusPage(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mock状态查看</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; margin-bottom: 30px; }
        .config-section { margin: 20px 0; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px; background-color: #fafafa; }
        .config-item { margin: 10px 0; padding: 10px; background: white; border-radius: 5px; }
        .enabled { border-left: 4px solid #28a745; background-color: #f8fff9; }
        .disabled { border-left: 4px solid #dc3545; background-color: #fff8f8; }
        .back-link { display: inline-block; margin-bottom: 20px; color: #007bff; text-decoration: none; }
        .back-link:hover { text-decoration: underline; }
        pre { background: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <a href="/" class="back-link">← 返回首页</a>
        <h1>🔍 Mock配置状态</h1>

        <div class="config-section">
            <h3>📋 当前Mock配置 (mod.yml)</h3>
            <pre>mock:
  global:
    enabled: false
  groups:
    "用户管理":
      enabled: true
  services:
    "order-info":
      enabled: true</pre>
        </div>

        <div class="config-section">
            <h3>🎯 服务Mock状态</h3>

            <div class="config-item enabled">
                <strong>user-info</strong> (用户管理组)<br>
                <small>Mock状态: 启用 - 分组级配置</small>
            </div>

            <div class="config-item enabled">
                <strong>user-list</strong> (用户管理组)<br>
                <small>Mock状态: 启用 - 分组级配置</small>
            </div>

            <div class="config-item enabled">
                <strong>order-info</strong> (订单管理组)<br>
                <small>Mock状态: 启用 - 服务级配置</small>
            </div>

            <div class="config-item disabled">
                <strong>send-message</strong> (消息服务组)<br>
                <small>Mock状态: 关闭 - 使用实际Handler</small>
            </div>
        </div>

        <div class="config-section">
            <h3>⚙️ 配置优先级</h3>
            <p>Mock配置的优先级从高到低为：</p>
            <ol>
                <li><strong>服务级配置</strong> - mock.services.{service_name}.enabled</li>
                <li><strong>分组级配置</strong> - mock.groups.{group_name}.enabled</li>
                <li><strong>全局级配置</strong> - mock.global.enabled</li>
            </ol>
        </div>
    </div>
</body>
</html>`

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}
