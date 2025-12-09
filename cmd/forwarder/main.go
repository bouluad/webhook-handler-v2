package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webhook-handler-project/internal/config"
	"webhook-handler-project/internal/queue"
	"webhook-handler-project/internal/forward"
)

// ForwarderConfig extends the base config with target tool details
type ForwarderConfig struct {
	*config.Config
}

func loadForwarderConfig() *ForwarderConfig {
	baseCfg := config.LoadBaseConfig() 
	
    // Ensure forwarder-specific envs are present
    baseCfg.TargetToolURL = getEnvStrict("TARGET_TOOL_URL")
    baseCfg.TargetToolAuthToken = getEnv("TARGET_TOOL_AUTH_TOKEN", "") // Auth token can be optional
	
	return &ForwarderConfig{Config: baseCfg}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvStrict(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("Environment variable %s is required and not set.", key)
	return "" 
}

func main() {
	// 1. Load Configuration
	cfg := loadForwarderConfig()

	// 2. Initialize Service Bus Consumer
	consumer, err := queue.NewServiceBusConsumer(cfg.ServiceBusConnectionString, cfg.ServiceBusQueueName)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Service Bus consumer: %v", err)
	}
	defer consumer.Close(context.Background())

	// 3. Initialize Target Tool Client
	toolClient := forward.NewClient(cfg.TargetToolURL, cfg.TargetToolAuthToken)

	// 4. Start the message processing loop
    ctx, cancel := context.WithCancel(context.Background())
    
    log.Printf("Starting ASB consumer for queue: %s. Target URL: %s", cfg.ServiceBusQueueName, cfg.TargetToolURL)
	go startConsumerLoop(ctx, consumer, toolClient)

	// 5. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop 

	log.Println("Forwarder Shutting down...")
    cancel() 
    time.Sleep(5 * time.Second)

	log.Println("Forwarder gracefully stopped.")
}

func startConsumerLoop(ctx context.Context, consumer *queue.ServiceBusConsumer, toolClient *forward.Client) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // Receive messages with a timeout
            messages, err := consumer.Receive(ctx, 1, 10 * time.Second) // Receive 1 message at a time
            if err != nil {
                log.Printf("Consumer Error receiving messages: %v", err)
                time.Sleep(2 * time.Second) 
                continue
            }
            
            for _, msg := range messages {
                rawPayload := msg.Body
                
                // Process and forward
                if err := toolClient.ForwardPayload(rawPayload); err != nil {
                    log.Printf("FORWARDING FAILED for message ID %s: %v. Abandoning message.", msg.MessageID, err)
                    consumer.Abandon(ctx, msg)
                } else {
                    log.Printf("FORWARDING SUCCEEDED for message ID %s. Completing message.", msg.MessageID)
                    consumer.Complete(ctx, msg)
                }
            }
        }
    }
}
