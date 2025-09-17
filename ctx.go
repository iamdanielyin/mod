package mod

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"reflect"
)

type Context struct {
	*fiber.Ctx
	RequestID string
}

func (c *Context) GetRequestID() string {
	if c.RequestID == "" {
		c.RequestID = NextSnowflakeStringID()
	}
	return c.RequestID
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

type IntlError struct {
	code   int
	msg    string
	detail string
}

func (r IntlError) Error() string {
	return fmt.Sprintf("%s (%d)", r.msg, r.code)
}

func (r IntlError) Code() int {
	return r.code
}

func (r IntlError) Msg() string {
	return r.msg
}

func (r IntlError) Detail() string {
	return r.detail
}

func Reply(code int, msg string) error {
	return &IntlError{code: code, msg: msg}
}

func ReplyWithDetail(code int, msg, detail string) error {
	return &IntlError{code: code, msg: msg, detail: detail}
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