package mod

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"reflect"
)

type Context struct {
	*fiber.Ctx
	RequestID string
	logger    *logrus.Logger
}

func (c *Context) GetRequestID() string {
	if c.RequestID == "" {
		c.RequestID = NextSnowflakeStringID()
	}
	return c.RequestID
}

// GetLogger returns the logger instance
func (c *Context) GetLogger() *logrus.Logger {
	return c.logger
}

// Logger methods with automatic rid inclusion
func (c *Context) Debug(args ...interface{}) {
	if c.logger != nil {
		c.logger.WithField("rid", c.GetRequestID()).Debug(args...)
	}
}

func (c *Context) Debugf(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger.WithField("rid", c.GetRequestID()).Debugf(format, args...)
	}
}

func (c *Context) Info(args ...interface{}) {
	if c.logger != nil {
		c.logger.WithField("rid", c.GetRequestID()).Info(args...)
	}
}

func (c *Context) Infof(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger.WithField("rid", c.GetRequestID()).Infof(format, args...)
	}
}

func (c *Context) Warn(args ...interface{}) {
	if c.logger != nil {
		c.logger.WithField("rid", c.GetRequestID()).Warn(args...)
	}
}

func (c *Context) Warnf(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger.WithField("rid", c.GetRequestID()).Warnf(format, args...)
	}
}

func (c *Context) Error(args ...interface{}) {
	if c.logger != nil {
		c.logger.WithField("rid", c.GetRequestID()).Error(args...)
	}
}

func (c *Context) Errorf(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger.WithField("rid", c.GetRequestID()).Errorf(format, args...)
	}
}

func (c *Context) WithFields(fields logrus.Fields) *logrus.Entry {
	if c.logger != nil {
		if fields == nil {
			fields = logrus.Fields{}
		}
		fields["rid"] = c.GetRequestID()
		return c.logger.WithFields(fields)
	}
	return nil
}

// Handler 结构体，可以存储类型信息
type Handler struct {
	Func       func(ctx *Context, args, reply any) error
	InputType  reflect.Type
	OutputType reflect.Type
}

type Service struct {
	Name        string  `validate:"required"`
	DisplayName string  `validate:"required"`
	Handler     Handler `validate:"required"`

	Description string
	SkipAuth    bool
	ReturnRaw   bool
	Group       string // 在文档中的分组
	Sort        int    // 在文档中的排序值，从小到大排列
}

// MakeHandler 创建带类型信息的 Handler
func MakeHandler[I any, O any](handler func(ctx *Context, args *I, reply *O) error) Handler {
	return Handler{
		Func: func(ctx *Context, args any, reply any) error {
			a, ok := args.(*I)
			if !ok {
				return fmt.Errorf("invalid args type")
			}
			r, ok := reply.(*O)
			if !ok {
				return fmt.Errorf("invalid reply type")
			}
			return handler(ctx, a, r)
		},
		InputType:  reflect.TypeOf((*I)(nil)).Elem(),
		OutputType: reflect.TypeOf((*O)(nil)).Elem(),
	}
}

type StdReply struct {
	code   int
	msg    string
	detail string
}

func (r StdReply) Error() string {
	return fmt.Sprintf("%s (%d)", r.msg, r.code)
}

func (r StdReply) Code() int {
	return r.code
}

func (r StdReply) Msg() string {
	return r.msg
}

func (r StdReply) Detail() string {
	return r.detail
}

func Reply(code int, msg string) error {
	return &StdReply{code: code, msg: msg}
}

func ReplyWithDetail(code int, msg, detail string) error {
	return &StdReply{code: code, msg: msg, detail: detail}
}

// 统一响应格式
type ApiResponse struct {
	Code   int         `json:"code"`
	Data   interface{} `json:"data,omitempty"`
	Msg    string      `json:"msg"`
	Detail string      `json:"detail,omitempty"`
	Rid    string      `json:"rid"`
}

// 生成成功响应
func NewSuccessResponse(ctx *Context, data interface{}) *ApiResponse {
	return &ApiResponse{
		Code: 0,
		Data: data,
		Msg:  "success",
		Rid:  ctx.GetRequestID(),
	}
}

// 生成错误响应
func NewErrorResponse(ctx *Context, code int, msg string, detail ...string) *ApiResponse {
	resp := &ApiResponse{
		Code: code,
		Msg:  msg,
		Rid:  ctx.GetRequestID(),
	}
	if len(detail) > 0 && detail[0] != "" {
		resp.Detail = detail[0]
	}
	return resp
}

// JWT related methods

// GetJWTClaims returns the JWT claims from the context
func (c *Context) GetJWTClaims() *JWTClaims {
	if claims, ok := c.Locals("jwt_claims").(*JWTClaims); ok {
		return claims
	}
	return nil
}

// GetJWTToken returns the JWT token string from the context
func (c *Context) GetJWTToken() string {
	if token, ok := c.Locals("jwt_token").(string); ok {
		return token
	}
	return ""
}

// GetUserID returns the user ID from JWT claims
func (c *Context) GetUserID() string {
	if userID, ok := c.Locals("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetUsername returns the username from JWT claims
func (c *Context) GetUsername() string {
	if username, ok := c.Locals("username").(string); ok {
		return username
	}
	return ""
}

// GetUserEmail returns the user email from JWT claims
func (c *Context) GetUserEmail() string {
	if email, ok := c.Locals("user_email").(string); ok {
		return email
	}
	return ""
}

// GetUserRole returns the user role from JWT claims
func (c *Context) GetUserRole() string {
	if role, ok := c.Locals("user_role").(string); ok {
		return role
	}
	return ""
}

// IsAuthenticated checks if the request has valid JWT authentication
func (c *Context) IsAuthenticated() bool {
	return c.GetJWTClaims() != nil
}

// HasRole checks if the current user has the specified role
func (c *Context) HasRole(role string) bool {
	return c.GetUserRole() == role
}

// HasAnyRole checks if the current user has any of the specified roles
func (c *Context) HasAnyRole(roles ...string) bool {
	userRole := c.GetUserRole()
	for _, role := range roles {
		if userRole == role {
			return true
		}
	}
	return false
}
