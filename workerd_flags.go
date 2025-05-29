package workerd

import (
	"flag"
	"fmt"
)

type cliFlags struct {
	service     string
	configPath  string
	name        string
	displayName string
	description string
	concurrency int
}

func parseFlags() *cliFlags {
	flags := &cliFlags{}
	flag.StringVar(&flags.service, "service", "run", "Control the system service (install, uninstall, start, stop, restart, run)")
	flag.StringVar(&flags.configPath, "config", "", "Path to either a file or directory to load configuration from")
	flag.StringVar(&flags.name, "name", "", "Service name")
	flag.StringVar(&flags.displayName, "display-name", "", "Service display name")
	flag.StringVar(&flags.description, "description", "", "Service description")
	flag.IntVar(&flags.concurrency, "concurrency", 1, "Number of concurrent workers")
	flag.Parse()
	return flags
}

// WorkerdWithFlags handles command-line interface concerns for workerd
type WorkerdWithFlags struct {
	opts []Option
}

// NewWorkerdWithFlags creates a new WorkerdWithFlags instance with flag definitions
func NewWorkerdWithFlags(opts ...Option) *WorkerdWithFlags {
	return &WorkerdWithFlags{opts}
}

// BuildOptions builds workerd options from WorkerdWithFlags flags
func (c *WorkerdWithFlags) buildOptions(flags *cliFlags) []Option {
	opts := c.opts

	if flags.configPath != "" {
		opts = append(opts, WithConfigPath(flags.configPath))
	}
	if flags.service != "" {
		opts = append(opts, WithServiceFlag(flags.service))
	}
	if flags.name != "" {
		opts = append(opts, WithName(flags.name))
	}
	if flags.displayName != "" {
		opts = append(opts, WithDisplayName(flags.displayName))
	}
	if flags.description != "" {
		opts = append(opts, WithDescription(flags.description))
	}
	if flags.concurrency > 0 {
		opts = append(opts, WithConcurrency(flags.concurrency))
	}

	return opts
}

// Run runs the workerd with command line interface with mux
func (c *WorkerdWithFlags) Run() error {
	flags := parseFlags()
	// Build options
	opts := c.buildOptions(flags)

	// Create and run workerd
	workerd, err := NewWorkerd(opts...)
	if err != nil {
		return fmt.Errorf("failed to create workerd: %v", err)
	}

	if err := workerd.Run(); err != nil {
		return fmt.Errorf("failed to run workerd: %v", err)
	}

	return nil
}
