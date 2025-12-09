package queue

import (
	"context"
	"log"
	"time"

	azservicebus "github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// Aliases for clarity
type ServiceBusPublisher struct {
	sender *azservicebus.Sender
}
type ServiceBusConsumer struct {
	receiver *azservicebus.Receiver
}
type ReceivedMessage = azservicebus.ReceivedMessage

// --- PUBLISHER ---

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

func (p *ServiceBusPublisher) Publish(ctx context.Context, payloadBody []byte) error {
	message := azservicebus.NewMessage(payloadBody)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.sender.SendMessage(ctx, message, nil); err != nil {
		return err
	}
	return nil
}

func (p *ServiceBusPublisher) Close(ctx context.Context) {
	p.sender.Close(ctx)
}

// --- CONSUMER ---

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

func (c *ServiceBusConsumer) Receive(ctx context.Context, maxMessages int, maxWaitTime time.Duration) ([]*ReceivedMessage, error) {
	return c.receiver.ReceiveMessages(ctx, maxMessages, &azservicebus.ReceiveMessagesOptions{MaxWaitTime: &maxWaitTime})
}

func (c *ServiceBusConsumer) Complete(ctx context.Context, msg *ReceivedMessage) error {
	return c.receiver.CompleteMessage(ctx, msg, nil)
}

func (c *ServiceBusConsumer) Abandon(ctx context.Context, msg *ReceivedMessage) error {
	return c.receiver.AbandonMessage(ctx, msg, nil)
}

func (c *ServiceBusConsumer) Close(ctx context.Context) {
	c.receiver.Close(ctx)
}
