package routes

import (
    "github.com/everestp/pizza-shop/handler"
    "github.com/gin-gonic/gin"
)

// RegisterWebSocketRoutes sets up the live communication path.
// It takes a RouterGroup (like "/ws") and the Handler that knows how to manage connections.
func RegisterWebSocketRoutes(router *gin.RouterGroup, websocketHandler handler.IWebSocketHandler) {
    
    // This defines the specific endpoint for WebSockets.
    // If the group is "/ws", the full URL will be "ws://yourdomain.com/ws/"
    router.GET(
        "/", 
        websocketHandler.HandleConnection, // The function that upgrades HTTP to WebSocket
    )
}

/* FUTURE REFERENCE:
If you want to add more WebSocket-related paths later, 
you would add them here. For example:

router.GET("/admin", websocketHandler.HandleAdminConnection)
*/