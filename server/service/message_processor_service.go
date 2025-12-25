package service

import (
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/everestp/pizza-shop/constants"
    "github.com/everestp/pizza-shop/logger"
    "github.com/everestp/pizza-shop/utils"
    "github.com/rabbitmq/amqp091-go"
)

// IMessageProcessor is the "Contract." 
// Any struct that wants to process messages must have the ProcessMessage method.
type IMessageProcessor interface {
    ProcessMessage(message interface{}) error
}

// MessageProcessor is the "Brain" of the operation.
// It connects RabbitMQ (the messenger) to WebSockets (the live update for users).
type MessageProcessor struct {
    publisher  IMessagePubliser                 // To send events back to RabbitMQ
    connection *map[string]IWebSocketConnection // List of users currently online via WebSockets
    mutex      sync.RWMutex                     // The "Lock" to prevent crashes when multiple people use the map
}

// ProcessMessage is the entry point for every message coming from the queue.
func (mp *MessageProcessor) ProcessMessage(message interface{}) error {
    // 1. Convert the generic message into a RabbitMQ 'Delivery' object
    msg := message.(amqp091.Delivery)
    
    var event map[string]interface{}
    var err error

    // 2. Parse JSON: Convert the message bytes into a Go map (key-value pairs)
    if err = json.Unmarshal(msg.Body, &event); err != nil {
        logger.Log(fmt.Sprintf("JSON Error: Cannot read message body: %v", err))
        // Nack(false, true) means: "I failed, put this back in the queue to try again."
        msg.Nack(false, true) 
        return err
    }

    logger.Log(fmt.Sprintf("Step 1: Received message for processing: %v", event))

    // 3. State Machine: Decide what to do based on the "order_status"
    if val, ok := event["order_status"]; ok {
        switch val {
        case constants.ORDER_ORDERED:
            // Flow: Customer ordered -> Send to Kitchen
            err = mp.handleOrderOrdered(event)
            
        case constants.ORDER_PREPARING:
            // Flow: Kitchen is cooking -> Simulate time and move to Prepared
            err = mp.handleOrderPreparing(event)
            
        case constants.ORDER_PREPARED:
            // Flow: Pizza is ready -> Notify the user via WebSocket
            err = mp.handleOrderPrepared(event)
            
        default:
            logger.Log("Unknown Status: Skipping processing.")
        }

        // 4. If any of the logic above fails, Nack the message so we don't lose it
        if err != nil {
            logger.Log(fmt.Sprintf("Processing Error: %v", err))
            msg.Nack(false, true)
            return err
        }
    }

    // 5. Success! Tell RabbitMQ to delete the message from the queue
    msg.Ack(false)
    return nil
}

// handleOrderOrdered: Moves the order from "Customer" to "Kitchen"
func (mp *MessageProcessor) handleOrderOrdered(event map[string]interface{}) error {
    logger.Log("Action: Accepting order and sending to Kitchen queue.")
    
    // Set the new status
    event["order_status"] = constants.ORDER_PREPARING
    
    // Publish the updated event back to RabbitMQ
    err := mp.publisher.PublishEvent(constants.KITCHEN_ORDER_QUEUE, event)
    if err != nil {
        mp.sendErrorToUser(err, event)
    }
    return err
}

// handleOrderPreparing: Represents the "Chef" actually making the pizza
func (mp *MessageProcessor) handleOrderPreparing(event map[string]interface{}) error {
    logger.Log(fmt.Sprintf("Action: Chef started preparing order #%v", event["order_no"]))
    
    // 1. Simulate the "Cooking Time" (1 to 6 seconds)
    time.Sleep(utils.GenerateRandomDuration(1, 6))
    
    // 2. Set new status
    event["order_status"] = constants.ORDER_PREPARED
    
    // 3. Publish the update back to RabbitMQ
    err := mp.publisher.PublishEvent(constants.KITCHEN_ORDER_QUEUE, event)
    if err != nil {
        mp.sendErrorToUser(err, event)
    }
    return err
}

// handleOrderPrepared: Final step. Sends a "Your Pizza is Ready" alert to the UI
func (mp *MessageProcessor) handleOrderPrepared(event map[string]interface{}) error {
    logger.Log(fmt.Sprintf("Action: Order #%v is ready! Notifying customer.", event["order_no"]))
    
    event["order_status"] = constants.ORDER_DELIVERED
    
    // Prepare the JSON data for the WebSocket
    message := map[string]interface{}{
        "message": constants.ORDER_PREPARED_SUCCESSFULLY,
        "order":   event,
    }
    
    return mp.broadcastToWebSocket(message)
}

// broadcastToWebSocket: A helper to send messages to the Frontend safely
func (mp *MessageProcessor) broadcastToWebSocket(data interface{}) error {
    bytes, _ := json.Marshal(data)

    if mp.connection != nil {
        // LOCKING: Because many messages might finish at once, we use a Mutex.
        // This stops the app from crashing due to "concurrent map access."
        mp.mutex.Lock()
        defer mp.mutex.Unlock()

        // In this demo, we use the key "pizza" to find the user.
        socket := (*mp.connection)["pizza"]
        if socket != nil {
            return socket.SendMessage(bytes)
        }
    }
    return nil
}

// sendErrorToUser: Notifies the frontend if something goes wrong in the backend
func (mp *MessageProcessor) sendErrorToUser(err error, event map[string]interface{}) {
    logger.Log(fmt.Sprintf("Error Trace: %v | Data: %v", err, event))
    
    errMsg := map[string]string{
        "message": constants.ORDER_CANCELLED,
        "error":   err.Error(),
    }
    mp.broadcastToWebSocket(errMsg)
}

// GetMessageProcessorService: The "Constructor" to initialize this service
func GetMessageProcessorService(publisher IMessagePubliser, connection *map[string]IWebSocketConnection) *MessageProcessor {
    return &MessageProcessor{
        publisher:  publisher,
        connection: connection,
    }
}