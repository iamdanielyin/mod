package types

// 登录相关结构体
type LoginArgs struct {
	Username string `validate:"" desc:"用户名"`
	Password string `validate:"required" desc:"密码"`
	Token    string `mod:"from=query" desc:"令牌"`
}

type LoginReply struct {
	Uid   string `json:"uid" desc:"用户ID"`
	Token string `json:"token" desc:"访问令牌"`
}

// 用户相关结构体
type UserArgs struct {
	UserID string `validate:"required" desc:"用户ID"`
	Name   string `desc:"用户名"`
}

type UserReply struct {
	ID   string `json:"id" desc:"用户ID"`
	Name string `json:"name" desc:"用户名"`
	Role string `json:"role" desc:"用户角色"`
}

// Token 测试相关结构体
type TokenQueryArgs struct {
	Token string `validate:"required" desc:"要查询的Token"`
}

type TokenQueryReply struct {
	Valid   bool   `json:"valid" desc:"Token是否有效"`
	Message string `json:"message" desc:"查询结果消息"`
	Data    string `json:"data" desc:"Token关联的数据"`
}

type TokenLogoutArgs struct {
	Token string `validate:"required" desc:"要删除的Token"`
}

type TokenLogoutReply struct {
	Success bool   `json:"success" desc:"是否成功删除"`
	Message string `json:"message" desc:"操作结果消息"`
}

type TokenBatchTestArgs struct {
	Count int `validate:"min=1,max=1000" desc:"要创建的Token数量"`
}

type TokenBatchTestReply struct {
	TotalCreated int      `json:"total_created" desc:"成功创建的Token数量"`
	TotalErrors  int      `json:"total_errors" desc:"创建失败的Token数量"`
	Tokens       []string `json:"tokens" desc:"创建成功的Token列表"`
	Errors       []string `json:"errors" desc:"错误信息列表"`
	Message      string   `json:"message" desc:"批量测试结果消息"`
}

// 复杂嵌套结构示例

// Address 地址信息
type Address struct {
	Province string `json:"province" validate:"required" desc:"省份"`
	City     string `json:"city" validate:"required" desc:"城市"`
	District string `json:"district" desc:"区县"`
	Detail   string `json:"detail" validate:"required" desc:"详细地址"`
	Zipcode  string `json:"zipcode" desc:"邮政编码"`
}

// Contact 联系方式
type Contact struct {
	Phone string `json:"phone" validate:"required" desc:"手机号码"`
	Email string `json:"email" validate:"email" desc:"邮箱地址"`
	QQ    string `json:"qq" desc:"QQ号码"`
}

// Product 商品信息
type Product struct {
	ID          int64    `json:"id" validate:"required" desc:"商品ID"`
	Name        string   `json:"name" validate:"required" desc:"商品名称"`
	Price       float64  `json:"price" validate:"required,min=0" desc:"商品价格"`
	Description string   `json:"description" desc:"商品描述"`
	Category    string   `json:"category" validate:"required" desc:"商品分类"`
	Stock       int      `json:"stock" validate:"min=0" desc:"库存数量"`
	Images      []string `json:"images" desc:"商品图片列表"`
	Tags        []string `json:"tags" desc:"商品标签"`
}

// OrderItem 订单项
type OrderItem struct {
	Product  Product `json:"product" validate:"required" desc:"商品信息"`
	Quantity int     `json:"quantity" validate:"required,min=1" desc:"购买数量"`
	Price    float64 `json:"price" validate:"required,min=0" desc:"单价"`
	Subtotal float64 `json:"subtotal" desc:"小计"`
}

// User 用户信息
type User struct {
	ID       int64    `json:"id" desc:"用户ID"`
	Username string   `json:"username" validate:"required" desc:"用户名"`
	Nickname string   `json:"nickname" desc:"昵称"`
	Avatar   string   `json:"avatar" desc:"头像URL"`
	Contact  Contact  `json:"contact" validate:"required" desc:"联系方式"`
	Address  *Address `json:"address" desc:"地址信息"`
	Roles    []string `json:"roles" desc:"用户角色列表"`
}

// Pagination 分页信息
type Pagination struct {
	Page     int `json:"page" validate:"min=1" desc:"页码"`
	PageSize int `json:"page_size" validate:"min=1,max=100" desc:"每页数量"`
	Total    int `json:"total" desc:"总记录数"`
}

// OrderFilter 订单过滤条件
type OrderFilter struct {
	Status    string  `json:"status" mod:"from=query" desc:"订单状态"`
	StartTime string  `json:"start_time" mod:"from=query" desc:"开始时间"`
	EndTime   string  `json:"end_time" mod:"from=query" desc:"结束时间"`
	MinAmount float64 `json:"min_amount" mod:"from=query" desc:"最小金额"`
	MaxAmount float64 `json:"max_amount" mod:"from=query" desc:"最大金额"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID        int64       `json:"user_id" validate:"required" desc:"用户ID"`
	Items         []OrderItem `json:"items" validate:"required,dive" desc:"订单项列表"`
	ShipAddress   Address     `json:"ship_address" validate:"required" desc:"收货地址"`
	Remark        string      `json:"remark" desc:"订单备注"`
	CouponCode    string      `json:"coupon_code" desc:"优惠券代码"`
	PaymentMethod string      `json:"payment_method" validate:"required" desc:"支付方式"`
}

// Order 订单信息
type Order struct {
	ID             int64       `json:"id" desc:"订单ID"`
	OrderNo        string      `json:"order_no" desc:"订单号"`
	User           User        `json:"user" desc:"用户信息"`
	Items          []OrderItem `json:"items" desc:"订单项列表"`
	ShipAddress    Address     `json:"ship_address" desc:"收货地址"`
	TotalAmount    float64     `json:"total_amount" desc:"订单总金额"`
	ActualAmount   float64     `json:"actual_amount" desc:"实际支付金额"`
	DiscountAmount float64     `json:"discount_amount" desc:"优惠金额"`
	Status         string      `json:"status" desc:"订单状态"`
	PaymentMethod  string      `json:"payment_method" desc:"支付方式"`
	PaymentStatus  string      `json:"payment_status" desc:"支付状态"`
	Remark         string      `json:"remark" desc:"订单备注"`
	CreatedAt      string      `json:"created_at" desc:"创建时间"`
	UpdatedAt      string      `json:"updated_at" desc:"更新时间"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	Order   Order  `json:"order" desc:"订单信息"`
	PayURL  string `json:"pay_url" desc:"支付链接"`
	QRCode  string `json:"qr_code" desc:"支付二维码"`
	Message string `json:"message" desc:"提示信息"`
}

// GetOrderListRequest 获取订单列表请求
type GetOrderListRequest struct {
	UserID     int64       `json:"user_id" mod:"from=header,name=user-id" validate:"required" desc:"用户ID"`
	Filter     OrderFilter `json:"filter" desc:"过滤条件"`
	Pagination Pagination  `json:"pagination" validate:"required" desc:"分页参数"`
}

// GetOrderListResponse 获取订单列表响应
type GetOrderListResponse struct {
	Orders     []Order    `json:"orders" desc:"订单列表"`
	Pagination Pagination `json:"pagination" desc:"分页信息"`
	Summary    struct {
		TotalOrders int     `json:"total_orders" desc:"总订单数"`
		TotalAmount float64 `json:"total_amount" desc:"总金额"`
		PaidOrders  int     `json:"paid_orders" desc:"已支付订单数"`
		PaidAmount  float64 `json:"paid_amount" desc:"已支付金额"`
	} `json:"summary" desc:"统计信息"`
}

// GetUserProfileRequest 获取用户资料请求
type GetUserProfileRequest struct {
	UserID      int64    `mod:"from=param,name=id" validate:"required" desc:"用户ID"`
	IncludeAddr bool     `mod:"from=query" desc:"是否包含地址信息"`
	Fields      []string `mod:"from=query" desc:"指定返回字段"`
}

// UserProfile 用户详细资料
type UserProfile struct {
	User      User      `json:"user" desc:"基本信息"`
	Addresses []Address `json:"addresses" desc:"地址列表"`
	Orders    struct {
		Recent []Order `json:"recent" desc:"最近订单"`
		Stats  struct {
			TotalOrders int     `json:"total_orders" desc:"总订单数"`
			TotalAmount float64 `json:"total_amount" desc:"总消费金额"`
		} `json:"stats" desc:"订单统计"`
	} `json:"orders" desc:"订单相关信息"`
	Preferences map[string]interface{} `json:"preferences" desc:"用户偏好设置"`
}

// BatchUpdateProductsRequest 批量更新商品请求
type BatchUpdateProductsRequest struct {
	Products []struct {
		ID     int64                  `json:"id" validate:"required" desc:"商品ID"`
		Fields map[string]interface{} `json:"fields" validate:"required" desc:"要更新的字段"`
	} `json:"products" validate:"required,dive" desc:"商品更新列表"`
	UpdatedBy string `json:"updated_by" validate:"required" desc:"更新操作者"`
}

// BatchUpdateResult 批量更新结果
type BatchUpdateResult struct {
	Success []struct {
		ID      int64  `json:"id" desc:"商品ID"`
		Message string `json:"message" desc:"更新信息"`
	} `json:"success" desc:"成功更新的商品"`
	Failed []struct {
		ID    int64  `json:"id" desc:"商品ID"`
		Error string `json:"error" desc:"失败原因"`
	} `json:"failed" desc:"更新失败的商品"`
	Summary struct {
		Total        int `json:"total" desc:"总数"`
		SuccessCount int `json:"success_count" desc:"成功数量"`
		FailedCount  int `json:"failed_count" desc:"失败数量"`
	} `json:"summary" desc:"更新汇总"`
}
