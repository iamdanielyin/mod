package mod

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

type Config struct {
	fiber.Config
	ServicePrefix string
	TokenKey      string
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

	app := &App{
		App:       fiber.New(cfg.Config),
		cfg:       cfg,
		tokenKeys: SplitAndTrimSpace(cfg.TokenKey, ","),
	}
	return app
}

type App struct {
	*fiber.App
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
	if err := app.Listen(a); err != nil {
		panic(err)
	}
}

func (app *App) Register(svc Service) error {
	if err := validate.Struct(&svc); err != nil {
		return err
	}
	app.Add(fiber.MethodPost, fmt.Sprintf("%s/%s", app.cfg.ServicePrefix, svc.Name), func(fc *fiber.Ctx) error {
		svc.Handler()
	})
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
