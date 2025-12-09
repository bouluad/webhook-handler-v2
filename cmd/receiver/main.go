package main

import (
	"context"
	"log"

	"webhook-handler-project/internal/config"
	"webhook-handler-project/internal/queue"
	"webhook-handler-project/internal/receiver"
)

func main() {
	// 1. Load Configuration
    baseCfg := config.LoadBaseConfig()
    
    // Enforce Receiver's strict requirements
    baseCfg.ServiceBusConnectionString = config.GetEnvStrict("AZURE_SERVICE_BUS_CONN_STRING")
    baseCfg.ServiceBusQueueName = config.GetEnvStrict("AZURE_SERVICE_BUS_QUEUE_NAME")
    baseCfg.GitHubSecret = config.GetEnvStrict("GITHUB_WEBHOOK_SECRET")


	// 2. Initialize Azure Service Bus Publisher
	publisher, err := queue.NewServiceBusPublisher(baseCfg.ServiceBusConnectionString, baseCfg.ServiceBusQueueName)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Service Bus publisher: %v", err)
	}
	defer publisher.Close(context.Background()) 

	// 3. Initialize and Run the Server
    ctx := context.Background()
	server := receiver.NewServer(baseCfg, publisher)
    
	if err := server.Run(ctx); err != nil {
        log.Fatalf("FATAL: Receiver server failed to run: %v", err)
    }

    log.Println("Receiver application exited.")
}
