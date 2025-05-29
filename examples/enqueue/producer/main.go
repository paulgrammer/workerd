// enqueue.go - Task enqueuer example
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// TaskPayload represents a generic task payload
type TaskPayload struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
	Created time.Time              `json:"created"`
}

func main() {
	redisAddr := flag.String("redis", "localhost:6379", "Redis address")
	taskType := flag.String("task", "example:send_email", "Task type to enqueue")
	taskData := flag.String("data", `{"email":"test@example.com","subject":"Hello World"}`, "Task data as JSON")
	count := flag.Int("count", 1, "Number of tasks to enqueue")
	delay := flag.Duration("delay", 0, "Delay before processing (e.g., 30s, 5m)")

	flag.Parse()

	// Create Redis client for enqueueing tasks
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: *redisAddr,
	})
	defer client.Close()

	// Parse task data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(*taskData), &data); err != nil {
		log.Fatalf("Failed to parse task data: %v", err)
	}

	// Enqueue tasks
	for i := 0; i < *count; i++ {
		payload := TaskPayload{
			ID:      fmt.Sprintf("task-%d-%d", time.Now().Unix(), i),
			Type:    *taskType,
			Data:    data,
			Created: time.Now(),
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal payload for task %d: %v", i, err)
			continue
		}

		task := asynq.NewTask(*taskType, payloadBytes)

		// Set task options
		opts := []asynq.Option{
			asynq.MaxRetry(3),
			asynq.Queue("default"),
		}

		if *delay > 0 {
			opts = append(opts, asynq.ProcessIn(*delay))
		}

		// Enqueue the task
		info, err := client.Enqueue(task, opts...)
		if err != nil {
			log.Printf("Failed to enqueue task %d: %v", i, err)
			continue
		}

		fmt.Printf("Enqueued task: ID=%s, Queue=%s, Type=%s\n",
			info.ID, info.Queue, info.Type)
	}

	fmt.Printf("Successfully enqueued %d tasks\n", *count)
}
