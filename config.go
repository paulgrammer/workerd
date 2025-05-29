package workerd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jinzhu/configor"
)

type RedisClient struct {
	// Network type to use, either tcp or unix.
	// Default is tcp.
	Network string `json:"network" yaml:"network" env:"ASYNQ_REDIS_NETWORK" default:"tcp"`

	// Redis server address in "host:port" format.
	Addr string `json:"address" yaml:"address" env:"ASYNQ_REDIS_ADDRESS" default:"127.0.0.1:6379"`

	// Username to authenticate the current connection when Redis ACLs are used.
	Username string `json:"username" yaml:"username" env:"ASYNQ_REDIS_USERNAME"`

	// Password to authenticate the current connection.
	Password string `json:"password" yaml:"password" env:"ASYNQ_REDIS_PASSWORD"`

	// Redis DB to select after connecting to a server.
	DB int `json:"db" yaml:"db" env:"ASYNQ_REDIS_DB" default:"0"`

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration `json:"dialTimeout" yaml:"dialTimeout" env:"ASYNQ_REDIS_DIAL_TIMEOUT" default:"5s"`

	// Timeout for socket reads. Default is 3 seconds.
	ReadTimeout time.Duration `json:"readTimeout" yaml:"readTimeout" env:"ASYNQ_REDIS_READ_TIMEOUT" default:"3s"`

	// Timeout for socket writes. Default is equal to ReadTimeout.
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout" env:"ASYNQ_REDIS_WRITE_TIMEOUT" default:"3s"`

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU.
	PoolSize int `json:"poolSize" yaml:"poolSize" env:"ASYNQ_REDIS_POOL_SIZE" default:"10"`
}

type AsynqConfig struct {
	RedisClient RedisClient `json:"redisClient" yaml:"redisClient" required:"true"`
}

func (a *AsynqConfig) GetRedisClientOpt() (*asynq.RedisClientOpt, error) {
	if a == nil {
		return nil, fmt.Errorf("AsynqConfig is nil")
	}

	if err := a.validate(); err != nil {
		return nil, fmt.Errorf("invalid asynq configuration: %w", err)
	}

	return &asynq.RedisClientOpt{
		Network:      a.RedisClient.Network,
		Addr:         a.RedisClient.Addr,
		Username:     a.RedisClient.Username,
		Password:     a.RedisClient.Password,
		DB:           a.RedisClient.DB,
		DialTimeout:  a.RedisClient.DialTimeout,
		ReadTimeout:  a.RedisClient.ReadTimeout,
		WriteTimeout: a.RedisClient.WriteTimeout,
		PoolSize:     a.RedisClient.PoolSize,
	}, nil
}

// validate validates the AsynqConfig and its RedisClient configuration
func (a *AsynqConfig) validate() error {
	if a.RedisClient.Addr == "" {
		return fmt.Errorf("redis address cannot be empty")
	}
	if a.RedisClient.Network == "" {
		return fmt.Errorf("redis network cannot be empty")
	}
	if a.RedisClient.DB < 0 {
		return fmt.Errorf("redis DB must be non-negative, got %d", a.RedisClient.DB)
	}
	if a.RedisClient.PoolSize <= 0 {
		return fmt.Errorf("redis pool size must be positive, got %d", a.RedisClient.PoolSize)
	}
	if a.RedisClient.DialTimeout <= 0 {
		return fmt.Errorf("redis dial timeout must be positive, got %v", a.RedisClient.DialTimeout)
	}
	if a.RedisClient.ReadTimeout <= 0 {
		return fmt.Errorf("redis read timeout must be positive, got %v", a.RedisClient.ReadTimeout)
	}
	if a.RedisClient.WriteTimeout <= 0 {
		return fmt.Errorf("redis write timeout must be positive, got %v", a.RedisClient.WriteTimeout)
	}
	return nil
}

// workerConfig defines the workers's settings
type workerConfig struct {
	AsynqConfig *AsynqConfig `json:"asynq" yaml:"asynq"`
	LogLevel    slog.Level   `json:"loglevel" yaml:"loglevel" env:"LOG_LEVEL" default:"DEBUG"`
	Name        string       `json:"name" yaml:"name" env:"WORKER_NAME" default:"workerd"`
	DisplayName string       `json:"display_name" yaml:"display_name" env:"WORKER_DISPLAY_NAME" default:"Workerd Service"`
	Description string       `json:"description" yaml:"description" env:"WORKER_DESCRIPTION" default:"Default background worker service"`
	Concurrency int          `json:"concurrency" yaml:"concurrency" env:"WORKER_CONCURRENCY" default:"10"`
}

func newWorkerConfig(files ...string) (*workerConfig, error) {
	config := &workerConfig{
		AsynqConfig: new(AsynqConfig),
	}

	// Load configuration from files
	configorInstance := configor.New(&configor.Config{
		AutoReload:           false,
		Debug:                false,
		Silent:               false,
		Verbose:              false,
		ErrorOnUnmatchedKeys: true,
	})

	if len(files) > 0 {
		if err := configorInstance.Load(config, files...); err != nil {
			return nil, fmt.Errorf("failed to load configuration from files %v: %w", files, err)
		}
	} else {
		// Load from environment variables only
		if err := configorInstance.Load(config); err != nil {
			return nil, fmt.Errorf("failed to load configuration from environment: %w", err)
		}
	}

	// Validate loaded configuration
	if err := validateWorkerConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validateWorkerConfig validates the workerConfig
func validateWorkerConfig(config *workerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	if config.AsynqConfig == nil {
		return fmt.Errorf("asynq configuration is required")
	}

	if config.Concurrency < 0 {
		return fmt.Errorf("concurrency must be non-negative, got %d", config.Concurrency)
	}

	// Validate asynq config
	if err := config.AsynqConfig.validate(); err != nil {
		return fmt.Errorf("asynq configuration invalid: %w", err)
	}

	return nil
}
