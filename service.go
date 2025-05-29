package workerd

import (
	"fmt"
	"log"

	"github.com/kardianos/service"
)

// ServiceManager handles system service management concerns
type ServiceManager struct {
	workerd *Workerd
	service service.Service
	logger  service.Logger
}

// NewServiceManager creates a new service manager for the given workerd instance
func NewServiceManager(w *Workerd) (*ServiceManager, error) {
	if w == nil {
		return nil, fmt.Errorf("workerd cannot be nil")
	}

	sm := &ServiceManager{
		workerd: w,
	}

	svc, err := sm.createService()
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}
	sm.service = svc

	logger, err := sm.createLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create service logger: %w", err)
	}
	sm.logger = logger

	// Start error logging goroutine
	go sm.startErrorLogging()

	return sm, nil
}

// createService creates the system service configuration
func (sm *ServiceManager) createService() (service.Service, error) {
	svcConfig := &service.Config{
		Name:        sm.workerd.name,
		DisplayName: sm.workerd.displayName,
		Description: sm.workerd.description,
		Arguments:   []string{"-service", "run"},
	}

	if sm.workerd.configPath != "" {
		svcConfig.Arguments = append(svcConfig.Arguments, "-config", sm.workerd.configPath)
	}

	s, err := service.New(sm.workerd, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return s, nil
}

// createLogger creates the service logger
func (sm *ServiceManager) createLogger() (service.Logger, error) {
	errs := make(chan error, 5)
	logger, err := sm.service.Logger(errs)
	if err != nil {
		return nil, fmt.Errorf("failed to create service logger: %w", err)
	}

	// Store error channel for logging goroutine
	sm.workerd.errorChan = errs

	return logger, nil
}

// startErrorLogging starts the error logging goroutine
func (sm *ServiceManager) startErrorLogging() {
	if sm.workerd.errorChan == nil {
		return
	}

	for {
		err := <-sm.workerd.errorChan
		if err != nil {
			log.Print(err)
			if sm.workerd.log != nil {
				sm.workerd.log.Error("Service error", "error", err)
			}
		}
	}
}

// HandleControl handles service control commands
func (sm *ServiceManager) HandleControl(action string) error {
	if sm.service == nil {
		return fmt.Errorf("service not initialized")
	}
	if action == "" {
		return fmt.Errorf("action cannot be empty")
	}

	switch action {
	case "run":
		if err := sm.service.Run(); err != nil {
			return fmt.Errorf("failed to run service: %w", err)
		}
	case "install", "uninstall", "start", "stop", "restart":
		if err := service.Control(sm.service, action); err != nil {
			return fmt.Errorf("service control action '%s' failed: %w (valid actions: %q)",
				action, err, service.ControlAction)
		}
	default:
		return fmt.Errorf("unknown service action '%s' (valid actions: run, %q)",
			action, service.ControlAction)
	}

	return nil
}

// GetService returns the underlying service instance
func (sm *ServiceManager) GetService() service.Service {
	return sm.service
}

// GetLogger returns the service logger
func (sm *ServiceManager) GetLogger() service.Logger {
	return sm.logger
}
