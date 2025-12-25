package handler

import (
	"github.com/everestp/pizza-shop/constants"
	"github.com/everestp/pizza-shop/service"
	"github.com/gin-gonic/gin"
)

// OrderHandler is the "Postman" of your API. 
// It receives HTTP requests and passes them to the RabbitMQ system.
type OrderHandler struct {
	messagePublisher service.IMessagePubliser // Dependency: Interface to talk to RabbitMQ
}

// CreateOrder handles the POST request when a user places a pizza order.
func (oh *OrderHandler) CreateOrder(ctx *gin.Context) {
	var payload map[string]any

	// 1. Bind JSON: Read the data sent by the user (e.g., pizza type, quantity).
	// If the JSON is broken, we return a 400 Bad Request immediately.
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(400, gin.H{
			"message":    "Invalid order data provided",
			"statusCode": 400,
		})
		return // Stop processing if input is bad
	}

	// 2. Initial State: Every new order starts with the status "ORDERED".
	// We add this to the payload so the Consumer knows how to process it later.
	payload["order_status"] = constants.ORDER_ORDERED

	// 3. Hand-off: Send the order to RabbitMQ. 
	// This makes our API fast because we don't wait for the chef to cook; 
	// we just put the order on the "To-Do List" (Queue).
	err := oh.messagePublisher.PublishEvent(constants.KITCHEN_ORDER_QUEUE, payload)
	if err != nil {
		ctx.JSON(500, gin.H{
			"message": "Failed to send order to kitchen",
			"error":   err.Error(),
		})
		return
	}

	// 4. Response: Tell the user "We got your order!" 
	// They can now wait for the WebSocket update.
	ctx.JSON(200, gin.H{
		"data":       payload,
		"statusCode": 200,
		"message":    "Order accepted successfully! The kitchen is being notified.",
	})
}

// GetOrderHandler is the Constructor. 
// Note: I fixed the parameter type to service.IMessagePubliser to match the struct.
func GetOrderHandler(messagePublisher service.IMessagePubliser) *OrderHandler {
	return &OrderHandler{
		messagePublisher: messagePublisher,
	}
}