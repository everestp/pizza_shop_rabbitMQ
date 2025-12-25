package main

import (
	"fmt"

	"github.com/everestp/pizza-shop/config"
	"github.com/everestp/pizza-shop/constants"
	"github.com/everestp/pizza-shop/handler"
	"github.com/everestp/pizza-shop/logger"
	"github.com/everestp/pizza-shop/middleware"
	"github.com/everestp/pizza-shop/routes"
	"github.com/everestp/pizza-shop/service"
	"github.com/gin-gonic/gin"
)

func main() {
    // 1. Initialize the Web Framework (Gin)
    app := gin.Default()

    // 2. Middleware Setup
    // Recovery ensures that if one request crashes, the whole server doesn't die.
    app.Use(gin.Recovery())
    // CorsMiddleware allows your frontend (React/Angular) to talk to this backend.
    app.Use(middleware.CorsMiddleware)

    // 3. Health Check (Ping)
    // Used by monitoring tools or just to check if the server is "alive."
    app.GET("/ping", func(ctx *gin.Context) {
        ctx.JSON(200, gin.H{
            "message":    "Pizza Shop is open",
            "statusCode": 200,
        })
    })

    // 4. Service Initialization
    // We create our RabbitMQ tools (Publisher to send, Consumer to listen).
    messagePublisher := service.GetMessagePublisher()
    messageConsumer := service.GetMessageConsumerService()

    // 5. Real-time Logic Setup
    // Start the WebSocket receptionist and the Processor (the brain).
    // Note how we pass the WebSocket 'Connection Map' directly into the processor.
    websocketHandler := handler.GetNewWebSocketHandler()
    messageProcessor := service.GetMessageProcessorService(messagePublisher, websocketHandler.GetConnectionMap())

    // 6. Start the Background Worker
    // We use a 'goroutine' (go func) because consuming messages is a blocking task.
    // It must run in the background while the Gin server handles HTTP requests.
    go func() {
        err := messageConsumer.ConsumeEventAndProcess(constants.KITCHEN_ORDER_QUEUE, messageProcessor)
        if err != nil {
            logger.Log(fmt.Sprintf("CRITICAL: failed to consume events: %v", err))
        }
    }()

    // 7. Route Registration
    // This connects the URL paths (/ws and /orders) to their respective handlers.
    routes.RegisterRoutes(app, messagePublisher, websocketHandler)

    // 8. Launch the Server
    port := config.GetEnvProperty("port")
    logger.Log(fmt.Sprintf("Pizza shop started successfully on port : %s", port))

    // This line blocks the main thread and keeps the app running.
    app.Run(fmt.Sprintf(":%s", port))
}