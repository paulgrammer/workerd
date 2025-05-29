package main

import (
	"log"
	"os"

	"github.com/hibiken/asynq"
	"github.com/paulgrammer/workerd"
)

func main() {
	mux := asynq.NewServeMux()
	wf := workerd.NewWorkerdWithFlags(workerd.WithServeMux(mux))

	// Run the workerd
	if err := wf.Run(); err != nil {
		log.Printf("Failed to run workerd: %v\n", err)
		os.Exit(1)
	}
}
