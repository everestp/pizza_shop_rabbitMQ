package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/everestp/pizza-shop/config"
    "github.com/everestp/pizza-shop/logger"
    "github.com/rabbitmq/amqp091-go"
)

// 1. The Interface (The "Contract")
// Use this for dependency injection and testing. Any struct that has 
// these two methods "implements" this interface.
type IMessagePubliser interface {
    PublishEvent(queueName string, body any) error
    DeclareQueue(queueName string) error
}

// 2. The Struct
// It holds a reference to the RabbitMQ connection configuration.
type MessagePublisher struct {
    conf *config.RabbitMQConection
}

// DeclareQueue ensures a queue exists before we try to send messages to it.
func (mp *MessagePublisher) DeclareQueue(queueName string) error {
    channel := mp.conf.GetChannel()
    if channel == nil {
        return fmt.Errorf("message channel is nil, please retry")
    }
    // Note: We aren't closing the channel here because GetChannel() 
    // management is handled by the config package.
    _, err := channel.QueueDeclare(
        queueName,
        true,  // Durable
        false, // Auto-delete
        false, // Exclusive
        false, // No-wait
        nil,   // Args
    )
    return err
}

// PublishEvent converts any Go object to JSON and sends it to RabbitMQ.
func (mp *MessagePublisher) PublishEvent(queueName string, body any) error {
    // A. Marshalling: Convert Go Struct -> JSON Bytes
    data, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("failed to marshal body: %w", err)
    }

    // B. Context with Timeout: Ensures the request doesn't hang forever 
    // if the RabbitMQ server is slow or unresponsive.
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    // C. Defaulting: Use the env variable if no queue name is provided.
    if queueName == "" {
        queueName = config.GetEnvProperty("rabbit_mq_default_queue")
    }

    // D. Channel Management
    channel := mp.conf.GetChannel()
    if channel == nil || channel.IsClosed() {
        panic("RabbitMQ channel is unavailable")
    }

    // E. The Actual Publish
    err = channel.PublishWithContext(ctx,
        "",         // Exchange: Empty string means "Direct" to the queue name
        queueName,  // Routing Key: In this case, our queue name
        false,      // Mandatory
        false,      // Immediate
        amqp091.Publishing{
            ContentType:  "application/json",
            Body:         data,
            DeliveryMode: amqp091.Persistent, // Message survives RabbitMQ restart
        },
    )

    if err != nil {
        return err
    }

    logger.Log(fmt.Sprintf("Event published successfully: %v", body))

    // F. Cleanup: Close the channel after the message is sent to free resources.
    channel.Close()
    return nil
}

// GetMessagePublisher is a Factory function. 
// It creates the publisher and starts the RabbitMQ connection.
func GetMessagePublisher() *MessagePublisher {
    rabbitMQConf := config.GetNewRabbitMQConnection()
    return &MessagePublisher{
        conf: rabbitMQConf,
    }
}