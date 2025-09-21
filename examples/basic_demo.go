package main

import (
	"fmt"
	"github.com/iamdanielyin/mod"
	"github.com/iamdanielyin/mod/examples/types"
	
)

func main() {
	app := mod.New(mod.Config{
		Name:        "ComplexDemo",
		DisplayName: "复杂参数演示",
		Description: "演示复杂嵌套参数的API文档生成",
	})

	// 展示基本的服务注册
	app.Register(mod.Service{
		Name:        "basic_login",
		DisplayName: "基础登录",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			reply.Uid = "basic_user"
			reply.Token = "basic_token_" + fmt.Sprintf("%d", time.Now().Unix())

			// 测试 token 存储到缓存
			userData := map[string]interface{}{
				"user_id":   reply.Uid,
				"username":  args.Username,
				"login_at":  time.Now().Format(time.RFC3339),
				"user_role": "user",
			}

			// 将 token 存储到配置的缓存中
			if err := app.SetToken(reply.Token, userData); err != nil {
				c.Error("Failed to store token", err)
			} else {
				c.Info("Token stored successfully", map[string]interface{}{
					"token": reply.Token,
					"user":  args.Username,
				})
			}

			return nil
		}),
	})

	// 展示错误处理
	app.Register(mod.Service{
		Name:        "error_demo",
		DisplayName: "错误处理演示",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			if args.Username == "error" {
				return mod.ReplyWithDetail(400, "演示错误", "This is a demo error")
			}
			reply.Uid = "demo_user"
			reply.Token = "demo_token"
			return nil
		}),
	})

	// 展示不同的配置选项
	app.Register(mod.Service{
		Name:        "config_demo",
		DisplayName: "配置演示",
		SkipAuth:    false,
		ReturnRaw:   true,
		Description: "展示不同配置选项的服务",
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.UserArgs, reply *types.UserReply) error {
			reply.ID = args.UserID
			reply.Name = "Config Demo"
			reply.Role = "demo"
			return nil
		}),
	})

	// 注册复杂服务
	RegisterComplexServices(app)

	// 注册 Token 测试服务
	RegisterTokenTestServices(app)

	app.Run(":3000")
}

// RegisterComplexServices 注册复杂结构的服务示例
func RegisterComplexServices(app *mod.App) {
	// 创建订单服务
	app.Register(mod.Service{
		Name:        "create_order",
		DisplayName: "创建订单",
		Description: "创建新的订单，支持多商品、地址、优惠券等复杂业务逻辑",
		Group:       "订单管理",
		Sort:        10,
		Handler: mod.MakeHandler[types.CreateOrderRequest, types.CreateOrderResponse](func(ctx *mod.Context, req *types.CreateOrderRequest, resp *types.CreateOrderResponse) error {
			// 模拟业务逻辑
			*resp = types.CreateOrderResponse{
				Order: types.Order{
					ID:             12345,
					OrderNo:        "ORDER20240101001",
					User:           types.User{ID: req.UserID, Username: "testuser"},
					Items:          req.Items,
					ShipAddress:    req.ShipAddress,
					TotalAmount:    299.99,
					ActualAmount:   279.99,
					DiscountAmount: 20.00,
					Status:         "pending",
					PaymentMethod:  req.PaymentMethod,
					PaymentStatus:  "unpaid",
					Remark:         req.Remark,
					CreatedAt:      "2024-01-01T10:00:00Z",
					UpdatedAt:      "2024-01-01T10:00:00Z",
				},
				PayURL:  "https://pay.example.com/order/12345",
				QRCode:  "https://qr.example.com/order/12345.png",
				Message: "订单创建成功，请尽快完成支付",
			}
			return nil
		}),
	})

	// 获取订单列表服务
	app.Register(mod.Service{
		Name:        "get_order_list",
		DisplayName: "获取订单列表",
		Description: "根据条件查询订单列表，支持多维度过滤和分页",
		Group:       "订单管理",
		Sort:        20,
		Handler: mod.MakeHandler[types.GetOrderListRequest, types.GetOrderListResponse](func(ctx *mod.Context, req *types.GetOrderListRequest, resp *types.GetOrderListResponse) error {
			// 模拟返回数据
			*resp = types.GetOrderListResponse{
				Orders: []types.Order{
					{
						ID:          12345,
						OrderNo:     "ORDER20240101001",
						User:        types.User{ID: req.UserID, Username: "testuser"},
						TotalAmount: 299.99,
						Status:      "completed",
					},
				},
				Pagination: types.Pagination{
					Page:     req.Pagination.Page,
					PageSize: req.Pagination.PageSize,
					Total:    100,
				},
			}
			resp.Summary.TotalOrders = 100
			resp.Summary.TotalAmount = 29999.00
			resp.Summary.PaidOrders = 85
			resp.Summary.PaidAmount = 25499.15
			return nil
		}),
	})

	// 获取用户资料服务
	app.Register(mod.Service{
		Name:        "get_user_profile",
		DisplayName: "获取用户资料",
		Description: "获取用户详细资料，包括基本信息、地址、订单统计等",
		Group:       "用户管理",
		Sort:        10,
		Handler: mod.MakeHandler[types.GetUserProfileRequest, types.UserProfile](func(ctx *mod.Context, req *types.GetUserProfileRequest, resp *types.UserProfile) error {
			// 模拟返回用户资料
			resp.User = types.User{
				ID:       req.UserID,
				Username: "testuser",
				Nickname: "测试用户",
				Contact: types.Contact{
					Phone: "13800138000",
					Email: "test@example.com",
				},
			}

			if req.IncludeAddr {
				resp.Addresses = []types.Address{
					{
						Province: "广东省",
						City:     "深圳市",
						District: "南山区",
						Detail:   "科技园南区",
						Zipcode:  "518000",
					},
				}
			}

			resp.Orders.Stats.TotalOrders = 25
			resp.Orders.Stats.TotalAmount = 7499.75

			resp.Preferences = map[string]interface{}{
				"language": "zh-CN",
				"theme":    "dark",
				"notifications": map[string]bool{
					"email": true,
					"sms":   false,
				},
			}
			return nil
		}),
	})

	// 批量更新商品服务
	app.Register(mod.Service{
		Name:        "batch_update_products",
		DisplayName: "批量更新商品",
		Description: "批量更新商品信息，支持部分字段更新和错误处理",
		Group:       "商品管理",
		Sort:        30,
		Handler: mod.MakeHandler[types.BatchUpdateProductsRequest, types.BatchUpdateResult](func(ctx *mod.Context, req *types.BatchUpdateProductsRequest, resp *types.BatchUpdateResult) error {
			// 模拟批量更新结果
			resp.Summary.Total = len(req.Products)
			resp.Summary.SuccessCount = len(req.Products) - 1
			resp.Summary.FailedCount = 1

			for i, product := range req.Products {
				if i == 0 {
					// 模拟第一个商品更新失败
					resp.Failed = append(resp.Failed, struct {
						ID    int64  `json:"id" desc:"商品ID"`
						Error string `json:"error" desc:"失败原因"`
					}{
						ID:    product.ID,
						Error: "商品不存在",
					})
				} else {
					resp.Success = append(resp.Success, struct {
						ID      int64  `json:"id" desc:"商品ID"`
						Message string `json:"message" desc:"更新信息"`
					}{
						ID:      product.ID,
						Message: "更新成功",
					})
				}
			}
			return nil
		}),
	})
}

// RegisterTokenTestServices 注册 Token 测试服务
func RegisterTokenTestServices(app *mod.App) {
	// Token 验证测试服务（需要认证）
	app.Register(mod.Service{
		Name:        "token_verify_test",
		DisplayName: "Token验证测试",
		Description: "测试 Token 验证功能，此服务需要有效的 Token",
		Group:       "Token测试",
		Sort:        10,
		SkipAuth:    false, // 需要 Token 验证
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.UserArgs, reply *types.UserReply) error {
			reply.ID = args.UserID
			reply.Name = "Token验证成功"
			reply.Role = "authenticated_user"

			c.Info("Token validation successful", map[string]interface{}{
				"user_id": args.UserID,
				"message": "用户通过Token验证",
			})

			return nil
		}),
	})

	// Token 查询测试服务
	app.Register(mod.Service{
		Name:        "token_query_test",
		DisplayName: "Token查询测试",
		Description: "查询指定 Token 的详细信息",
		Group:       "Token测试",
		Sort:        20,
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.TokenQueryArgs, reply *types.TokenQueryReply) error {
			// 查询 Token 数据
			data, err := app.GetTokenData(args.Token)
			if err != nil {
				c.Warn("Token query failed", map[string]interface{}{
					"token": args.Token,
					"error": err.Error(),
				})
				reply.Valid = false
				reply.Message = "Token 不存在或已过期"
				return nil
			}

			reply.Valid = true
			reply.Message = "Token 查询成功"
			reply.Data = string(data)

			c.Info("Token query successful", map[string]interface{}{
				"token":       args.Token,
				"data_length": len(data),
			})

			return nil
		}),
	})

	// Token 删除测试服务
	app.Register(mod.Service{
		Name:        "token_logout_test",
		DisplayName: "Token登出测试",
		Description: "删除指定 Token，模拟用户登出",
		Group:       "Token测试",
		Sort:        30,
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.TokenLogoutArgs, reply *types.TokenLogoutReply) error {
			// 删除 Token
			err := app.RemoveToken(args.Token)
			if err != nil {
				c.Error("Token logout failed", err)
				reply.Success = false
				reply.Message = "Token 删除失败: " + err.Error()
				return nil
			}

			reply.Success = true
			reply.Message = "Token 删除成功，用户已登出"

			c.Info("Token logout successful", map[string]interface{}{
				"token": args.Token,
			})

			return nil
		}),
	})

	// 批量 Token 测试服务
	app.Register(mod.Service{
		Name:        "token_batch_test",
		DisplayName: "Token批量测试",
		Description: "批量创建和测试多个 Token，用于性能测试",
		Group:       "Token测试",
		Sort:        40,
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.TokenBatchTestArgs, reply *types.TokenBatchTestReply) error {
			count := args.Count
			if count <= 0 || count > 1000 {
				count = 10 // 默认创建 10 个
			}

			var tokens []string
			var errors []string

			// 批量创建 Token
			for i := 0; i < count; i++ {
				token := fmt.Sprintf("batch_token_%d_%d", time.Now().Unix(), i)
				userData := map[string]interface{}{
					"batch_id":  fmt.Sprintf("batch_%d", time.Now().Unix()),
					"token_id":  i,
					"username":  fmt.Sprintf("batch_user_%d", i),
					"create_at": time.Now().Format(time.RFC3339),
				}

				err := app.SetToken(token, userData)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Token %d: %v", i, err))
				} else {
					tokens = append(tokens, token)
				}
			}

			reply.TotalCreated = len(tokens)
			reply.TotalErrors = len(errors)
			reply.Tokens = tokens
			reply.Errors = errors
			reply.Message = fmt.Sprintf("批量创建完成: 成功 %d 个, 失败 %d 个", len(tokens), len(errors))

			c.Info("Batch token test completed", map[string]interface{}{
				"requested": count,
				"created":   len(tokens),
				"errors":    len(errors),
			})

			return nil
		}),
	})
}
