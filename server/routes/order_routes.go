package routes

import (
    "github.com/everestp/pizza-shop/handler"
    "github.com/everestp/pizza-shop/service"
    "github.com/gin-gonic/gin"
)

// RegisterOrderRoutes connects the "Orders" URL paths to their logic.
// It takes a RouterGroup (e.g., "/orders") and the RabbitMQ Publisher.
func RegisterOrderRoutes(router *gin.RouterGroup, messagePublisher service.IMessagePubliser) {

    // 1. Initialize the Handler
    // We "inject" the messagePublisher so the handler can send messages to RabbitMQ.
    oh := handler.GetOrderHandler(messagePublisher)

    // 2. Define the Endpoint
    // This creates the path: POST http://localhost:PORT/orders/create
    router.POST(
        "/create",
        oh.CreateOrder, // This function handles the JSON input and RabbitMQ publishing.
    )
}