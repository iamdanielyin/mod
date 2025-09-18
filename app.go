package mod

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"html/template"
	"reflect"
	"sort"
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

	// 注册文档路由
	app.Get("/services/docs", app.handleDocs)

	return app
}

type App struct {
	*fiber.App
	logger    *logrus.Logger
	cfg       Config
	tokenKeys []string
	services  []Service // 存储已注册的服务用于生成文档
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

	// 保存服务信息用于生成文档
	app.services = append(app.services, svc)

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

// 文档生成相关结构体
type DocField struct {
	Name        string
	Type        string
	Description string
	Required    bool
	From        string // query, header, form, param
	Tag         string
}

type DocService struct {
	Service
	ServicePath  string
	InputFields  []DocField
	OutputFields []DocField
}

type DocGroup struct {
	Name     string
	Services []DocService
}

// 处理文档请求
func (app *App) handleDocs(c *fiber.Ctx) error {
	// 按组分类并排序服务
	groups := app.groupAndSortServices()

	// 生成HTML
	html := app.generateDocsHTML(groups)

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

// 按组分类并排序服务
func (app *App) groupAndSortServices() []DocGroup {
	groupMap := make(map[string][]DocService)

	// 处理每个服务
	for _, svc := range app.services {
		docSvc := DocService{
			Service:     svc,
			ServicePath: fmt.Sprintf("%s/%s", app.cfg.ServicePrefix, svc.Name),
		}

		// 解析输入参数
		if svc.Handler.InputType != nil {
			docSvc.InputFields = app.parseStructFields(svc.Handler.InputType)
		}

		// 解析输出参数
		if svc.Handler.OutputType != nil {
			docSvc.OutputFields = app.parseStructFields(svc.Handler.OutputType)
		}

		// 按组分类
		groupName := svc.Group
		if groupName == "" {
			groupName = "默认分组"
		}
		groupMap[groupName] = append(groupMap[groupName], docSvc)
	}

	// 转换为有序数组
	var groups []DocGroup
	var groupNames []string
	for groupName := range groupMap {
		groupNames = append(groupNames, groupName)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		services := groupMap[groupName]
		// 按Sort字段排序服务
		sort.Slice(services, func(i, j int) bool {
			if services[i].Sort == services[j].Sort {
				return services[i].Name < services[j].Name
			}
			return services[i].Sort < services[j].Sort
		})

		groups = append(groups, DocGroup{
			Name:     groupName,
			Services: services,
		})
	}

	return groups
}

// 解析结构体字段
func (app *App) parseStructFields(t reflect.Type) []DocField {
	var fields []DocField

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fields
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		docField := DocField{
			Name: field.Name,
			Type: app.getFieldTypeString(field.Type),
		}

		// 解析标签
		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			if strings.Contains(validateTag, "required") {
				docField.Required = true
			}
		}

		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				docField.Name = parts[0]
			}
		}

		if modTag := field.Tag.Get("mod"); modTag != "" {
			docField.From = app.parseModTagFrom(modTag)
			docField.Tag = modTag
		} else {
			docField.From = "body"
		}

		if descTag := field.Tag.Get("desc"); descTag != "" {
			docField.Description = descTag
		}

		fields = append(fields, docField)
	}

	return fields
}

// 获取字段类型字符串
func (app *App) getFieldTypeString(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return "*" + app.getFieldTypeString(t.Elem())
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "bool"
	case reflect.Slice:
		return "[]" + app.getFieldTypeString(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", app.getFieldTypeString(t.Key()), app.getFieldTypeString(t.Elem()))
	case reflect.Struct:
		return t.Name()
	default:
		return t.String()
	}
}

// 解析mod标签的from参数
func (app *App) parseModTagFrom(modTag string) string {
	parts := strings.Split(modTag, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == "from" {
			return strings.TrimSpace(kv[1])
		}
	}
	return "query"
}

// 生成HTML文档
func (app *App) generateDocsHTML(groups []DocGroup) string {
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API 文档</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji';
            line-height: 1.5715;
            color: rgba(0, 0, 0, 0.85);
            background-color: #f0f2f5;
        }

        .container {
            display: flex;
            height: 100vh;
        }

        .sidebar {
            width: 300px;
            background: #fff;
            border-right: 1px solid #f0f0f0;
            overflow-y: auto;
            position: fixed;
            height: 100vh;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
        }

        .sidebar-header {
            padding: 16px 24px;
            border-bottom: 1px solid #f0f0f0;
            background: #001529;
            color: #fff;
        }

        .sidebar-header h1 {
            font-size: 16px;
            font-weight: 600;
            margin: 0;
        }

        .group {
            margin: 0;
        }

        .group-title {
            padding: 8px 24px;
            background: #fafafa;
            font-weight: 500;
            font-size: 12px;
            color: rgba(0, 0, 0, 0.45);
            border-bottom: 1px solid #f0f0f0;
            cursor: pointer;
            transition: background-color 0.3s;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .group-title:hover {
            background: #f5f5f5;
        }

        .service-list {
            background: white;
        }

        .service-item {
            padding: 12px 24px 12px 48px;
            cursor: pointer;
            border-bottom: 1px solid #f0f0f0;
            transition: all 0.3s;
            font-size: 14px;
            color: rgba(0, 0, 0, 0.85);
        }

        .service-item:hover {
            background: #f5f5f5;
            color: #1890ff;
        }

        .service-item.active {
            background: #e6f7ff;
            border-right: 2px solid #1890ff;
            color: #1890ff;
            font-weight: 500;
        }

        .main-content {
            flex: 1;
            margin-left: 300px;
            padding: 24px;
            overflow-y: auto;
        }

        .api-section {
            background: white;
            border-radius: 6px;
            margin-bottom: 16px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
            border: 1px solid #f0f0f0;
            overflow: hidden;
        }

        .api-header {
            padding: 16px 24px;
            background: #1890ff;
            color: white;
            border-bottom: 1px solid #40a9ff;
        }

        .api-title {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 8px;
        }

        .api-path {
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
            font-size: 12px;
            background: rgba(255, 255, 255, 0.2);
            padding: 4px 8px;
            border-radius: 4px;
            display: inline-block;
            margin-bottom: 12px;
            border: 1px solid rgba(255, 255, 255, 0.3);
        }

        .api-meta {
            display: flex;
            gap: 24px;
            flex-wrap: wrap;
            font-size: 12px;
        }

        .meta-item {
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .meta-label {
            color: rgba(255, 255, 255, 0.85);
            font-weight: 400;
        }

        .meta-value {
            font-weight: 500;
            padding: 2px 6px;
            background: rgba(255, 255, 255, 0.15);
            border-radius: 4px;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .auth-required-header {
            background: #ff4d4f;
            color: #fff;
        }

        .auth-not-required-header {
            background: #52c41a;
            color: #fff;
        }

        .api-body {
            padding: 24px;
        }

        .params-section {
            margin-bottom: 32px;
        }

        .section-title {
            font-size: 16px;
            font-weight: 600;
            margin-bottom: 16px;
            color: rgba(0, 0, 0, 0.85);
            border-bottom: none;
            padding-bottom: 0;
        }

        .params-table {
            width: 100%;
            border-collapse: collapse;
            background: white;
            border-radius: 6px;
            overflow: hidden;
            border: 1px solid #f0f0f0;
        }

        .params-table th,
        .params-table td {
            padding: 12px 16px;
            text-align: left;
            border-bottom: 1px solid #f0f0f0;
        }

        .params-table th {
            background: #fafafa;
            font-weight: 500;
            color: rgba(0, 0, 0, 0.85);
            font-size: 14px;
        }

        .params-table td {
            font-size: 14px;
            color: rgba(0, 0, 0, 0.85);
        }

        .params-table tr:last-child td {
            border-bottom: none;
        }

        .params-table tr:hover {
            background: #fafafa;
        }

        .field-name {
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
            font-weight: 600;
            color: #1890ff;
        }

        .field-type {
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
            color: #722ed1;
            background: #f9f0ff;
            padding: 2px 6px;
            border-radius: 4px;
            border: 1px solid #d3adf7;
        }

        .required {
            color: #ff4d4f;
            font-weight: 500;
        }

        .not-required {
            color: rgba(0, 0, 0, 0.45);
        }

        .from-tag {
            font-size: 12px;
            background: #1890ff;
            color: white;
            padding: 2px 6px;
            border-radius: 4px;
            font-weight: 400;
            display: inline-block;
        }

        .empty-state {
            text-align: center;
            color: rgba(0, 0, 0, 0.45);
            font-style: italic;
            padding: 48px 24px;
            background: #fafafa;
            border-radius: 6px;
            border: 1px dashed #d9d9d9;
        }

        @media (max-width: 768px) {
            .container {
                flex-direction: column;
            }

            .sidebar {
                position: relative;
                width: 100%;
                height: auto;
            }

            .main-content {
                margin-left: 0;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="sidebar">
            <div class="sidebar-header">
                <h1>API 文档</h1>
            </div>
            {{range .}}
            <div class="group">
                <div class="group-title">{{.Name}}</div>
                <div class="service-list">
                    {{range .Services}}
                    <div class="service-item" onclick="scrollToService('service-{{.Name}}')">
                        {{.DisplayName}}
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>

        <div class="main-content">
            {{range .}}
            {{range .Services}}
            <div class="api-section" id="service-{{.Name}}">
                <div class="api-header">
                    <div class="api-title">{{.DisplayName}}</div>
                    <div class="api-path">POST {{.ServicePath}}</div>
                    <div class="api-meta">
                        <div class="meta-item">
                            <span class="meta-label">服务名称:</span>
                            <span class="meta-value">{{.Name}}</span>
                        </div>
                        <div class="meta-item">
                            <span class="meta-label">认证:</span>
                            <span class="meta-value {{if .SkipAuth}}auth-not-required-header{{else}}auth-required-header{{end}}">
                                {{if .SkipAuth}}不需要{{else}}需要{{end}}
                            </span>
                        </div>
                        {{if .Description}}
                        <div class="meta-item">
                            <span class="meta-label">描述:</span>
                            <span class="meta-value">{{.Description}}</span>
                        </div>
                        {{end}}
                    </div>
                </div>
                <div class="api-body">

                    {{if .InputFields}}
                    <div class="params-section">
                        <div class="section-title">请求参数</div>
                        <table class="params-table">
                            <thead>
                                <tr>
                                    <th>参数名</th>
                                    <th>类型</th>
                                    <th>来源</th>
                                    <th>必填</th>
                                    <th>描述</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range .InputFields}}
                                <tr>
                                    <td><span class="field-name">{{.Name}}</span></td>
                                    <td><span class="field-type">{{.Type}}</span></td>
                                    <td><span class="from-tag">{{.From}}</span></td>
                                    <td><span class="{{if .Required}}required{{else}}not-required{{end}}">{{if .Required}}是{{else}}否{{end}}</span></td>
                                    <td>{{.Description}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                    {{else}}
                    <div class="params-section">
                        <div class="section-title">请求参数</div>
                        <div class="empty-state">无参数</div>
                    </div>
                    {{end}}

                    {{if .OutputFields}}
                    <div class="params-section">
                        <div class="section-title">返回参数</div>
                        <table class="params-table">
                            <thead>
                                <tr>
                                    <th>参数名</th>
                                    <th>类型</th>
                                    <th>描述</th>
                                </tr>
                            </thead>
                            <tbody>
                                {{range .OutputFields}}
                                <tr>
                                    <td><span class="field-name">{{.Name}}</span></td>
                                    <td><span class="field-type">{{.Type}}</span></td>
                                    <td>{{.Description}}</td>
                                </tr>
                                {{end}}
                            </tbody>
                        </table>
                    </div>
                    {{else}}
                    <div class="params-section">
                        <div class="section-title">返回参数</div>
                        <div class="empty-state">无返回参数</div>
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
            {{end}}
        </div>
    </div>

    <script>
        function scrollToService(serviceId) {
            const element = document.getElementById(serviceId);
            if (element) {
                element.scrollIntoView({ behavior: 'smooth', block: 'start' });

                // 更新激活状态
                document.querySelectorAll('.service-item').forEach(item => {
                    item.classList.remove('active');
                });
                event.target.classList.add('active');
            }
        }

        // 滚动监听，自动更新侧边栏激活状态
        function updateActiveService() {
            const sections = document.querySelectorAll('.api-section');
            const serviceItems = document.querySelectorAll('.service-item');

            let current = '';
            sections.forEach(section => {
                const rect = section.getBoundingClientRect();
                if (rect.top <= 100) {
                    current = section.id;
                }
            });

            serviceItems.forEach(item => {
                item.classList.remove('active');
                if (item.getAttribute('onclick').includes(current)) {
                    item.classList.add('active');
                }
            });
        }

        window.addEventListener('scroll', updateActiveService);
        document.addEventListener('DOMContentLoaded', updateActiveService);
    </script>
</body>
</html>`

	t := template.Must(template.New("docs").Parse(tmpl))
	var buf strings.Builder
	t.Execute(&buf, groups)
	return buf.String()
}
