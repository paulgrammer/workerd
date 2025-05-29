package workerd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/kardianos/service"
)

// Workerd represents the worker daemon
type Workerd struct {
	*asynq.ServeMux
	serviceFlag string
	srv         *asynq.Server
	config      *workerConfig
	log         *slog.Logger
	configPath  string
	name        string
	displayName string
	description string
	concurrency int
	errorChan   chan error
}

// === Functional Option Type ===
type Option func(*Workerd)

// === Option Functions ===
func WithLogger(logger *slog.Logger) Option {
	return func(w *Workerd) {
		w.log = logger
	}
}

func WithServeMux(mux *asynq.ServeMux) Option {
	return func(w *Workerd) {
		w.ServeMux = mux
	}
}

func WithName(name string) Option {
	return func(w *Workerd) {
		w.name = name
	}
}

func WithDisplayName(displayName string) Option {
	return func(w *Workerd) {
		w.displayName = displayName
	}
}

func WithDescription(desc string) Option {
	return func(w *Workerd) {
		w.description = desc
	}
}

func WithConcurrency(n int) Option {
	return func(w *Workerd) {
		w.concurrency = n
	}
}

func WithConfigPath(path string) Option {
	return func(w *Workerd) {
		w.configPath = path
	}
}

func WithServiceFlag(serviceFlag string) Option {
	return func(w *Workerd) {
		w.serviceFlag = serviceFlag
	}
}

// === Service Interface Implementation ===
func (w *Workerd) Start(s service.Service) error {
	w.log.Info("Workerd service starting...")

	// Start the asynq server
	if err := w.srv.Start(w.ServeMux); err != nil {
		w.log.Error("could not start asynq server", "error", err)
		return err
	}

	w.log.Info("Workerd service started successfully")
	return nil
}

func (w *Workerd) Stop(s service.Service) error {
	w.log.Info("Workerd service stopping...")
	w.srv.Shutdown()
	w.log.Info("Workerd service stopped")
	return nil
}

// === Utility Functions ===
func splitConfigPath(configPath string) []string {
	if len(configPath) == 0 {
		return []string{}
	}
	return strings.Split(configPath, ",")
}

// initializeComponents initializes the logger, ServeMux, and asynq server
func (w *Workerd) initializeComponents(config *workerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Initialize logger if not provided
	if w.log == nil {
		w.log = newLogger(config.LogLevel)
	}

	// Initialize ServeMux if not provided
	if w.ServeMux == nil {
		w.ServeMux = asynq.NewServeMux()
	}

	// Initialize asynq server using ServerBuilder
	serverBuilder, err := NewServerBuilder(config)
	if err != nil {
		return fmt.Errorf("failed to create server builder: %w", err)
	}

	w.srv, err = serverBuilder.BuildServer(w.concurrency)
	if err != nil {
		return fmt.Errorf("failed to build asynq server: %w", err)
	}

	return nil
}

// === Constructor ===
func NewWorkerd(opts ...Option) (*Workerd, error) {
	w := &Workerd{
		name:        "workerd",
		displayName: "Workerd Service",
		description: "Background worker service",
		concurrency: 10,
	}

	// Apply functional options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(w)
	}

	// Extract options into WorkerdOptions struct
	options := &WorkerdOptions{
		Name:        w.name,
		DisplayName: w.displayName,
		Description: w.description,
		Concurrency: w.concurrency,
		Logger:      w.log,
		ConfigPath:  w.configPath,
		ServiceFlag: w.serviceFlag,
	}

	// Load configuration
	fileConfig, err := newWorkerConfig(splitConfigPath(w.configPath)...)
	if err != nil {
		return nil, fmt.Errorf("failed to load worker config: %w", err)
	}

	// Create default configuration
	defaultConfig := &workerConfig{
		Name:        "workerd",
		DisplayName: "Workerd Service",
		Description: "Background worker service",
		Concurrency: 10,
	}

	// Merge configurations using ConfigMerger
	merger := NewConfigMerger().
		WithDefaults(defaultConfig).
		WithFileConfig(fileConfig).
		WithOptions(options)

	mergedConfig, err := merger.Merge()
	if err != nil {
		return nil, fmt.Errorf("failed to merge configurations: %w", err)
	}

	// Apply merged configuration to workerd instance
	w.name = mergedConfig.Name
	w.displayName = mergedConfig.DisplayName
	w.description = mergedConfig.Description
	w.concurrency = mergedConfig.Concurrency
	w.configPath = mergedConfig.ConfigPath
	w.serviceFlag = mergedConfig.ServiceFlag
	w.config = mergedConfig.Config

	// Use provided logger or create default
	if mergedConfig.Logger != nil {
		w.log = mergedConfig.Logger
	}

	// Initialize components
	if err := w.initializeComponents(mergedConfig.Config); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return w, nil
}

// GetLogger returns the logger instance
func (w *Workerd) GetLogger() *slog.Logger {
	return w.log
}

// Run is the main entry point that handles both service and standalone modes
func (w *Workerd) Run() error {
	// Initialize service manager
	serviceManager, err := NewServiceManager(w)
	if err != nil {
		return fmt.Errorf("failed to create service manager: %w", err)
	}

	// Handle service control
	return serviceManager.HandleControl(w.serviceFlag)
}
