package main

import (
	"github.com/iamdanielyin/mod"
	"github.com/iamdanielyin/mod/examples/types"
)

func main() {
	app := mod.New() // 使用默认配置

	// 注册复杂业务服务
	RegisterComplexServices(app)

	app.Run(":3001")
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
