package config

import (
	"fmt"
	"log"
	"strconv"

	"github.com/everestp/pizza-shop/logger"
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQConection acts as a wrapper for the RabbitMQ physical connection
// and the default queue name used by this specific service.
type RabbitMQConection struct {
	conn  *amqp091.Connection // The underlying TCP connection
	queue string              // The name of the default queue for this app
}

// GetNewRabbitMQConnection initializes a new connection by reading environment variables.
// It uses a 'fail-fast' approach (panics if it can't connect) which is common during app startup.
func GetNewRabbitMQConnection() *RabbitMQConection {
	// 1. Retrieve credentials from environment variables
	host := GetEnvProperty("rabbit_mq_host")
	port := GetEnvProperty("rabbit_mq_port")
	username := GetEnvProperty("rabbit_mq_username")
	password := GetEnvProperty("rabbit_mq_password")
	queue := GetEnvProperty("rabbit_mq_default_queue")

	// 2. Convert port string to integer for formatting
	PORT, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Invalid RabbitMQ port provided: %v", err))
	}

	// 3. Construct the AMQP Connection String (amqp://user:pass@host:port/)
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, PORT)
	
	// 4. Dial opens the TCP connection to the broker
	conn, err := amqp091.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Failed to connect to RabbitMQ: %v", err))
	}

	log.Println("Successfully established RabbitMQ connection")

	return &RabbitMQConection{
		conn:  conn,
		queue: queue,
	}
}

// Connect is a helper method used to re-establish a connection if the original one drops.
func (r *RabbitMQConection) Connect() *amqp091.Connection {
	// Note: In a production app, you might want to DRY (Don't Repeat Yourself) 
	// by moving the URL construction logic to a separate private helper method.
	host := GetEnvProperty("rabbit_mq_host")
	port := GetEnvProperty("rabbit_mq_port")
	username := GetEnvProperty("rabbit_mq_username")
	password := GetEnvProperty("rabbit_mq_password")

	PORT, _ := strconv.Atoi(port)
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, PORT)

	conn, err := amqp091.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("Failed to re-connect to RabbitMQ: %v", err))
	}

	log.Println("RabbitMQ connection restored")
	return conn
}

// DeclareQueue ensures a specific queue exists on the RabbitMQ broker.
// RabbitMQ is idempotent: if the queue already exists with these settings, it does nothing.
func (r *RabbitMQConection) DeclareQueue(queueName string) error {
	// Channels are 'virtual connections' inside a TCP connection. 
	// They are cheap to create; TCP connections are expensive.
	channel, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("error creating channel: %w", err)
	}
	defer channel.Close() // Close the channel as soon as the queue is declared

	_, err = channel.QueueDeclare(
		queueName, // Name of the queue
		true,      // Durable: The queue will survive a broker restart
		false,     // Delete when unused: The queue won't be deleted if consumers disconnect
		false,     // Exclusive: Can be used by other connections
		false,     // No-wait: Do not wait for a server response
		nil,       // Arguments: Additional config (like TTL)
	)
	return err
}

// GetConnection returns the active connection. If nil, it tries to connect.
func (r *RabbitMQConection) GetConnection() *amqp091.Connection {
	if r.conn == nil || r.conn.IsClosed() {
		r.conn = r.Connect()
	}
	return r.conn
}

// GetChannel opens a new channel for performing operations (Publishing/Consuming).
// You should usually open a channel, do your work, and then close it.
func (r *RabbitMQConection) GetChannel() *amqp091.Channel {
	// Ensure connection exists before trying to open a channel
	if r.conn == nil || r.conn.IsClosed() {
		r.conn = r.Connect()
	}

	channel, err := r.conn.Channel()
	if err != nil {
		logger.Log("Failed to open channel, retrying...")
		// Simple retry logic
		channel, err = r.conn.Channel()
		if err != nil {
			logger.Log("Permanent channel failure")
			return nil
		}
	}
	return channel
}

// GetQueue returns the default queue name defined in environment variables.
func (r *RabbitMQConection) GetQueue() string {
	return r.queue
}

// Close gracefully shuts down the RabbitMQ connection. 
// Should be called when the application stops (e.g., using defer in main.go).
func (r *RabbitMQConection) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}