package mod_test

import (
	"errors"
	"github.com/iamdanielyin/mod"
	"testing"
)

type Args struct {
	Code string `mod:"from=body"`
}

type Reply struct {
	Uid   string
	Token string
}

func TestNew(t *testing.T) {
	app := mod.New()

	app.Register(mod.Service{
		Name:        "code2session",
		DisplayName: "微信自动登录",
		Handler: mod.Handle(func(c *mod.Context, args *Args, reply *Reply) error {
			code := args.Code
			if code == "" {
				return errors.New("参数异常")
			}
			reply.Token = "123"
			return nil
		}),
	})

	app.Run()
}
