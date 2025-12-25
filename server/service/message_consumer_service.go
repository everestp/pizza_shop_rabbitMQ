package service

import (
	"fmt"

	"github.com/everestp/pizza-shop/config"
	"github.com/everestp/pizza-shop/logger"
	"github.com/rabbitmq/amqp091-go"
)

// 1. The Interface
// This defines what a Consumer must do. Note that it takes an 'IMessageProcessor',
// which is another interface that tells this service HOW to handle the data.
type IMessageConsumerService interface {
	DeclareQueue(queueName string) error
	ConsumeEventAndProcess(queueName string, processor IMessageProcessor) error
}

type MessageConsumerService struct {
	conf *config.RabbitMQConection
}

// DeclareQueue ensures the queue exists before we start listening.
// It's a safety step to avoid errors if the consumer starts before the publisher.
func (mcs *MessageConsumerService) DeclareQueue(queueName string) error {
	channel := mcs.conf.GetChannel()
	if channel == nil {
		return fmt.Errorf("message channel is nil, please retry")
	}

	_, err := channel.QueueDeclare(
		queueName,
		true,  // Durable: Queue survives RabbitMQ restart
		false, // Auto-delete: No
		false, // Exclusive: No
		false, // No-wait: No
		nil,
	)
	return err
}

// ConsumeEventAndProcess starts a long-running loop that waits for messages.
func (mcs *MessageConsumerService) ConsumeEventAndProcess(queueName string, processor IMessageProcessor) error {
	channel := mcs.conf.GetChannel()
	if channel == nil {
		return fmt.Errorf("message channel is nil, please retry")
	}

	logger.Log("Starting message consumption...")

	// 2. Consume returns a Go Channel (msgs) where messages will arrive.
	msgs, err := channel.Consume(
		queueName, // The queue to listen to
		"",        // Consumer tag (unique ID for this consumer instance)
		false,     // Auto-Ack: Set to false so we manually acknowledge successful processing
		false,     // Exclusive
		false,     // No-local
		false,     // No-wait
		nil,       // Args
	)
	if err != nil {
		return fmt.Errorf("failed to consume message: %w", err)
	}

	// 3. The Worker Loop
	// We run this in a Goroutine so it doesn't block the rest of the app.
	go func() {
		for msg := range msgs {
			// 4. Parallel Processing
			// We start a NEW Goroutine for every single message.
			// This allows the app to process multiple pizzas at the same time!
			go func(d amqp091.Delivery) {
				err := processor.ProcessMessage(d)
				if err != nil {
					logger.Log(fmt.Sprintf("Message processing failed: %v", err))
					// In the future, you might want to msg.Nack() here to retry
				}
			}(msg)
		}
	}()

	// 5. Block Forever
	// This prevents the function from returning, keeping the consumer alive.
	select {}
}

// GetMessageConsumerService is the factory function to initialize the service.
func GetMessageConsumerService() *MessageConsumerService {
	rabbitMQConf := config.GetNewRabbitMQConnection()
	return &MessageConsumerService{
		conf: rabbitMQConf,
	}
}