package main

import (
	"fmt"
	"github.com/iamdanielyin/mod"
	"github.com/iamdanielyin/mod/examples/types"
	"time"
)

func main() {
	app := mod.New() // 使用默认配置

	// 注册基础登录服务，用于生成Token
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

	// 注册 Token 测试服务
	RegisterTokenTestServices(app)

	app.Run(":3002")
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
