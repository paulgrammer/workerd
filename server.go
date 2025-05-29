package workerd

import (
	"fmt"

	"github.com/hibiken/asynq"
)

// ServerBuilder handles asynq server creation and configuration
type ServerBuilder struct {
	config *workerConfig
}

// NewServerBuilder creates a new server builder
func NewServerBuilder(config *workerConfig) (*ServerBuilder, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	return &ServerBuilder{config: config}, nil
}

// BuildServer creates and configures an asynq server
func (sb *ServerBuilder) BuildServer(concurrency int) (*asynq.Server, error) {
	if concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be positive, got %d", concurrency)
	}

	// Get Redis client options
	redisOpt, err := sb.config.AsynqConfig.GetRedisClientOpt()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis client options: %w", err)
	}

	// Create server configuration
	serverConfig := asynq.Config{
		Concurrency: concurrency,
		// Additional server configurations can be added here
	}

	// Create and return the server
	server := asynq.NewServer(redisOpt, serverConfig)
	if server == nil {
		return nil, fmt.Errorf("failed to create asynq server")
	}

	return server, nil
}

// BuildServerWithDefaults creates a server with default configuration
func (sb *ServerBuilder) BuildServerWithDefaults() (*asynq.Server, error) {
	defaultConcurrency := 10
	if sb.config.Concurrency > 0 {
		defaultConcurrency = sb.config.Concurrency
	}
	return sb.BuildServer(defaultConcurrency)
}

// ValidateServerConfig validates server configuration parameters
func (sb *ServerBuilder) ValidateServerConfig(concurrency int) error {
	if concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive, got %d", concurrency)
	}
	
	if sb.config.AsynqConfig == nil {
		return fmt.Errorf("asynq configuration is required")
	}

	// Validate Redis configuration
	if err := sb.config.AsynqConfig.validate(); err != nil {
		return fmt.Errorf("invalid asynq configuration: %w", err)
	}

	return nil
}