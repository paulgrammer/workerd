package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/paulgrammer/workerd"
)

func main() {
	serviceFlag := flag.String("service", "run", "Control the system service (install, uninstall, start, stop, restart, run)")
	configPath := flag.String("config", "", "Path to either a file or directory to load configuration from")
	name := flag.String("name", "workerd", "Service name")
	displayName := flag.String("display-name", "Workerd Service", "Service display name")
	description := flag.String("description", "Background worker service for job processing", "Service description")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent workers")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	// Build workerd options
	opts := []workerd.Option{
		workerd.WithName(*name),
		workerd.WithDisplayName(*displayName),
		workerd.WithDescription(*description),
		workerd.WithConcurrency(*concurrency),
	}

	if *configPath != "" {
		opts = append(opts, workerd.WithConfigPath(*configPath))
	}

	if *serviceFlag != "" {
		opts = append(opts, workerd.WithServiceFlag(*serviceFlag))
	}

	// Create workerd instance
	w, err := workerd.NewWorkerd(opts...)
	if err != nil {
		log.Printf("Failed to create workerd: %v\n", err)
		os.Exit(1)
	}

	// Register some example task handlers
	registerTaskHandlers(w)

	// Run the workerd
	if err := w.Run(); err != nil {
		log.Printf("Failed to run workerd: %v\n", err)
		os.Exit(1)
	}
}

// registerTaskHandlers registers example task handlers
func registerTaskHandlers(w *workerd.Workerd) {
	// Register a simple task handler
	w.HandleFunc("example:send_email", handleSendEmail)
	w.HandleFunc("example:process_image", handleProcessImage)
	w.HandleFunc("example:cleanup", handleCleanup)

	// You can register more handlers as needed
	w.Handle("example:complex_task", asynq.HandlerFunc(handleComplexTask))
}

// Example task handlers
func handleSendEmail(ctx context.Context, t *asynq.Task) error {
	// Extract task payload
	payload := string(t.Payload())

	// Log the task processing
	log.Printf("Processing email task: %s", payload)

	// Simulate email sending
	time.Sleep(2 * time.Second)

	log.Printf("Email sent successfully: %s", payload)
	return nil
}

func handleProcessImage(ctx context.Context, t *asynq.Task) error {
	payload := string(t.Payload())

	log.Printf("Processing image task: %s", payload)

	// Simulate image processing
	time.Sleep(5 * time.Second)

	log.Printf("Image processed successfully: %s", payload)
	return nil
}

func handleCleanup(ctx context.Context, t *asynq.Task) error {
	payload := string(t.Payload())

	log.Printf("Processing cleanup task: %s", payload)

	// Simulate cleanup work
	time.Sleep(1 * time.Second)

	log.Printf("Cleanup completed: %s", payload)
	return nil
}

func handleComplexTask(ctx context.Context, t *asynq.Task) error {
	payload := string(t.Payload())

	log.Printf("Processing complex task: %s", payload)

	// Simulate complex processing with multiple steps
	for i := 1; i <= 3; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Printf("Complex task step %d/3: %s", i, payload)
			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("Complex task completed: %s", payload)
	return nil
}
