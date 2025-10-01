package main

import (
	"time"

	"github.com/iamdanielyin/mod"
)

// LogTestRequest represents log test request
type LogTestRequest struct {
	Level   string `json:"level" validate:"required" desc:"日志级别 (debug,info,warn,error)"`
	Message string `json:"message" validate:"required" desc:"日志消息"`
	Data    string `json:"data,omitempty" desc:"附加数据"`
}

// LogTestResponse represents log test response
type LogTestResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// LogInfoRequest represents log info request
type LogInfoRequest struct {
	// Empty struct for log info request
}

// LogInfoResponse represents log configuration info
type LogInfoResponse struct {
	ConsoleEnabled bool   `json:"console_enabled"`
	ConsoleLevel   string `json:"console_level"`
	FileEnabled    bool   `json:"file_enabled"`
	FilePath       string `json:"file_path,omitempty"`
	LokiEnabled    bool   `json:"loki_enabled"`
	LokiURL        string `json:"loki_url,omitempty"`
	SLSEnabled     bool   `json:"sls_enabled"`
}

func main() {
	app := mod.New()

	// Register log test service
	app.Register(mod.Service{
		Name:        "log_test",
		DisplayName: "日志测试",
		Description: "测试不同级别的日志输出",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *LogTestRequest, resp *LogTestResponse) error {
			logger := ctx.Logger()

			// 根据请求的级别记录日志
			switch req.Level {
			case "debug":
				logger.WithField("data", req.Data).Debug(req.Message)
			case "info":
				logger.WithField("data", req.Data).Info(req.Message)
			case "warn":
				logger.WithField("data", req.Data).Warn(req.Message)
			case "error":
				logger.WithField("data", req.Data).Error(req.Message)
			default:
				return mod.Reply(400, "无效的日志级别")
			}

			resp.Success = true
			resp.Message = "日志记录成功"
			resp.Timestamp = time.Now().Format(time.RFC3339)

			return nil
		}),
		Group:    "日志功能",
		Sort:     1,
		SkipAuth: true,
	})

	// Register log info service
	app.Register(mod.Service{
		Name:        "log_info",
		DisplayName: "日志配置信息",
		Description: "获取当前日志配置信息",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *LogInfoRequest, resp *LogInfoResponse) error {
			config := ctx.App().GetModConfig()
			if config == nil {
				resp.ConsoleEnabled = true
				resp.ConsoleLevel = "info"
				return nil
			}

			logging := config.Logging
			resp.ConsoleEnabled = logging.Console.Enabled
			resp.ConsoleLevel = logging.Console.Level
			resp.FileEnabled = logging.File.Enabled
			resp.FilePath = logging.File.Path
			resp.LokiEnabled = logging.Loki.Enabled
			resp.LokiURL = logging.Loki.URL
			resp.SLSEnabled = logging.SLS.Enabled

			return nil
		}),
		Group:    "日志功能",
		Sort:     2,
		SkipAuth: true,
	})

	app.Run(":8080")
}