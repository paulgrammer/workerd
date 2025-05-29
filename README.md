# Workerd - Background Worker Service

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Workerd is a robust, production-ready background worker service built on top of [Asynq](https://github.com/hibiken/asynq). It provides both standalone and system service deployment options with comprehensive configuration management and graceful shutdown capabilities.

## Features

- 🚀 **Dual Mode Operation**: Run as standalone application or system service
- ⚙️ **Flexible Configuration**: JSON-based configuration with sensible defaults
- 🔧 **Functional Options**: Clean, extensible API design
- 📊 **Structured Logging**: Configurable log levels with structured output
- 🛡️ **Graceful Shutdown**: Proper signal handling and task completion
- 🔄 **Redis Backend**: Reliable job queue using Redis
- 📈 **Concurrent Processing**: Configurable worker concurrency
- 🎯 **Task Routing**: Type-based task handler registration
- 🏗️ **Production Ready**: Built for enterprise deployment

## Installation

```bash
go get github.com/paulgrammer/workerd
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "github.com/hibiken/asynq"
    "github.com/paulgrammer/workerd"
)

func main() {
    // Create workerd instance
    w := workerd.NewWorkerd(
        workerd.WithName("my-worker"),
        workerd.WithConcurrency(20),
    )

    // Register task handlers
    w.HandleFunc("email:send", handleSendEmail)
    w.HandleFunc("image:process", handleImageProcess)

    // Run the workerd
    if err := w.Run(); err != nil {
        log.Fatal(err)
    }
}

func handleSendEmail(ctx context.Context, t *asynq.Task) error {
    // Process email sending task
    log.Printf("Sending email: %s", string(t.Payload()))
    return nil
}

func handleImageProcess(ctx context.Context, t *asynq.Task) error {
    // Process image manipulation task
    log.Printf("Processing image: %s", string(t.Payload()))
    return nil
}
```

### Using CLI

```bash
# Run in standalone mode
./workerd

# Run with custom configuration
./workerd -config config.json

# Install as system service
sudo ./workerd -service install

# Start the service
sudo ./workerd -service start

# Stop the service
sudo ./workerd -service stop
```

## Configuration

### Configuration File (config.json)

```json
{
  "name": "workerd",
  "display_name": "Workerd Service",
  "description": "Background worker service for job processing",
  "concurrency": 15,
  "log_level": "info",
  "redis": {
    "addr": "localhost:6379",
    "password": "",
    "db": 0,
    "pool_size": 20
  }
}
```

### Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | "workerd" | Service name |
| `display_name` | string | "Workerd Service" | Human-readable service name |
| `description` | string | "Background worker service" | Service description |
| `concurrency` | int | 10 | Number of concurrent workers |
| `log_level` | string | "info" | Log level (debug, info, warn, error) |
| `redis.addr` | string | "localhost:6379" | Redis server address |
| `redis.password` | string | "" | Redis password |
| `redis.db` | int | 0 | Redis database number |
| `redis.pool_size` | int | 10 | Redis connection pool size |

## API Reference

### Constructor

```go
func NewWorkerd(opts ...Option) *Workerd
```

Creates a new workerd instance with the specified options.

### Functional Options

```go
func WithName(name string) Option
func WithDisplayName(displayName string) Option
func WithDescription(desc string) Option
func WithConcurrency(n int) Option
func WithConfigPath(path string) Option
func WithServiceFlag(serviceFlag string) Option
func WithLogger(logger *slog.Logger) Option
func WithServeMux(mux *asynq.ServeMux) Option
```

### Methods

#### Task Registration

```go
// Register a handler function
func (w *Workerd) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error)

// Register a handler
func (w *Workerd) Handle(pattern string, handler asynq.Handler)
```

## Command Line Interface

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `-service` | string | Service control action (install, uninstall, start, stop, restart, run) |
| `-config` | string | Path to configuration file or directory |
| `-name` | string | Service name |
| `-display-name` | string | Service display name |
| `-description` | string | Service description |
| `-concurrency` | int | Number of concurrent workers |
| `-help` | bool | Print usage information |

### Service Commands

```bash
# Install service
sudo ./workerd -service install

# Uninstall service
sudo ./workerd -service uninstall

# Start service
sudo ./workerd -service start

# Stop service
sudo ./workerd -service stop

# Restart service
sudo ./workerd -service restart
```

## Task Enqueueing

Create tasks using the asynq client:

```go
package main

import (
    "encoding/json"
    "github.com/hibiken/asynq"
)

func main() {
    client := asynq.NewClient(asynq.RedisClientOpt{
        Addr: "localhost:6379",
    })
    defer client.Close()

    // Create task payload
    payload := map[string]interface{}{
        "email": "user@example.com",
        "subject": "Welcome!",
    }

    payloadBytes, _ := json.Marshal(payload)
    task := asynq.NewTask("email:send", payloadBytes)

    // Enqueue task
    info, err := client.Enqueue(task)
    if err != nil {
        panic(err)
    }

    println("Enqueued task:", info.ID)
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Dependencies

- [hibiken/asynq](https://github.com/hibiken/asynq) - Simple, reliable, and efficient distributed task queue
- [kardianos/service](https://github.com/kardianos/service) - Run go programs as a service on major platforms

## Support

- Create an [issue](https://github.com/koodeyo/veloxpack/issues) for bug reports
- Start a [discussion](https://github.com/koodeyo/veloxpack/discussions) for questions
- Check the [documentation](https://pkg.go.dev/github.com/paulgrammer/workerd) for API reference
