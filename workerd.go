package workerd

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
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
	logger      service.Logger
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

// === Constructor ===
func NewWorkerd(opts ...Option) *Workerd {
	w := &Workerd{
		name:        "workerd",
		displayName: "Workerd Service",
		description: "Background worker service",
		concurrency: 10,
	}

	// Apply functional options
	for _, opt := range opts {
		opt(w)
	}

	// Load config
	config, err := newWorkerConfig(splitConfigPath(w.configPath)...)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	w.config = config

	// Apply config values if not set via options
	if w.name == "" && config.Name != "" {
		w.name = config.Name
	}
	if w.displayName == "" && config.DisplayName != "" {
		w.displayName = config.DisplayName
	}
	if w.description == "" && config.Description != "" {
		w.description = config.Description
	}
	if w.concurrency <= 0 && config.Concurrency > 0 {
		w.concurrency = config.Concurrency
	}
	if w.log == nil {
		w.log = newLogger(config.LogLevel)
	}
	if w.ServeMux == nil {
		w.ServeMux = asynq.NewServeMux()
	}

	w.srv = asynq.NewServer(config.AsynqConfig.GetRedisClientOpt(),
		asynq.Config{Concurrency: w.concurrency},
	)

	return w
}

// GetLogger returns the logger instance
func (w *Workerd) GetLogger() *slog.Logger {
	return w.log
}

// newService starts the workerd as a system service
func (w *Workerd) newService() (service.Service, error) {
	svcConfig := &service.Config{
		Name:        w.name,
		DisplayName: w.displayName,
		Description: w.description,
		Arguments:   []string{"-service", "run"},
	}

	if w.configPath != "" {
		svcConfig.Arguments = append(svcConfig.Arguments, "-config", w.configPath)
	}

	s, err := service.New(w, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	// Setup service logger
	errs := make(chan error, 5)
	w.logger, err = s.Logger(errs)
	if err != nil {
		return nil, fmt.Errorf("failed to create service logger: %w", err)
	}

	// Start error logging goroutine
	go w.logServiceErrors(errs)

	return s, nil
}

// HandleServiceControl handles service control commands
func (w *Workerd) HandleServiceControl(s service.Service, action string) error {
	switch action {
	case "run":
		err := s.Run()
		if err != nil {
			return fmt.Errorf("failed to run service: %w", err)
		}
	default:
		err := service.Control(s, action)
		if err != nil {
			return fmt.Errorf("service control action '%s' failed: %w (valid actions: %q)",
				action, err, service.ControlAction)
		}
	}

	return nil
}

// Run is the main entry point that handles both service and standalone modes
func (w *Workerd) Run() error {
	// Initialize service
	s, err := w.newService()
	if err != nil {
		return err
	}

	// Handle service control
	return w.HandleServiceControl(s, w.serviceFlag)
}

// logServiceErrors logs service errors
func (w *Workerd) logServiceErrors(errs chan error) {
	for {
		err := <-errs
		if err != nil {
			log.Print(err)
			if w.log != nil {
				w.log.Error("Service error", "error", err)
			}
		}
	}
}

// === CLI Integration ===

// Run runs the workerd with command line interface with mux
func Run(mux *asynq.ServeMux) {
	serviceFlag := flag.String("service", "run", "Control the system service (install, uninstall, start, stop, restart, run)")
	configPath := flag.String("config", "", "Path to either a file or directory to load configuration from")
	name := flag.String("name", "", "Service name")
	displayName := flag.String("display-name", "", "Service display name")
	description := flag.String("description", "", "Service description")
	concurrency := flag.Int("concurrency", 0, "Number of concurrent workers")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	// Build options
	opts := []Option{WithServeMux(mux)}

	if *configPath != "" {
		opts = append(opts, WithConfigPath(*configPath))
	}
	if *serviceFlag != "" {
		opts = append(opts, WithServiceFlag(*serviceFlag))
	}
	if *name != "" {
		opts = append(opts, WithName(*name))
	}
	if *displayName != "" {
		opts = append(opts, WithDisplayName(*displayName))
	}
	if *description != "" {
		opts = append(opts, WithDescription(*description))
	}
	if *concurrency > 0 {
		opts = append(opts, WithConcurrency(*concurrency))
	}

	// Create and run workerd
	workerd := NewWorkerd(opts...)

	if err := workerd.Run(); err != nil {
		fmt.Printf("Failed to run workerd: %v\n", err)
		os.Exit(1)
	}
}
