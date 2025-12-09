package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
    
	"webhook-handler-project/internal/config"
	"webhook-handler-project/internal/forward"
	"webhook-handler-project/internal/queue"
    "webhook-handler-project/internal/forwarder"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadBaseConfig() 
    
    // Enforce Forwarder's strict requirements
    cfg.ServiceBusConnectionString = config.GetEnvStrict("AZURE_SERVICE_BUS_CONN_STRING")
    cfg.ServiceBusQueueName = config.GetEnvStrict("AZURE_SERVICE_BUS_QUEUE_NAME")
    cfg.TargetToolURL = config.GetEnvStrict("TARGET_TOOL_URL")
    cfg.TargetToolAuthToken = config.GetEnv("TARGET_TOOL_AUTH_TOKEN", "") // Auth token is optional

	// 2. Initialize Service Bus Consumer
	consumer, err := queue.NewServiceBusConsumer(cfg.ServiceBusConnectionString, cfg.ServiceBusQueueName)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Service Bus consumer: %v", err)
	}
	defer consumer.Close(context.Background())

	// 3. Initialize Target Tool Client
	toolClient := forward.NewClient(cfg.TargetToolURL, cfg.TargetToolAuthToken)

	// 4. Initialize and Run the Consumer
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    forwarder := forwarder.NewConsumer(cfg, consumer, toolClient)
    
    // Start the consuming loop
    go forwarder.Run(ctx) 

	// 5. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop 

	log.Println("Forwarder Shutting down...")
    cancel() // Triggers context cancellation
    
    // Give a small moment for the loop to check ctx.Done() and exit
    time.Sleep(1 * time.Second) 

	log.Println("Forwarder application exited.")
}
