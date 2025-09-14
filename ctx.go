package mod

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	*fiber.Ctx
}

type Service struct {
	Name        string                              `validate:"required"`
	DisplayName string                              `validate:"required"`
	Handler     func(c *Context, in, out any) error `validate:"required"`

	Description string
	SkipAuth    bool
	ReturnRaw   bool
}

type Handler func(ctx *Context, args, reply any) error

func Handle[I any, O any](handler func(ctx *Context, args *I, reply *O) error) Handler {
	return func(ctx *Context, args any, reply any) error {
		a, ok := args.(*I)
		if !ok {
			return fmt.Errorf("invalid args type")
		}
		r, ok := reply.(*O)
		if !ok {
			return fmt.Errorf("invalid reply type")
		}
		return handler(ctx, a, r)
	}
}

type IntlError struct {
	code int
	msg  string
}

func (r IntlError) Error() string {
	return fmt.Sprintf("%s (%d)", r.msg, r.code)
}

func Reply(code int, msg string) error {
	return &IntlError{code, msg}
}
