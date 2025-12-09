package forwarder

import (
	"context"
	"fmt"
	"log"
	"time"

	"webhook-handler-project/internal/config"
	"webhook-handler-project/internal/forward"
	"webhook-handler-project/internal/queue"
)

// Consumer manages the long-running Service Bus message processing.
type Consumer struct {
	cfg       *config.Config
	consumer  *queue.ServiceBusConsumer
	toolClient *forward.Client
}

// NewConsumer creates a new instance of the forwarder consumer.
func NewConsumer(cfg *config.Config, consumer *queue.ServiceBusConsumer, toolClient *forward.Client) *Consumer {
	return &Consumer{
		cfg:       cfg,
		consumer:  consumer,
		toolClient: toolClient,
	}
}

// Run starts the message processing loop until the context is cancelled.
func (c *Consumer) Run(ctx context.Context) {
    log.Printf("Starting ASB consumer for queue: %s. Target URL: %s", c.cfg.ServiceBusQueueName, c.cfg.TargetToolURL)
    
    for {
        select {
        case <-ctx.Done():
            log.Println("Consumer loop received shutdown signal.")
            return
        default:
            // Attempt to receive a batch of 1 message
            messages, err := c.consumer.Receive(ctx, 1, 10 * time.Second)
            if err != nil {
                log.Printf("Consumer Error receiving messages: %v", err)
                time.Sleep(2 * time.Second) // Wait before retrying after an error
                continue
            }
            
            for _, msg := range messages {
                c.processMessage(ctx, msg)
            }
        }
    }
}

func (c *Consumer) processMessage(ctx context.Context, msg *queue.ReceivedMessage) {
    rawPayload := msg.Body
    
    // Process and forward
    if err := c.toolClient.ForwardPayload(rawPayload); err != nil {
        // Forwarding failed. Abandon the message so it can be redelivered.
        log.Printf("FORWARDING FAILED for message ID %s: %v. Abandoning message.", msg.MessageID, err)
        if abandonErr := c.consumer.Abandon(ctx, msg); abandonErr != nil {
             log.Printf("CRITICAL: Failed to abandon message ID %s: %v", msg.MessageID, abandonErr)
        }
    } else {
        // Forwarding succeeded. Complete the ASB message.
        log.Printf("FORWARDING SUCCEEDED for message ID %s. Completing message.", msg.MessageID)
        if completeErr := c.consumer.Complete(ctx, msg); completeErr != nil {
             log.Printf("CRITICAL: Failed to complete message ID %s: %v", msg.MessageID, completeErr)
        }
    }
}
