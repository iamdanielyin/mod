package mod

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/dgraph-io/badger/v4"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"html/template"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// ModConfig represents the structure of mod.yml configuration file
type ModConfig struct {
	App struct {
		Name        string `yaml:"name"`
		DisplayName string `yaml:"display_name"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	} `yaml:"app"`

	Cache struct {
		BigCache struct {
			Enabled            bool   `yaml:"enabled"`
			Shards             int    `yaml:"shards"`
			LifeWindow         string `yaml:"life_window"`
			CleanWindow        string `yaml:"clean_window"`
			MaxEntriesInWindow int    `yaml:"max_entries_in_window"`
			MaxEntrySize       int    `yaml:"max_entry_size"`
			Verbose            bool   `yaml:"verbose"`
			HardMaxCacheSize   int    `yaml:"hard_max_cache_size"`
		} `yaml:"bigcache"`

		Badger struct {
			Enabled                 bool   `yaml:"enabled"`
			Path                    string `yaml:"path"`
			InMemory                bool   `yaml:"in_memory"`
			SyncWrites              bool   `yaml:"sync_writes"`
			ValueLogFileSize        int    `yaml:"value_log_file_size"`
			NumCompactors           int    `yaml:"num_compactors"`
			NumLevelZeroTables      int    `yaml:"num_level_zero_tables"`
			NumLevelZeroTablesStall int    `yaml:"num_level_zero_tables_stall"`
			ValueLogLoadSize        int    `yaml:"value_log_load_size"`
			TTL                     string `yaml:"ttl"` // Token 过期时间
		} `yaml:"badger"`

		Redis struct {
			Enabled      bool   `yaml:"enabled"`
			Address      string `yaml:"address"`
			Password     string `yaml:"password"`
			DB           int    `yaml:"db"`
			PoolSize     int    `yaml:"pool_size"`
			MinIdleConns int    `yaml:"min_idle_conns"`
			DialTimeout  string `yaml:"dial_timeout"`
			ReadTimeout  string `yaml:"read_timeout"`
			WriteTimeout string `yaml:"write_timeout"`
			IdleTimeout  string `yaml:"idle_timeout"`
			MaxConnAge   string `yaml:"max_conn_age"`
			TTL          string `yaml:"ttl"` // Token 过期时间
		} `yaml:"redis"`
	} `yaml:"cache"`

	RSAKeys struct {
		PrivateKey string `yaml:"private_key"`
		PublicKey  string `yaml:"public_key"`
	} `yaml:"rsa_keys"`

	FileUpload struct {
		Local struct {
			Enabled      bool     `yaml:"enabled"`
			UploadDir    string   `yaml:"upload_dir"`
			MaxSize      string   `yaml:"max_size"`
			AllowedTypes []string `yaml:"allowed_types"`
		} `yaml:"local"`

		S3 struct {
			Enabled   bool   `yaml:"enabled"`
			Bucket    string `yaml:"bucket"`
			Region    string `yaml:"region"`
			AccessKey string `yaml:"access_key"`
			SecretKey string `yaml:"secret_key"`
			Endpoint  string `yaml:"endpoint"`
		} `yaml:"s3"`

		OSS struct {
			Enabled         bool   `yaml:"enabled"`
			Bucket          string `yaml:"bucket"`
			Endpoint        string `yaml:"endpoint"`
			AccessKeyID     string `yaml:"access_key_id"`
			AccessKeySecret string `yaml:"access_key_secret"`
		} `yaml:"oss"`
	} `yaml:"file_upload"`

	StaticMounts []struct {
		URLPrefix  string `yaml:"url_prefix"`
		LocalPath  string `yaml:"local_path"`
		Browseable bool   `yaml:"browseable"`
		IndexFile  string `yaml:"index_file"`
	} `yaml:"static_mounts"`

	Logging struct {
		Console struct {
			Enabled bool   `yaml:"enabled"`
			Level   string `yaml:"level"`
		} `yaml:"console"`

		Loki struct {
			Enabled   bool              `yaml:"enabled"`
			URL       string            `yaml:"url"`
			Labels    map[string]string `yaml:"labels"`
			BatchSize int               `yaml:"batch_size"`
			Timeout   string            `yaml:"timeout"`
		} `yaml:"loki"`

		SLS struct {
			Enabled         bool   `yaml:"enabled"`
			Endpoint        string `yaml:"endpoint"`
			Project         string `yaml:"project"`
			Logstore        string `yaml:"logstore"`
			AccessKeyID     string `yaml:"access_key_id"`
			AccessKeySecret string `yaml:"access_key_secret"`
		} `yaml:"sls"`

		File struct {
			Enabled    bool   `yaml:"enabled"`
			Path       string `yaml:"path"`
			MaxSize    string `yaml:"max_size"`
			MaxBackups int    `yaml:"max_backups"`
			MaxAge     string `yaml:"max_age"`
			Compress   bool   `yaml:"compress"`
		} `yaml:"file"`
	} `yaml:"logging"`

	Token struct {
		JWT struct {
			Enabled               bool   `yaml:"enabled"`
			SecretKey             string `yaml:"secret_key"`
			Issuer                string `yaml:"issuer"`
			ExpireDuration        string `yaml:"expire_duration"`
			RefreshExpireDuration string `yaml:"refresh_expire_duration"`
			Algorithm             string `yaml:"algorithm"`
		} `yaml:"jwt"`

		Validation struct {
			Enabled          bool   `yaml:"enabled"`
			SkipExpiredCheck bool   `yaml:"skip_expired_check"`
			CacheStrategy    string `yaml:"cache_strategy"` // "bigcache", "badger", "redis"
			CacheKeyPrefix   string `yaml:"cache_key_prefix"`
		} `yaml:"validation"`
	} `yaml:"token"`

	Settings struct {
		Port           int    `yaml:"port"`
		ReadTimeout    string `yaml:"read_timeout"`
		WriteTimeout   string `yaml:"write_timeout"`
		MaxConnections int    `yaml:"max_connections"`
		CacheStrategy  string `yaml:"cache_strategy"`
	} `yaml:"settings"`
}

// loadModConfig attempts to load configuration from mod.yml file
func loadModConfig() (*ModConfig, error) {
	var configPath string

	// First, check MOD_PATH environment variable
	if envPath := os.Getenv("MOD_PATH"); envPath != "" {
		configPath = envPath
	} else {
		// Second, check for mod.yml in current directory
		if _, err := os.Stat("mod.yml"); err == nil {
			configPath = "mod.yml"
		} else {
			// No configuration file found
			return nil, nil
		}
	}

	// Read the configuration file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config ModConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return &config, nil
}

// mergeConfigs merges ModConfig into Config, with manual config taking precedence
func mergeConfigs(fileConfig *ModConfig, manualConfig Config) Config {
	// Start with manual config values
	merged := manualConfig

	// Store the complete ModConfig for later use
	merged.ModConfig = fileConfig

	// Only override if manual config has default/empty values
	if merged.Name == "" && fileConfig.App.Name != "" {
		merged.Name = fileConfig.App.Name
	}
	if merged.DisplayName == "" && fileConfig.App.DisplayName != "" {
		merged.DisplayName = fileConfig.App.DisplayName
	}
	if merged.Description == "" && fileConfig.App.Description != "" {
		merged.Description = fileConfig.App.Description
	}

	// Server settings from settings section
	if merged.BodyLimit <= 0 && fileConfig.Settings.MaxConnections > 0 {
		// Use max_connections as a proxy for body limit if not explicitly set
		merged.BodyLimit = 100 * 1024 * 1024 // Default 100MB
	}

	return merged
}

// applyLoggingConfig applies logging configuration from mod.yml to logger
func applyLoggingConfig(logger *logrus.Logger, config *ModConfig) {
	if config == nil {
		return
	}

	// Set log level from console logging config
	if config.Logging.Console.Enabled && config.Logging.Console.Level != "" {
		if level, err := logrus.ParseLevel(config.Logging.Console.Level); err == nil {
			logger.SetLevel(level)
		}
	}

	// Always use text formatter for now, but can be extended
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
}

type Config struct {
	fiber.Config
	Name        string
	DisplayName string
	Description string

	ServicePrefix string
	TokenKey      string
	Logger        *logrus.Logger

	// ModConfig holds the complete configuration from mod.yml
	ModConfig *ModConfig `json:"-"`
}

func New(config ...Config) *App {
	var cfg Config
	var fileConfig *ModConfig
	var err error

	if len(config) > 0 {
		cfg = config[0]
	}

	// Try to load configuration from mod.yml file
	if fileConfig, err = loadModConfig(); err != nil {
		// Log warning but continue with manual config
		logrus.Warnf("Failed to load mod.yml config: %v", err)
	} else if fileConfig != nil {
		// Merge file config with manual config, manual takes precedence
		cfg = mergeConfigs(fileConfig, cfg)
		logrus.Infof("Loaded configuration from mod.yml")
	}

	// Apply default values if still empty
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

	// Apply logging configuration from file if available
	if fileConfig != nil {
		applyLoggingConfig(cfg.Logger, fileConfig)
	}

	app := &App{
		App:       fiber.New(cfg.Config),
		cfg:       cfg,
		logger:    cfg.Logger,
		tokenKeys: SplitAndTrimSpace(cfg.TokenKey, ","),
	}

	// 初始化 Token 缓存
	if fileConfig != nil && fileConfig.Token.Validation.Enabled {
		switch fileConfig.Token.Validation.CacheStrategy {
		case "bigcache":
			if fileConfig.Cache.BigCache.Enabled {
				app.initTokenCache(fileConfig)
			}
		case "badger":
			if fileConfig.Cache.Badger.Enabled {
				app.initBadgerDB(fileConfig)
			}
		case "redis":
			if fileConfig.Cache.Redis.Enabled {
				app.initRedisClient(fileConfig)
			}
		}
	}

	// 注册文档路由
	app.Get("/services/docs", app.handleDocs)

	return app
}

// initTokenCache 初始化 Token 缓存
func (app *App) initTokenCache(config *ModConfig) {
	if !config.Cache.BigCache.Enabled {
		return
	}

	// 解析配置参数
	lifeWindow, err := time.ParseDuration(config.Cache.BigCache.LifeWindow)
	if err != nil {
		app.logger.WithError(err).Warn("Invalid BigCache life_window, using default 24h")
		lifeWindow = 24 * time.Hour
	}

	cleanWindow, err := time.ParseDuration(config.Cache.BigCache.CleanWindow)
	if err != nil {
		app.logger.WithError(err).Warn("Invalid BigCache clean_window, using default 1h")
		cleanWindow = time.Hour
	}

	maxEntries := config.Cache.BigCache.MaxEntriesInWindow
	if maxEntries <= 0 {
		maxEntries = 10000 // 默认值
	}

	// 创建 BigCache 配置
	bigCacheConfig := bigcache.Config{
		Shards:             config.Cache.BigCache.Shards,
		LifeWindow:         lifeWindow,
		CleanWindow:        cleanWindow,
		MaxEntriesInWindow: maxEntries,
		MaxEntrySize:       config.Cache.BigCache.MaxEntrySize,
		Verbose:            config.Cache.BigCache.Verbose,
		HardMaxCacheSize:   config.Cache.BigCache.HardMaxCacheSize,
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}

	// 初始化 BigCache
	cache, err := bigcache.New(context.Background(), bigCacheConfig)
	if err != nil {
		app.logger.WithError(err).Error("Failed to initialize BigCache for token validation")
		return
	}

	app.tokenCache = cache
	app.logger.Info("BigCache for token validation initialized successfully")
}

// initBadgerDB 初始化 BadgerDB
func (app *App) initBadgerDB(config *ModConfig) {
	if !config.Cache.Badger.Enabled {
		return
	}

	dbPath := config.Cache.Badger.Path
	if dbPath == "" {
		dbPath = "./data/tokens" // 默认路径
	}

	// 创建 BadgerDB 选项
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = &badgerLogger{logger: app.logger} // 使用自定义 logger
	opts.InMemory = config.Cache.Badger.InMemory
	opts.SyncWrites = config.Cache.Badger.SyncWrites

	if config.Cache.Badger.ValueLogFileSize > 0 {
		opts.ValueLogFileSize = int64(config.Cache.Badger.ValueLogFileSize)
	}
	if config.Cache.Badger.NumCompactors > 0 {
		opts.NumCompactors = config.Cache.Badger.NumCompactors
	}
	if config.Cache.Badger.NumLevelZeroTables > 0 {
		opts.NumLevelZeroTables = config.Cache.Badger.NumLevelZeroTables
	}
	if config.Cache.Badger.NumLevelZeroTablesStall > 0 {
		opts.NumLevelZeroTablesStall = config.Cache.Badger.NumLevelZeroTablesStall
	}

	// 打开 BadgerDB
	db, err := badger.Open(opts)
	if err != nil {
		app.logger.WithError(err).WithField("path", dbPath).Error("Failed to initialize BadgerDB for token validation")
		return
	}

	app.badgerDB = db
	app.logger.WithField("path", dbPath).Info("BadgerDB for token validation initialized successfully")
}

// badgerLogger 实现 BadgerDB 的 Logger 接口
type badgerLogger struct {
	logger *logrus.Logger
}

func (bl *badgerLogger) Errorf(f string, v ...interface{}) {
	bl.logger.Errorf("BadgerDB: "+f, v...)
}

func (bl *badgerLogger) Warningf(f string, v ...interface{}) {
	bl.logger.Warnf("BadgerDB: "+f, v...)
}

func (bl *badgerLogger) Infof(f string, v ...interface{}) {
	bl.logger.Infof("BadgerDB: "+f, v...)
}

func (bl *badgerLogger) Debugf(f string, v ...interface{}) {
	bl.logger.Debugf("BadgerDB: "+f, v...)
}

// initRedisClient 初始化 Redis 客户端
func (app *App) initRedisClient(config *ModConfig) {
	if !config.Cache.Redis.Enabled {
		return
	}

	// 从主 Redis 配置获取连接信息
	redisConfig := config.Cache.Redis
	if redisConfig.Address == "" {
		app.logger.Error("Redis address not configured for token validation")
		return
	}

	// 创建 Redis 客户端选项
	opts := &redis.Options{
		Addr:         redisConfig.Address,
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
	}

	// 解析超时时间
	if redisConfig.DialTimeout != "" {
		if dialTimeout, err := time.ParseDuration(redisConfig.DialTimeout); err == nil {
			opts.DialTimeout = dialTimeout
		}
	}
	if redisConfig.ReadTimeout != "" {
		if readTimeout, err := time.ParseDuration(redisConfig.ReadTimeout); err == nil {
			opts.ReadTimeout = readTimeout
		}
	}
	if redisConfig.WriteTimeout != "" {
		if writeTimeout, err := time.ParseDuration(redisConfig.WriteTimeout); err == nil {
			opts.WriteTimeout = writeTimeout
		}
	}
	if redisConfig.IdleTimeout != "" {
		if idleTimeout, err := time.ParseDuration(redisConfig.IdleTimeout); err == nil {
			opts.ConnMaxIdleTime = idleTimeout
		}
	}
	if redisConfig.MaxConnAge != "" {
		if maxConnAge, err := time.ParseDuration(redisConfig.MaxConnAge); err == nil {
			opts.ConnMaxLifetime = maxConnAge
		}
	}

	// 创建 Redis 客户端
	rdb := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		app.logger.WithError(err).WithField("address", redisConfig.Address).Error("Failed to connect to Redis for token validation")
		return
	}

	app.redisClient = rdb
	app.logger.WithField("address", redisConfig.Address).Info("Redis client for token validation initialized successfully")
}

type App struct {
	*fiber.App
	logger      *logrus.Logger
	cfg         Config
	tokenKeys   []string
	services    []Service          // 存储已注册的服务用于生成文档
	tokenCache  *bigcache.BigCache // Token验证缓存
	badgerDB    *badger.DB         // BadgerDB 实例
	redisClient *redis.Client      // Redis 客户端
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

// GetModConfig returns the loaded mod.yml configuration
// Returns nil if no mod.yml was loaded
func (app *App) GetModConfig() *ModConfig {
	return app.cfg.ModConfig
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

			// 验证 token 的有效性
			if !app.validateToken(token) {
				app.logger.WithFields(logrus.Fields{
					"service": svc.Name,
					"token":   token,
					"rid":     ctx.GetRequestID(),
				}).Warn("Token validation failed")
				return fc.Status(401).JSON(NewErrorResponse(ctx, 401, "Invalid token"))
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

			if intlErr, ok := err.(*StdReply); ok {
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

// validateToken 验证 token 的有效性
// 当 SkipAuth 为 false 时，需要验证 token 是否在缓存中存在
func (app *App) validateToken(token string) bool {
	// 如果没有配置 token 验证，或者验证被禁用，则跳过验证
	if app.cfg.ModConfig == nil || !app.cfg.ModConfig.Token.Validation.Enabled {
		return true
	}

	if token == "" {
		return false
	}

	config := app.cfg.ModConfig.Token.Validation
	cacheKey := config.CacheKeyPrefix + token

	// 根据配置的缓存策略进行验证
	switch config.CacheStrategy {
	case "bigcache":
		if app.tokenCache != nil {
			// 查询 BigCache 中是否存在该 token
			_, err := app.tokenCache.Get(cacheKey)
			if err != nil {
				// 如果是 bigcache.ErrEntryNotFound，说明 token 不存在或已过期
				if err == bigcache.ErrEntryNotFound {
					app.logger.WithFields(logrus.Fields{
						"token":     token,
						"cache_key": cacheKey,
					}).Debug("Token not found in BigCache")
					return false
				}
				// 其他错误，记录日志但允许通过（避免缓存问题影响正常业务）
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Warn("BigCache query error, allowing token validation to pass")
				return true
			}
			// Token 存在，验证通过
			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
			}).Debug("Token validated successfully in BigCache")
			return true
		}
	case "badger":
		if app.badgerDB != nil {
			// 查询 BadgerDB 中是否存在该 token
			err := app.badgerDB.View(func(txn *badger.Txn) error {
				_, err := txn.Get([]byte(cacheKey))
				return err
			})

			if err != nil {
				if err == badger.ErrKeyNotFound {
					app.logger.WithFields(logrus.Fields{
						"token":     token,
						"cache_key": cacheKey,
					}).Debug("Token not found in BadgerDB")
					return false
				}
				// 其他错误，记录日志但允许通过
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Warn("BadgerDB query error, allowing token validation to pass")
				return true
			}

			// Token 存在，验证通过
			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
			}).Debug("Token validated successfully in BadgerDB")
			return true
		}
	case "redis":
		if app.redisClient != nil {
			// 查询 Redis 中是否存在该 token
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			exists, err := app.redisClient.Exists(ctx, cacheKey).Result()
			if err != nil {
				// Redis 查询错误，记录日志但允许通过
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Warn("Redis query error, allowing token validation to pass")
				return true
			}

			if exists == 0 {
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
				}).Debug("Token not found in Redis")
				return false
			}

			// Token 存在，验证通过
			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
			}).Debug("Token validated successfully in Redis")
			return true
		}
	}

	// 如果没有匹配的缓存策略，默认返回 false
	app.logger.WithFields(logrus.Fields{
		"token":          token,
		"cache_strategy": config.CacheStrategy,
		"cache_key":      cacheKey,
	}).Warn("Token validation failed: no valid cache strategy configured")

	return false
}

// SetToken 将 token 添加到缓存中
// 这个方法可以在用户登录时调用，将有效的 token 存储到缓存中
func (app *App) SetToken(token string, data interface{}) error {
	if app.cfg.ModConfig == nil || !app.cfg.ModConfig.Token.Validation.Enabled {
		return nil
	}

	config := app.cfg.ModConfig.Token.Validation
	cacheKey := config.CacheKeyPrefix + token

	switch config.CacheStrategy {
	case "bigcache":
		if app.tokenCache != nil {
			// 将数据序列化为 JSON
			var value []byte
			var err error
			if data != nil {
				value, err = json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to marshal token data: %w", err)
				}
			} else {
				value = []byte("1") // 如果没有数据，存储一个简单标记
			}

			err = app.tokenCache.Set(cacheKey, value)
			if err != nil {
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Error("Failed to set token in BigCache")
				return fmt.Errorf("failed to set token in BigCache: %w", err)
			}

			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
			}).Debug("Token set successfully in BigCache")
			return nil
		}
	case "badger":
		if app.badgerDB != nil {
			// 将数据序列化为 JSON
			var value []byte
			var err error
			if data != nil {
				value, err = json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to marshal token data: %w", err)
				}
			} else {
				value = []byte("1") // 如果没有数据，存储一个简单标记
			}

			// 解析 TTL
			var ttl time.Duration
			if config.Cache.Badger.TTL != "" {
				ttl, err = time.ParseDuration(config.Cache.Badger.TTL)
				if err != nil {
					app.logger.WithError(err).Warn("Invalid BadgerDB TTL, using default 24h")
					ttl = 24 * time.Hour
				}
			} else {
				ttl = 24 * time.Hour // 默认 24 小时
			}

			// 存储到 BadgerDB
			err = app.badgerDB.Update(func(txn *badger.Txn) error {
				entry := badger.NewEntry([]byte(cacheKey), value).WithTTL(ttl)
				return txn.SetEntry(entry)
			})

			if err != nil {
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Error("Failed to set token in BadgerDB")
				return fmt.Errorf("failed to set token in BadgerDB: %w", err)
			}

			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
				"ttl":       ttl.String(),
			}).Debug("Token set successfully in BadgerDB")
			return nil
		}
	case "redis":
		if app.redisClient != nil {
			// 将数据序列化为 JSON
			var value string
			if data != nil {
				valueBytes, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to marshal token data: %w", err)
				}
				value = string(valueBytes)
			} else {
				value = "1" // 如果没有数据，存储一个简单标记
			}

			// 解析 TTL
			var ttl time.Duration
			if config.Cache.Redis.TTL != "" {
				var err error
				ttl, err = time.ParseDuration(config.Cache.Redis.TTL)
				if err != nil {
					app.logger.WithError(err).Warn("Invalid Redis TTL, using default 24h")
					ttl = 24 * time.Hour
				}
			} else {
				ttl = 24 * time.Hour // 默认 24 小时
			}

			// 存储到 Redis
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			err := app.redisClient.Set(ctx, cacheKey, value, ttl).Err()
			if err != nil {
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Error("Failed to set token in Redis")
				return fmt.Errorf("failed to set token in Redis: %w", err)
			}

			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
				"ttl":       ttl.String(),
			}).Debug("Token set successfully in Redis")
			return nil
		}
	}

	return fmt.Errorf("no valid cache strategy configured for token storage")
}

// RemoveToken 从缓存中删除 token
// 这个方法可以在用户登出时调用，使 token 失效
func (app *App) RemoveToken(token string) error {
	if app.cfg.ModConfig == nil || !app.cfg.ModConfig.Token.Validation.Enabled {
		return nil
	}

	config := app.cfg.ModConfig.Token.Validation
	cacheKey := config.CacheKeyPrefix + token

	switch config.CacheStrategy {
	case "bigcache":
		if app.tokenCache != nil {
			err := app.tokenCache.Delete(cacheKey)
			if err != nil && err != bigcache.ErrEntryNotFound {
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Error("Failed to remove token from BigCache")
				return fmt.Errorf("failed to remove token from BigCache: %w", err)
			}

			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
			}).Debug("Token removed successfully from BigCache")
			return nil
		}
	case "badger":
		if app.badgerDB != nil {
			err := app.badgerDB.Update(func(txn *badger.Txn) error {
				return txn.Delete([]byte(cacheKey))
			})

			if err != nil && err != badger.ErrKeyNotFound {
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Error("Failed to remove token from BadgerDB")
				return fmt.Errorf("failed to remove token from BadgerDB: %w", err)
			}

			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
			}).Debug("Token removed successfully from BadgerDB")
			return nil
		}
	case "redis":
		if app.redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			deleted, err := app.redisClient.Del(ctx, cacheKey).Result()
			if err != nil {
				app.logger.WithFields(logrus.Fields{
					"token":     token,
					"cache_key": cacheKey,
					"error":     err.Error(),
				}).Error("Failed to remove token from Redis")
				return fmt.Errorf("failed to remove token from Redis: %w", err)
			}

			app.logger.WithFields(logrus.Fields{
				"token":     token,
				"cache_key": cacheKey,
				"deleted":   deleted,
			}).Debug("Token removed successfully from Redis")
			return nil
		}
	}

	return fmt.Errorf("no valid cache strategy configured for token removal")
}

// GetTokenData 从缓存中获取 token 相关的数据
// 这个方法可以用来获取存储在 token 中的用户信息等数据
func (app *App) GetTokenData(token string) ([]byte, error) {
	if app.cfg.ModConfig == nil || !app.cfg.ModConfig.Token.Validation.Enabled {
		return nil, fmt.Errorf("token validation not enabled")
	}

	config := app.cfg.ModConfig.Token.Validation
	cacheKey := config.CacheKeyPrefix + token

	switch config.CacheStrategy {
	case "bigcache":
		if app.tokenCache != nil {
			data, err := app.tokenCache.Get(cacheKey)
			if err != nil {
				if err == bigcache.ErrEntryNotFound {
					return nil, fmt.Errorf("token not found")
				}
				return nil, fmt.Errorf("failed to get token data from BigCache: %w", err)
			}
			return data, nil
		}
	case "badger":
		if app.badgerDB != nil {
			var data []byte
			err := app.badgerDB.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte(cacheKey))
				if err != nil {
					return err
				}
				return item.Value(func(val []byte) error {
					data = append([]byte(nil), val...) // 复制数据
					return nil
				})
			})

			if err != nil {
				if err == badger.ErrKeyNotFound {
					return nil, fmt.Errorf("token not found")
				}
				return nil, fmt.Errorf("failed to get token data from BadgerDB: %w", err)
			}
			return data, nil
		}
	case "redis":
		if app.redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			val, err := app.redisClient.Get(ctx, cacheKey).Result()
			if err != nil {
				if err == redis.Nil {
					return nil, fmt.Errorf("token not found")
				}
				return nil, fmt.Errorf("failed to get token data from Redis: %w", err)
			}
			return []byte(val), nil
		}
	}

	return nil, fmt.Errorf("no valid cache strategy configured for token data retrieval")
}

// Close 关闭应用时释放资源
func (app *App) Close() error {
	var errors []error

	// 关闭 BadgerDB
	if app.badgerDB != nil {
		if err := app.badgerDB.Close(); err != nil {
			app.logger.WithError(err).Error("Failed to close BadgerDB")
			errors = append(errors, fmt.Errorf("failed to close BadgerDB: %w", err))
		} else {
			app.logger.Info("BadgerDB closed successfully")
		}
	}

	// 关闭 Redis 客户端
	if app.redisClient != nil {
		if err := app.redisClient.Close(); err != nil {
			app.logger.WithError(err).Error("Failed to close Redis client")
			errors = append(errors, fmt.Errorf("failed to close Redis client: %w", err))
		} else {
			app.logger.Info("Redis client closed successfully")
		}
	}

	// 关闭 BigCache（BigCache v3 会自动清理，无需手动关闭）

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing app: %v", errors)
	}

	return nil
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
	Name          string
	Type          string
	Description   string
	Required      bool
	From          string // query, header, form, param
	Tag           string
	Level         int        // 嵌套层级，0为顶层
	Parent        string     // 父字段名
	Children      []DocField // 子字段（用于对象类型）
	IsObject      bool       // 是否为对象类型
	IsArray       bool       // 是否为数组类型
	ArrayItemType string     // 数组元素类型
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

// DocData contains all documentation data including app info and service groups
type DocData struct {
	AppInfo struct {
		Name        string
		DisplayName string
		Description string
		Version     string
	}
	Groups []DocGroup
}

// 处理文档请求
func (app *App) handleDocs(c *fiber.Ctx) error {
	// 按组分类并排序服务
	groups := app.groupAndSortServices()

	// 准备文档数据
	docData := DocData{
		Groups: groups,
	}

	// 设置应用信息
	docData.AppInfo.Name = app.cfg.Name
	docData.AppInfo.DisplayName = app.cfg.DisplayName
	docData.AppInfo.Description = app.cfg.Description

	// 如果有mod配置，优先使用mod配置中的信息
	if modConfig := app.cfg.ModConfig; modConfig != nil {
		if modConfig.App.Name != "" {
			docData.AppInfo.Name = modConfig.App.Name
		}
		if modConfig.App.DisplayName != "" {
			docData.AppInfo.DisplayName = modConfig.App.DisplayName
		}
		if modConfig.App.Description != "" {
			docData.AppInfo.Description = modConfig.App.Description
		}
		if modConfig.App.Version != "" {
			docData.AppInfo.Version = modConfig.App.Version
		}
	}

	// 设置默认值
	if docData.AppInfo.DisplayName == "" {
		docData.AppInfo.DisplayName = "API 文档"
	}

	// 生成HTML
	html := app.generateDocsHTML(docData)

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
	return app.parseStructFieldsRecursive(t, 0, "")
}

// 递归解析结构体字段
func (app *App) parseStructFieldsRecursive(t reflect.Type, level int, parentPath string) []DocField {
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
			Name:     field.Name,
			Type:     app.getFieldTypeString(field.Type),
			Level:    level,
			Parent:   parentPath,
			IsObject: false,
			IsArray:  false,
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

		// 分析字段类型，处理嵌套结构
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		currentPath := docField.Name
		if parentPath != "" {
			currentPath = parentPath + "." + docField.Name
		}

		switch fieldType.Kind() {
		case reflect.Struct:
			// 检查是否为基本类型的结构体（如time.Time等）
			if app.isBasicStructType(fieldType) {
				docField.Type = fieldType.Name()
			} else {
				docField.IsObject = true
				docField.Type = "object"
				// 递归解析子字段
				docField.Children = app.parseStructFieldsRecursive(fieldType, level+1, currentPath)
			}

		case reflect.Slice, reflect.Array:
			docField.IsArray = true
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}

			if elemType.Kind() == reflect.Struct && !app.isBasicStructType(elemType) {
				docField.Type = "array<object>"
				docField.ArrayItemType = "object"
				// 直接将数组元素的字段作为子字段，不增加 [item] 层级
				docField.Children = app.parseStructFieldsRecursive(elemType, level+1, currentPath)
			} else {
				elemTypeName := app.getFieldTypeString(elemType)
				docField.Type = "array<" + elemTypeName + ">"
				docField.ArrayItemType = elemTypeName
			}

		case reflect.Map:
			keyType := app.getFieldTypeString(fieldType.Key())
			valueType := fieldType.Elem()
			if valueType.Kind() == reflect.Interface && valueType.String() == "interface {}" {
				docField.Type = "map<" + keyType + ", any>"
			} else {
				valueTypeName := app.getFieldTypeString(valueType)
				docField.Type = "map<" + keyType + ", " + valueTypeName + ">"
			}
		}

		fields = append(fields, docField)
	}

	return fields
}

// 检查是否为基本类型的结构体
func (app *App) isBasicStructType(t reflect.Type) bool {
	basicStructs := map[string]bool{
		"time.Time": true,
		"Time":      true,
	}
	return basicStructs[t.String()] || basicStructs[t.Name()]
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
func (app *App) generateDocsHTML(docData DocData) string {
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.AppInfo.DisplayName}}{{if .AppInfo.Version}} v{{.AppInfo.Version}}{{end}}</title>
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
            flex-direction: column;
        }

        .top-header {
            position: fixed;
            top: 0;
            left: 299px;
            right: 0;
            height: 75px;
            background: #001529;
            border-bottom: none;
            z-index: 1001;
            display: flex;
            align-items: center;
            padding: 0 24px;
            transition: left 0.3s ease;
			height: 66px;
    		box-sizing: border-box;
        }

        .top-header.sidebar-collapsed {
            left: 0;
        }

        .menu-toggle {
            background: #001529;
            border: none;
            border-radius: 4px;
            padding: 8px;
            cursor: pointer;
            transition: all 0.3s;
            color: #fff;
            display: flex;
            align-items: center;
            justify-content: center;
            width: 40px;
            height: 40px;
        }

        .menu-toggle:hover {
            background: #1890ff;
        }

        .menu-toggle-icon {
            width: 20px;
            height: 14px;
            position: relative;
            transform: rotate(0deg);
            transition: .3s ease-in-out;
        }

        .menu-toggle-icon span {
            display: block;
            position: absolute;
            height: 2px;
            width: 100%;
            background: #fff;
            border-radius: 1px;
            opacity: 1;
            left: 0;
            transform: rotate(0deg);
            transition: .25s ease-in-out;
        }

        .menu-toggle-icon span:nth-child(1) {
            top: 0px;
        }

        .menu-toggle-icon span:nth-child(2) {
            top: 6px;
        }

        .menu-toggle-icon span:nth-child(3) {
            top: 12px;
        }

        .menu-toggle.open .menu-toggle-icon span:nth-child(1) {
            top: 6px;
            transform: rotate(135deg);
        }

        .menu-toggle.open .menu-toggle-icon span:nth-child(2) {
            opacity: 0;
            left: -20px;
        }

        .menu-toggle.open .menu-toggle-icon span:nth-child(3) {
            top: 6px;
            transform: rotate(-135deg);
        }

        .sidebar {
            width: 300px;
            background: #fff;
            border-right: 1px solid #f0f0f0;
            position: fixed;
            height: 100vh;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
            display: flex;
            flex-direction: column;
            overflow: hidden;
            z-index: 1000;
            transition: transform 0.3s ease;
        }

        .sidebar.collapsed {
            transform: translateX(-100%);
        }

        .sidebar-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            z-index: 999;
            opacity: 0;
            visibility: hidden;
            transition: all 0.3s ease;
        }

        .sidebar-overlay.show {
            opacity: 1;
            visibility: visible;
        }

        .sidebar-header {
            padding: 16px 24px;
            background: #001529;
            color: #fff;
            flex-shrink: 0;
    		height: 66px;
    		box-sizing: border-box;
    		display: flex;
    		align-items: center;
        }

        .sidebar-content {
            flex: 1;
            overflow-y: auto;
            background: white;
        }

        .sidebar-header h1 {
            font-size: 16px;
            font-weight: 600;
            margin: 0;
        }

        .version {
            font-size: 12px;
            font-weight: 400;
            color: rgba(255, 255, 255, 0.8);
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
            margin-top: 75px;
            padding: 24px;
            overflow-y: auto;
            transition: margin-left 0.3s ease;
        }

        .main-content.sidebar-collapsed {
            margin-left: 0;
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
            border-radius: 4px;
            display: flex;
            align-items: center;
            margin-bottom: 12px;
            border: 1px solid rgba(255, 255, 255, 0.3);
            max-width: fit-content;
            overflow: hidden;
        }

        .path-text {
            padding: 4px 8px;
            flex: 1;
        }

        .copy-btn-path {
            padding: 4px 8px;
            margin: 0;
            border: none;
            border-left: 1px solid rgba(255, 255, 255, 0.3);
            border-radius: 0;
            background: rgba(255, 255, 255, 0.1);
        }

        .copy-btn-path:hover {
            background: rgba(255, 255, 255, 0.2);
        }

        .copy-btn {
            background: rgba(255, 255, 255, 0.2);
            border: 1px solid rgba(255, 255, 255, 0.3);
            border-radius: 4px;
            padding: 4px;
            color: rgba(255, 255, 255, 0.8);
            cursor: pointer;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .copy-btn:hover {
            background: rgba(255, 255, 255, 0.3);
            color: #fff;
        }

        .copy-btn.copied {
            background: #52c41a;
            color: #fff;
        }

        .copy-btn-small {
            padding: 2px;
            margin-left: 6px;
        }

        .meta-item {
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .api-meta {
            display: flex;
            gap: 24px;
            flex-wrap: wrap;
            font-size: 12px;
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

        .auth-status-badge {
            font-weight: 500;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 11px;
            border: 1px solid;
        }

        .auth-required {
            background: #fff2f0;
            color: #ff4d4f;
            border-color: #ffccc7;
        }

        .auth-not-required {
            background: #f6ffed;
            color: #52c41a;
            border-color: #b7eb8f;
        }

        .meta-value-box {
            display: flex;
            align-items: center;
            background: rgba(255, 255, 255, 0.15);
            border-radius: 4px;
            border: 1px solid rgba(255, 255, 255, 0.2);
            overflow: hidden;
        }

        .meta-value-text {
            font-weight: 500;
            padding: 2px 6px;
            flex: 1;
        }

        .copy-btn-inline {
            padding: 2px 6px;
            margin: 0;
            border: none;
            border-left: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 0;
            background: rgba(255, 255, 255, 0.1);
        }

        .copy-btn-inline:hover {
            background: rgba(255, 255, 255, 0.2);
        }

        .api-description {
            margin-top: 12px;
            font-size: 13px;
            color: rgba(255, 255, 255, 0.85);
            line-height: 1.5;
            font-style: italic;
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
            padding: 8px 12px;
            text-align: left;
            border-bottom: 1px solid #f0f0f0;
        }

        .params-table th {
            background: #fafafa;
            font-weight: 500;
            color: rgba(0, 0, 0, 0.85);
            font-size: 13px;
        }

        .params-table td {
            font-size: 13px;
            color: rgba(0, 0, 0, 0.85);
        }

        .field-name-box {
            display: flex;
            align-items: center;
            gap: 4px;
        }

        .field-name {
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
            font-weight: 600;
            color: #1890ff;
        }

        .copy-btn-field {
            padding: 2px;
            margin: 0;
            border: 1px solid #d9d9d9;
            border-radius: 2px;
            background: #fafafa;
            color: rgba(0, 0, 0, 0.45);
            flex-shrink: 0;
        }

        .copy-btn-field:hover {
            background: #f0f0f0;
            color: #1890ff;
            border-color: #40a9ff;
        }

        .copy-btn-field.copied {
            background: #52c41a;
            color: #fff;
            border-color: #52c41a;
        }

        .params-table tr:last-child td {
            border-bottom: none;
        }

        .params-table tr:hover {
            background: #fafafa;
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

        .nested-field {
            border-left: 2px solid #e8f4ff;
            margin-left: 10px;
            padding-left: 10px;
        }

        .nested-field.level-1 {
            border-left-color: #bae7ff;
        }

        .nested-field.level-2 {
            border-left-color: #91d5ff;
        }

        .nested-field.level-3 {
            border-left-color: #69c0ff;
        }

        .field-path {
            color: rgba(0, 0, 0, 0.45);
            font-size: 11px;
            margin-left: 8px;
            font-style: italic;
        }

        .expand-btn {
            border: none;
            background: none;
            color: #1890ff;
            cursor: pointer;
            padding: 0 4px;
            font-size: 12px;
            margin-right: 4px;
            width: 16px;
            text-align: center;
        }

        .expand-btn:hover {
            background: #f0f8ff;
        }

        .expand-btn-placeholder {
            width: 16px;
            margin-right: 4px;
            display: inline-block;
        }

        .nested-table {
            margin-top: 8px;
            border: 1px solid #f0f0f0;
            border-radius: 4px;
        }

        .nested-table .params-table {
            margin: 0;
            border: none;
        }

        .nested-table .params-table th {
            background: #f8f9fa;
            font-size: 12px;
            padding: 6px 8px;
        }

        .nested-table .params-table td {
            font-size: 12px;
            padding: 6px 8px;
        }

        @media (max-width: 768px) {
            .top-header {
                left: 0;
                padding: 0 16px;
            }

            .menu-toggle {
                width: 36px;
                height: 36px;
            }

            .sidebar-overlay.show {
                display: block;
            }

            .main-content {
                margin-left: 0;
                padding: 16px;
            }

            .main-content.sidebar-collapsed {
                margin-left: 0;
            }

            .api-section {
                margin-bottom: 24px;
            }

            .api-header {
                padding: 12px 16px;
            }

            .api-title {
                font-size: 16px;
                margin-bottom: 6px;
            }

            .api-meta {
                flex-direction: column;
                gap: 8px;
                font-size: 11px;
            }

            .api-body {
                padding: 16px;
            }

            .params-table {
                font-size: 12px;
            }

            .params-table th,
            .params-table td {
                padding: 6px 8px;
            }

            .field-name-box {
                flex-direction: column;
                align-items: flex-start !important;
                gap: 4px;
            }

            .field-name {
                font-size: 13px;
                cursor: pointer;
                padding: 4px 8px;
                border-radius: 4px;
                transition: background-color 0.2s;
                display: inline-block;
            }

            .field-name:hover {
                background-color: rgba(24, 144, 255, 0.1);
                color: #1890ff;
            }

            .field-name:active {
                background-color: rgba(24, 144, 255, 0.2);
            }

            .field-type {
                font-size: 11px;
                padding: 1px 4px;
            }

            .copy-btn-field {
                display: none !important;
            }

            .field-path {
                font-size: 10px;
                margin-left: 0;
            }
        }

        @media (max-width: 480px) {
            .main-content {
                padding: 12px;
            }

            .api-header {
                padding: 10px 12px;
            }

            .api-body {
                padding: 12px;
            }

            .api-title {
                font-size: 14px;
            }

            .params-table th,
            .params-table td {
                padding: 4px 6px;
                font-size: 11px;
            }

            .field-name {
                font-size: 12px;
            }

            .api-path {
                font-size: 11px;
            }

            .meta-value,
            .meta-value-text {
                font-size: 10px;
                padding: 1px 4px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- 顶部固定区域 -->
        <div class="top-header">
            <!-- 汉堡包菜单按钮 -->
            <button class="menu-toggle" id="menuToggle" onclick="toggleSidebar()">
                <div class="menu-toggle-icon">
                    <span></span>
                    <span></span>
                    <span></span>
                </div>
            </button>
        </div>

        <!-- 侧边栏遮罩层 -->
        <div class="sidebar-overlay" id="sidebarOverlay" onclick="closeSidebar()"></div>

        <div class="sidebar" id="sidebar">
            <div class="sidebar-header">
                <h1>{{.AppInfo.DisplayName}}</h1>
                {{if .AppInfo.Version}}<div class="version">v{{.AppInfo.Version}}</div>{{end}}
            </div>
            <div class="sidebar-content">
                {{range .Groups}}
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
        </div>

        <div class="main-content" id="mainContent">
            {{range .Groups}}
            {{range .Services}}
            <div class="api-section" id="service-{{.Name}}">
                <div class="api-header">
                    <div class="api-title">{{.DisplayName}}</div>
                    <div class="api-path">
                        <span class="path-text">POST {{.ServicePath}}</span>
                        <button class="copy-btn copy-btn-path" onclick="copyToClipboard('{{.ServicePath}}', this)" title="复制接口地址">
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                            </svg>
                        </button>
                    </div>
                    <div class="api-meta">
                        <div class="meta-item">
                            <span class="meta-label">服务名称:</span>
                            <div class="meta-value-box">
                                <span class="meta-value-text">{{.Name}}</span>
                                <button class="copy-btn copy-btn-inline" onclick="copyToClipboard('{{.Name}}', this)" title="复制服务名称">
                                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                                    </svg>
                                </button>
                            </div>
                        </div>
                        <div class="meta-item">
                            <span class="meta-label">认证:</span>
                            <span class="meta-value auth-status-badge {{if .SkipAuth}}auth-not-required{{else}}auth-required{{end}}">{{if .SkipAuth}}不需要{{else}}需要{{end}}</span>
                        </div>
                        <div class="meta-item">
                            <span class="meta-label">返回格式:</span>
                            <span class="meta-value auth-status-badge {{if .ReturnRaw}}auth-not-required{{else}}auth-required{{end}}">{{if .ReturnRaw}}原始格式{{else}}标准格式{{end}}</span>
                        </div>
                    </div>
                    {{if .Description}}
                    <div class="api-description">{{.Description}}</div>
                    {{end}}
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
                                {{template "renderField" .}}
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
                        <div class="section-title">返回参数{{if not .ReturnRaw}} (标准格式){{else}} (原始格式){{end}}</div>
                        {{if not .ReturnRaw}}
                        <div class="return-format-note">
                            <div style="margin-bottom: 12px; padding: 8px; background: #f6ffed; border: 1px solid #b7eb8f; border-radius: 4px; font-size: 12px; color: #52c41a;">
                                <strong>标准返回格式：</strong>返回数据被包装在统一的响应结构中
                            </div>
                        </div>
                        <table class="params-table">
                            <thead>
                                <tr>
                                    <th>参数名</th>
                                    <th>类型</th>
                                    <th>描述</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>
                                        <div class="field-name-box">
                                            <span class="expand-btn-placeholder"></span>
                                            <span class="field-name">code</span>
                                        </div>
                                    </td>
                                    <td><span class="field-type">int</span></td>
                                    <td>响应状态码，0表示成功</td>
                                </tr>
                                <tr>
                                    <td>
                                        <div class="field-name-box">
                                            <span class="expand-btn-placeholder"></span>
                                            <span class="field-name">msg</span>
                                        </div>
                                    </td>
                                    <td><span class="field-type">string</span></td>
                                    <td>响应消息</td>
                                </tr>
                                <tr>
                                    <td>
                                        <div class="field-name-box">
                                            {{if .OutputFields}}
                                            <button class="expand-btn" onclick="toggleNested(this)">+</button>
                                            {{else}}
                                            <span class="expand-btn-placeholder"></span>
                                            {{end}}
                                            <span class="field-name">data</span>
                                        </div>
                                    </td>
                                    <td><span class="field-type">object</span></td>
                                    <td>实际业务数据</td>
                                </tr>
                                {{range .OutputFields}}
                                {{template "renderOutputFieldNested" .}}
                                {{end}}
                                <tr>
                                    <td>
                                        <div class="field-name-box">
                                            <span class="expand-btn-placeholder"></span>
                                            <span class="field-name">rid</span>
                                        </div>
                                    </td>
                                    <td><span class="field-type">string</span></td>
                                    <td>请求ID</td>
                                </tr>
                                <tr style="display: none;">
                                    <td>
                                        <div class="field-name-box">
                                            <span class="expand-btn-placeholder"></span>
                                            <span class="field-name">detail</span>
                                        </div>
                                    </td>
                                    <td><span class="field-type">string</span></td>
                                    <td>错误详情（仅错误时存在）</td>
                                </tr>
                            </tbody>
                        </table>
                        {{else}}
                        <div class="return-format-note">
                            <div style="margin-bottom: 12px; padding: 8px; background: #fff7e6; border: 1px solid #ffd591; border-radius: 4px; font-size: 12px; color: #fa8c16;">
                                <strong>原始返回格式：</strong>直接返回业务数据，不包装在标准响应结构中
                            </div>
                        </div>
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
                                {{template "renderOutputField" .}}
                                {{end}}
                            </tbody>
                        </table>
                        {{end}}
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
        function copyToClipboard(text, button) {
            navigator.clipboard.writeText(text).then(function() {
                // 复制成功的视觉反馈
                const originalClass = button.className;
                button.classList.add('copied');

                // 临时显示复制成功状态
                setTimeout(function() {
                    button.className = originalClass;
                }, 1500);
            }).catch(function(err) {
                // 降级处理：使用传统方法
                const textArea = document.createElement('textarea');
                textArea.value = text;
                document.body.appendChild(textArea);
                textArea.focus();
                textArea.select();
                try {
                    document.execCommand('copy');
                    const originalClass = button.className;
                    button.classList.add('copied');
                    setTimeout(function() {
                        button.className = originalClass;
                    }, 1500);
                } catch (err) {
                    console.error('复制失败:', err);
                }
                document.body.removeChild(textArea);
            });
        }

        // 移动端参数名点击复制功能
        function copyFieldName(text, element) {
            // 检查是否为移动端
            if (window.innerWidth <= 768) {
                // 创建临时的视觉反馈
                const originalBg = element.style.backgroundColor;
                const originalColor = element.style.color;

                // 设置复制成功的视觉效果
                element.style.backgroundColor = '#52c41a';
                element.style.color = '#fff';

                // 执行复制
                navigator.clipboard.writeText(text).then(function() {
                    // 1.5秒后恢复原样
                    setTimeout(function() {
                        element.style.backgroundColor = originalBg;
                        element.style.color = originalColor;
                    }, 1500);
                }).catch(function(err) {
                    // 降级处理
                    const textArea = document.createElement('textarea');
                    textArea.value = text;
                    document.body.appendChild(textArea);
                    textArea.focus();
                    textArea.select();
                    try {
                        document.execCommand('copy');
                        setTimeout(function() {
                            element.style.backgroundColor = originalBg;
                            element.style.color = originalColor;
                        }, 1500);
                    } catch (err) {
                        console.error('复制失败:', err);
                        element.style.backgroundColor = originalBg;
                        element.style.color = originalColor;
                    }
                    document.body.removeChild(textArea);
                });
            }
        }

        function scrollToService(serviceId) {
            const element = document.getElementById(serviceId);
            if (element) {
                element.scrollIntoView({ behavior: 'smooth', block: 'start' });

                // 更新激活状态
                document.querySelectorAll('.service-item').forEach(item => {
                    item.classList.remove('active');
                });
                event.target.classList.add('active');

                // 移动端自动关闭侧边栏
                if (window.innerWidth <= 768) {
                    closeSidebar();
                }
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
                // 只有当current不为空且匹配时才添加active类
                if (current && item.getAttribute('onclick').includes(current)) {
                    item.classList.add('active');
                }
            });
        }

        window.addEventListener('scroll', updateActiveService);
        document.addEventListener('DOMContentLoaded', updateActiveService);

        // 切换侧边栏显示/隐藏
        function toggleSidebar() {
            const sidebar = document.getElementById('sidebar');
            const menuToggle = document.getElementById('menuToggle');
            const overlay = document.getElementById('sidebarOverlay');
            const mainContent = document.getElementById('mainContent');
            const topHeader = document.querySelector('.top-header');

            const isCollapsed = sidebar.classList.contains('collapsed');

            if (isCollapsed) {
                // 显示侧边栏
                sidebar.classList.remove('collapsed');
                menuToggle.classList.add('open');
                mainContent.classList.remove('sidebar-collapsed');
                topHeader.classList.remove('sidebar-collapsed');

                // 移动端显示遮罩层
                if (window.innerWidth <= 768) {
                    overlay.classList.add('show');
                }
            } else {
                // 隐藏侧边栏
                closeSidebar();
            }
        }

        // 关闭侧边栏
        function closeSidebar() {
            const sidebar = document.getElementById('sidebar');
            const menuToggle = document.getElementById('menuToggle');
            const overlay = document.getElementById('sidebarOverlay');
            const mainContent = document.getElementById('mainContent');
            const topHeader = document.querySelector('.top-header');

            sidebar.classList.add('collapsed');
            menuToggle.classList.remove('open');
            mainContent.classList.add('sidebar-collapsed');
            topHeader.classList.add('sidebar-collapsed');
            overlay.classList.remove('show');
        }

        // 窗口大小变化时的处理
        window.addEventListener('resize', function() {
            const sidebar = document.getElementById('sidebar');
            const overlay = document.getElementById('sidebarOverlay');

            if (window.innerWidth > 768) {
                // 桌面端隐藏遮罩层
                overlay.classList.remove('show');
            } else {
                // 移动端如果侧边栏显示，则显示遮罩层
                if (!sidebar.classList.contains('collapsed')) {
                    overlay.classList.add('show');
                }
            }
        });

        // 初始化状态 - 默认展开侧边栏
        document.addEventListener('DOMContentLoaded', function() {
            const sidebar = document.getElementById('sidebar');
            const mainContent = document.getElementById('mainContent');
            const topHeader = document.querySelector('.top-header');
            const menuToggle = document.getElementById('menuToggle');

            // 默认状态是展开的
            sidebar.classList.remove('collapsed');
            mainContent.classList.remove('sidebar-collapsed');
            topHeader.classList.remove('sidebar-collapsed');
            menuToggle.classList.add('open'); // 设置菜单按钮为打开状态
        });

        // 展开/折叠嵌套字段
        function toggleNested(button) {
            const row = button.closest('tr');
            const currentLevel = parseInt(row.className.match(/level-(\d+)/)?.[1] || '0');
            const nextRows = [];
            let currentRow = row.nextElementSibling;

            // 只收集直接子级行（下一级别）
            while (currentRow && currentRow.classList.contains('nested-row')) {
                const rowLevel = parseInt(currentRow.className.match(/level-(\d+)/)?.[1] || '0');
                if (rowLevel === currentLevel + 1) {
                    nextRows.push(currentRow);
                } else if (rowLevel <= currentLevel) {
                    break;
                }
                currentRow = currentRow.nextElementSibling;
            }

            const isExpanded = button.textContent === '−';
            button.textContent = isExpanded ? '+' : '−';

            nextRows.forEach(r => {
                if (isExpanded) {
                    // 折叠时，隐藏直接子级并递归折叠其所有子级
                    r.style.display = 'none';
                    collapseAllChildren(r);
                } else {
                    // 展开时，只显示直接子级
                    r.style.display = '';
                }
            });
        }

        // 递归折叠所有子级
        function collapseAllChildren(parentRow) {
            const parentLevel = parseInt(parentRow.className.match(/level-(\d+)/)?.[1] || '0');
            let currentRow = parentRow.nextElementSibling;

            while (currentRow && currentRow.classList.contains('nested-row')) {
                const rowLevel = parseInt(currentRow.className.match(/level-(\d+)/)?.[1] || '0');
                if (rowLevel <= parentLevel) {
                    break;
                }

                // 隐藏所有更深层级的行
                currentRow.style.display = 'none';

                // 将展开按钮重置为+状态
                const expandBtn = currentRow.querySelector('.expand-btn');
                if (expandBtn) {
                    expandBtn.textContent = '+';
                }

                currentRow = currentRow.nextElementSibling;
            }
        }
    </script>

    <!-- 模板定义 -->
    {{define "renderField"}}
    <tr {{if gt .Level 0}}class="nested-row nested-field level-{{.Level}}" style="display: none;"{{end}}>
        <td>
            <div class="field-name-box" style="margin-left: {{mul .Level 20}}px;">
                {{if .Children}}
                <button class="expand-btn" onclick="toggleNested(this)">+</button>
                {{else}}
                <span class="expand-btn-placeholder"></span>
                {{end}}
                <span class="field-name" onclick="copyFieldName('{{.Name}}', this)" title="点击复制参数名">{{.Name}}</span>
                {{if .Parent}}<span class="field-path">({{.Parent}})</span>{{end}}
                <button class="copy-btn copy-btn-field" onclick="copyToClipboard('{{.Name}}', this)" title="复制参数名">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                    </svg>
                </button>
            </div>
        </td>
        <td><span class="field-type">{{.Type}}</span></td>
        <td><span class="from-tag">{{.From}}</span></td>
        <td><span class="{{if .Required}}required{{else}}not-required{{end}}">{{if .Required}}是{{else}}否{{end}}</span></td>
        <td>{{if .Description}}{{.Description}}{{else}}-{{end}}</td>
    </tr>
    {{range .Children}}
    {{template "renderField" .}}
    {{end}}
    {{end}}

    {{define "renderOutputField"}}
    <tr {{if gt .Level 0}}class="nested-row nested-field level-{{.Level}}" style="display: none;"{{end}}>
        <td>
            <div class="field-name-box" style="margin-left: {{mul .Level 20}}px;">
                {{if .Children}}
                <button class="expand-btn" onclick="toggleNested(this)">+</button>
                {{else}}
                <span class="expand-btn-placeholder"></span>
                {{end}}
                <span class="field-name" onclick="copyFieldName('{{.Name}}', this)" title="点击复制参数名">{{.Name}}</span>
                {{if .Parent}}<span class="field-path">({{.Parent}})</span>{{end}}
                <button class="copy-btn copy-btn-field" onclick="copyToClipboard('{{.Name}}', this)" title="复制参数名">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                    </svg>
                </button>
            </div>
        </td>
        <td><span class="field-type">{{.Type}}</span></td>
        <td>{{if .Description}}{{.Description}}{{else}}-{{end}}</td>
    </tr>
    {{range .Children}}
    {{template "renderOutputField" .}}
    {{end}}
    {{end}}

    {{define "renderOutputFieldNested"}}
    <tr class="nested-row nested-field level-1" style="display: none;">
        <td>
            <div class="field-name-box" style="margin-left: 20px;">
                {{if .Children}}
                <button class="expand-btn" onclick="toggleNested(this)">+</button>
                {{else}}
                <span class="expand-btn-placeholder"></span>
                {{end}}
                <span class="field-name" onclick="copyFieldName('{{.Name}}', this)" title="点击复制参数名">{{.Name}}</span>
                <button class="copy-btn copy-btn-field" onclick="copyToClipboard('{{.Name}}', this)" title="复制参数名">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                    </svg>
                </button>
            </div>
        </td>
        <td><span class="field-type">{{.Type}}</span></td>
        <td>{{if .Description}}{{.Description}}{{else}}-{{end}}</td>
    </tr>
    {{range .Children}}
    {{template "renderOutputFieldNestedChild" .}}
    {{end}}
    {{end}}

    {{define "renderOutputFieldNestedChild"}}
    <tr class="nested-row nested-field level-{{add .Level 1}}" style="display: none;">
        <td>
            <div class="field-name-box" style="margin-left: {{mul (add .Level 1) 20}}px;">
                {{if .Children}}
                <button class="expand-btn" onclick="toggleNested(this)">+</button>
                {{else}}
                <span class="expand-btn-placeholder"></span>
                {{end}}
                <span class="field-name" onclick="copyFieldName('{{.Name}}', this)" title="点击复制参数名">{{.Name}}</span>
                {{if .Parent}}<span class="field-path">({{.Parent}})</span>{{end}}
                <button class="copy-btn copy-btn-field" onclick="copyToClipboard('{{.Name}}', this)" title="复制参数名">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                    </svg>
                </button>
            </div>
        </td>
        <td><span class="field-type">{{.Type}}</span></td>
        <td>{{if .Description}}{{.Description}}{{else}}-{{end}}</td>
    </tr>
    {{range .Children}}
    {{template "renderOutputFieldNestedChild" .}}
    {{end}}
    {{end}}

</body>
</html>`

	// 创建模板函数映射
	funcMap := template.FuncMap{
		"mul": func(a, b int) int { return a * b },
		"gt":  func(a, b int) bool { return a > b },
		"add": func(a, b int) int { return a + b },
	}

	t := template.Must(template.New("docs").Funcs(funcMap).Parse(tmpl))
	var buf strings.Builder
	t.Execute(&buf, docData)
	return buf.String()
}
