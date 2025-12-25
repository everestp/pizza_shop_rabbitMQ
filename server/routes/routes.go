package routes

import (
    "github.com/everestp/pizza-shop/handler"
    "github.com/everestp/pizza-shop/service"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes is the "Master Switchboard". 
// It connects the Gin engine to all the different parts of your application.
func RegisterRoutes(r *gin.Engine, messagePublisher service.IMessagePubliser, websocketHandler handler.IWebSocketHandler) {

    // 1. Create a Base Group
    // All routes in the app start from here.
    router := r.Group("/")

    // 2. WebSocket Routes Group
    // Path: http://localhost:PORT/ws/
    // This group handles the "Live" connection between the user and the server.
    wsr := router.Group("/ws")
    {
        // We pass the websocketHandler here so it can manage the online users map.
        RegisterWebSocketRoutes(wsr, websocketHandler)
    }

    // 3. Order Routes Group
    // Path: http://localhost:PORT/orders/
    // This group handles the "Transactional" part (creating new pizza orders).
    or := router.Group("/orders")
    {
        // We pass the messagePublisher so that new orders can be pushed into RabbitMQ.
        RegisterOrderRoutes(or, messagePublisher)
    }

}