package workerd

import (
	"fmt"
	"log/slog"
)

// ConfigMerger handles merging configuration from multiple sources
type ConfigMerger struct {
	defaultConfig *workerConfig
	fileConfig    *workerConfig
	optionsConfig *WorkerdOptions
}

// WorkerdOptions represents configuration options from functional options
type WorkerdOptions struct {
	Name        string
	DisplayName string
	Description string
	Concurrency int
	Logger      *slog.Logger
	ConfigPath  string
	ServiceFlag string
}

// NewConfigMerger creates a new configuration merger
func NewConfigMerger() *ConfigMerger {
	return &ConfigMerger{}
}

// WithDefaults sets the default configuration
func (cm *ConfigMerger) WithDefaults(config *workerConfig) *ConfigMerger {
	cm.defaultConfig = config
	return cm
}

// WithFileConfig sets the configuration loaded from files
func (cm *ConfigMerger) WithFileConfig(config *workerConfig) *ConfigMerger {
	cm.fileConfig = config
	return cm
}

// WithOptions sets the configuration from functional options
func (cm *ConfigMerger) WithOptions(options *WorkerdOptions) *ConfigMerger {
	cm.optionsConfig = options
	return cm
}

// Merge merges configurations with priority: options > file > defaults
func (cm *ConfigMerger) Merge() (*MergedConfig, error) {
	if cm.defaultConfig == nil {
		return nil, fmt.Errorf("default configuration is required")
	}

	merged := &MergedConfig{
		Name:        cm.getStringValue("name"),
		DisplayName: cm.getStringValue("displayName"),
		Description: cm.getStringValue("description"),
		Concurrency: cm.getIntValue("concurrency"),
		ConfigPath:  cm.getStringValue("configPath"),
		ServiceFlag: cm.getStringValue("serviceFlag"),
		Logger:      cm.getLoggerValue(),
		Config:      cm.fileConfig,
	}

	// Use default config if file config is not available
	if merged.Config == nil {
		merged.Config = cm.defaultConfig
	}

	return merged, nil
}

// MergedConfig represents the final merged configuration
type MergedConfig struct {
	Name        string
	DisplayName string
	Description string
	Concurrency int
	ConfigPath  string
	ServiceFlag string
	Logger      *slog.Logger
	Config      *workerConfig
}

// getStringValue gets string value with priority: options > file > defaults
func (cm *ConfigMerger) getStringValue(field string) string {
	// Check options first
	if cm.optionsConfig != nil {
		switch field {
		case "name":
			if cm.optionsConfig.Name != "" {
				return cm.optionsConfig.Name
			}
		case "displayName":
			if cm.optionsConfig.DisplayName != "" {
				return cm.optionsConfig.DisplayName
			}
		case "description":
			if cm.optionsConfig.Description != "" {
				return cm.optionsConfig.Description
			}
		case "configPath":
			if cm.optionsConfig.ConfigPath != "" {
				return cm.optionsConfig.ConfigPath
			}
		case "serviceFlag":
			if cm.optionsConfig.ServiceFlag != "" {
				return cm.optionsConfig.ServiceFlag
			}
		}
	}

	// Check file config second
	if cm.fileConfig != nil {
		switch field {
		case "name":
			if cm.fileConfig.Name != "" {
				return cm.fileConfig.Name
			}
		case "displayName":
			if cm.fileConfig.DisplayName != "" {
				return cm.fileConfig.DisplayName
			}
		case "description":
			if cm.fileConfig.Description != "" {
				return cm.fileConfig.Description
			}
		}
	}

	// Fall back to defaults
	if cm.defaultConfig != nil {
		switch field {
		case "name":
			return cm.defaultConfig.Name
		case "displayName":
			return cm.defaultConfig.DisplayName
		case "description":
			return cm.defaultConfig.Description
		}
	}

	return ""
}

// getIntValue gets int value with priority: options > file > defaults
func (cm *ConfigMerger) getIntValue(field string) int {
	// Check options first
	if cm.optionsConfig != nil {
		switch field {
		case "concurrency":
			if cm.optionsConfig.Concurrency > 0 {
				return cm.optionsConfig.Concurrency
			}
		}
	}

	// Check file config second
	if cm.fileConfig != nil {
		switch field {
		case "concurrency":
			if cm.fileConfig.Concurrency > 0 {
				return cm.fileConfig.Concurrency
			}
		}
	}

	// Fall back to defaults
	if cm.defaultConfig != nil {
		switch field {
		case "concurrency":
			return cm.defaultConfig.Concurrency
		}
	}

	return 0
}

// getLoggerValue gets logger value with priority: options > defaults
func (cm *ConfigMerger) getLoggerValue() *slog.Logger {
	if cm.optionsConfig != nil && cm.optionsConfig.Logger != nil {
		return cm.optionsConfig.Logger
	}
	return nil
}