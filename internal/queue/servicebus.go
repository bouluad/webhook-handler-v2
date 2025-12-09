package queue

import (
	"context"
	"log"
	"time"

	azservicebus "github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// ServiceBusPublisher manages the connection and publishing.
type ServiceBusPublisher struct {
	sender *azservicebus.Sender
}

// ServiceBusConsumer manages the connection and consumption.
type ServiceBusConsumer struct {
	receiver *azservicebus.Receiver
}

// NewServiceBusPublisher initializes the Service Bus client and sender.
func NewServiceBusPublisher(connString, queueName string) (*ServiceBusPublisher, error) {
	client, err := azservicebus.NewClientFromConnectionString(connString, nil)
	if err != nil {
		return nil, err
	}
	sender, err := client.NewSender(queueName, nil)
	if err != nil {
		return nil, err
	}
	return &ServiceBusPublisher{sender: sender}, nil
}

// Publish sends the raw payload body to the queue.
func (p *ServiceBusPublisher) Publish(ctx context.Context, payloadBody []byte) error {
	message := azservicebus.NewMessage(payloadBody)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.sender.SendMessage(ctx, message, nil); err != nil {
		return err
	}
	return nil
}

// Close gracefully closes the sender connection.
func (p *ServiceBusPublisher) Close(ctx context.Context) {
	p.sender.Close(ctx)
}

// NewServiceBusConsumer initializes the Service Bus client and receiver.
func NewServiceBusConsumer(connString, queueName string) (*ServiceBusConsumer, error) {
	client, err := azservicebus.NewClientFromConnectionString(connString, nil)
	if err != nil {
		return nil, err
	}
	receiver, err := client.NewReceiver(queueName, nil)
	if err != nil {
		return nil, err
	}
	return &ServiceBusConsumer{receiver: receiver}, nil
}

// Receive attempts to receive messages.
func (c *ServiceBusConsumer) Receive(ctx context.Context, maxMessages int, maxWaitTime time.Duration) ([]*azservicebus.ReceivedMessage, error) {
    // Note: A real-world app might use the ProcessMessages loop for automatic handling.
	return c.receiver.ReceiveMessages(ctx, maxMessages, &azservicebus.ReceiveMessagesOptions{MaxWaitTime: &maxWaitTime})
}

// Complete marks a message as successfully processed.
func (c *ServiceBusConsumer) Complete(ctx context.Context, msg *azservicebus.ReceivedMessage) error {
	return c.receiver.CompleteMessage(ctx, msg, nil)
}

// Abandon marks a message for redelivery.
func (c *ServiceBusConsumer) Abandon(ctx context.Context, msg *azservicebus.ReceivedMessage) error {
	return c.receiver.AbandonMessage(ctx, msg, nil)
}

// Close gracefully closes the receiver connection.
func (c *ServiceBusConsumer) Close(ctx context.Context) {
	c.receiver.Close(ctx)
}
