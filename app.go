package mod

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

type Config struct {
	fiber.Config
	ServicePrefix string
	TokenKey      string
	Logger        *logrus.Logger
}

func New(config ...Config) *App {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Config.BodyLimit <= 0 {
		cfg.Config.BodyLimit = 100 * 1024 * 1024 // 100M
	}
	if cfg.ServicePrefix == "" {
		cfg.ServicePrefix = "/services"
	}
	if cfg.TokenKey == "" {
		cfg.TokenKey = "mod-key,mod-token"
	}
	if cfg.Logger == nil {
		cfg.Logger = logrus.StandardLogger()
		cfg.Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	}

	app := &App{
		App:       fiber.New(cfg.Config),
		cfg:       cfg,
		logger:    cfg.Logger,
		tokenKeys: SplitAndTrimSpace(cfg.TokenKey, ","),
	}
	return app
}

type App struct {
	*fiber.App
	logger    *logrus.Logger
	cfg       Config
	tokenKeys []string
}

func (app *App) Run(addr ...string) {
	var a string
	if len(addr) > 0 {
		a = addr[0]
	} else {
		a = ":8080"
	}
	app.logger.Info("Starting server on " + a)
	if err := app.Listen(a); err != nil {
		panic(err)
	}
}

func (app *App) Register(svc Service) error {
	if err := validate.Struct(&svc); err != nil {
		return err
	}

	// 构建服务路径
	servicePath := fmt.Sprintf("%s/%s", app.cfg.ServicePrefix, svc.Name)

	app.Add(fiber.MethodPost, servicePath, func(fc *fiber.Ctx) error {
		ctx := &Context{Ctx: fc, logger: app.logger}

		// 身份验证检查
		if !svc.SkipAuth {
			token := parseToken(fc, app.tokenKeys)
			if token == "" {
				return fc.Status(401).JSON(NewErrorResponse(ctx, 401, "Unauthorized"))
			}
		}

		// 创建输入参数实例
		var in, out interface{}
		if svc.Handler.InputType != nil {
			in = reflect.New(svc.Handler.InputType).Interface()
			// 解析请求参数到结构体
			if err := app.parseRequestParamsToStruct(fc, in); err != nil {
				app.logger.WithFields(logrus.Fields{
					"service": svc.Name,
					"error":   err.Error(),
					"body":    string(fc.Body()),
					"query":   fc.Context().QueryArgs().String(),
					"rid":     ctx.GetRequestID(),
				}).Error("Parameter parsing failed")
				return fc.Status(400).JSON(NewErrorResponse(ctx, 400, "Parameter parsing error", err.Error()))
			}

			// 参数验证
			if err := validate.Struct(in); err != nil {
				app.logger.WithFields(logrus.Fields{
					"service": svc.Name,
					"error":   err.Error(),
					"params":  fmt.Sprintf("%+v", in),
					"rid":     ctx.GetRequestID(),
				}).Error("Parameter validation failed")
				return fc.Status(400).JSON(NewErrorResponse(ctx, 400, "Parameter validation error", err.Error()))
			}
		}

		// 创建输出参数实例
		if svc.Handler.OutputType != nil {
			out = reflect.New(svc.Handler.OutputType).Interface()
		}

		// 调用服务处理函数
		if err := svc.Handler.Func(ctx, in, out); err != nil {
			app.logger.WithFields(logrus.Fields{
				"service": svc.Name,
				"error":   err.Error(),
				"params":  fmt.Sprintf("%+v", in),
				"rid":     ctx.GetRequestID(),
			}).Error("Service handler failed")

			if intlErr, ok := err.(*IntlError); ok {
				resp := NewErrorResponse(ctx, intlErr.Code(), intlErr.Msg(), intlErr.Detail())
				return fc.Status(intlErr.Code()).JSON(resp)
			}
			return fc.Status(500).JSON(NewErrorResponse(ctx, 500, err.Error()))
		}

		// 返回结果
		if svc.ReturnRaw {
			return fc.JSON(out)
		}
		return fc.JSON(NewSuccessResponse(ctx, out))
	})

	// 打印服务注册日志
	app.logger.WithFields(logrus.Fields{
		"service":     svc.Name,
		"displayName": svc.DisplayName,
		"method":      "POST",
		"path":        servicePath,
		"skipAuth":    svc.SkipAuth,
		"returnRaw":   svc.ReturnRaw,
	}).Info("Service registered")

	return nil
}

func parseToken(kc *fiber.Ctx, keys []string) string {
	cacheKey := "MOD_TOKEN"
	if v := kc.Context().UserValue(cacheKey); v != nil {
		if t, ok := v.(string); ok {
			return t
		}
	}
	var value string
	for _, key := range keys {
		if v := kc.Get(key); v != "" {
			value = v
			break
		}
	}
	if value == "" {
		for _, key := range keys {
			if v := kc.Query(key); v != "" {
				value = v
				break
			}
		}
	}
	if value != "" {
		kc.Context().SetUserValue(cacheKey, value)
	}
	return value
}

func (app *App) parseRequestParamsToStruct(fc *fiber.Ctx, in interface{}) error {
	if in == nil {
		return nil
	}

	rv := reflect.ValueOf(in)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("input parameter must be a pointer")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("input parameter must be a pointer to struct")
	}

	rt := rv.Type()

	// 首先解析 JSON body（如果存在）
	body := fc.Body()
	if len(body) > 0 {
		if err := json.Unmarshal(body, in); err != nil {
			return fmt.Errorf("failed to parse JSON body: %w", err)
		}
	}

	// 然后根据 mod 标签或默认规则解析其他来源的参数
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		if !field.CanSet() {
			continue
		}

		fieldName := fieldType.Name
		var value string

		// 检查 mod 标签
		modTag := fieldType.Tag.Get("mod")
		if modTag != "" {
			value = app.parseFieldValue(fc, modTag, fieldName)
		} else {
			// 如果没有 mod 标签，默认从多个来源尝试获取
			// 优先级：query -> form -> header
			// 尝试小写字段名
			lowerFieldName := strings.ToLower(fieldName)
			if v := fc.Query(lowerFieldName); v != "" {
				value = v
			} else if v := fc.FormValue(lowerFieldName); v != "" {
				value = v
			} else if v := fc.Get(lowerFieldName); v != "" {
				value = v
			} else {
				// 也尝试原始字段名
				if v := fc.Query(fieldName); v != "" {
					value = v
				} else if v := fc.FormValue(fieldName); v != "" {
					value = v
				} else if v := fc.Get(fieldName); v != "" {
					value = v
				}
			}
		}

		if value != "" {
			app.setFieldValue(field, value)
		}
	}

	return nil
}

func (app *App) parseFieldValue(fc *fiber.Ctx, modTag, fieldName string) string {
	// 解析 mod 标签，格式如 "from=query" 或 "from=header;name=custom-header"
	parts := strings.Split(modTag, ";")
	from := ""
	name := strings.ToLower(fieldName) // 默认使用小写字段名

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			switch key {
			case "from":
				from = value
			case "name":
				name = value
			}
		}
	}

	switch from {
	case "query":
		return fc.Query(name)
	case "header":
		return fc.Get(name)
	case "form":
		return fc.FormValue(name)
	case "param":
		return fc.Params(name)
	default:
		// 默认尝试从 query 获取
		return fc.Query(name)
	}
}

func (app *App) setFieldValue(field reflect.Value, value string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := parseInt(value); err == nil {
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := parseUint(value); err == nil {
			field.SetUint(uintVal)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := parseFloat(value); err == nil {
			field.SetFloat(floatVal)
		}
	case reflect.Bool:
		if boolVal, err := parseBool(value); err == nil {
			field.SetBool(boolVal)
		}
	}
}
