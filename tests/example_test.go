package mod_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iamdanielyin/mod"
	"github.com/iamdanielyin/mod/examples/types"
)

func TestServiceRegistration(t *testing.T) {
	app := mod.New()

	// 创建服务
	app.Register(mod.Service{
		Name:        "test_login",
		DisplayName: "测试登录",
		SkipAuth:    true,
		Handler: mod.MakeHandler(func(c *mod.Context, args *types.LoginArgs, reply *types.LoginReply) error {
			if args.Username == "" || args.Password == "" {
				return mod.Reply(400, "用户名和密码不能为空")
			}

			reply.Uid = "user123"
			reply.Token = "token456"
			return nil
		}),
	})

	// 测试 JSON body 请求
	t.Run("JSON body request", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"username": "testuser",
			"password": "testpass",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/services/test_login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// 测试 Query 参数请求
	t.Run("Query parameters request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/test_login?username=testuser&password=testpass&token=querytoken", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}