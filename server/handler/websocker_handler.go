package handler

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/everestp/pizza-shop/logger"
	"github.com/everestp/pizza-shop/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// IWebSocketHandler is the contract for managing WebSocket traffic.
type IWebSocketHandler interface {
	HandleConnection(ctx *gin.Context)
	GetConnectionMap() *map[string]service.IWebSocketConnection
}

// WebSocketHandler manages the lifecycle of browser-to-server connections.
type WebSocketHandler struct {
	upgrader   websocket.Upgrader                        // Tools to turn HTTP into WebSocket
	connection *map[string]service.IWebSocketConnection // The "Address Book" of online users
	mutex      sync.Mutex                                // The "Lock" to prevent map crashes
}

// HandleConnection is the main endpoint (e.g., /ws). It runs every time a user connects.
func (h *WebSocketHandler) HandleConnection(ctx *gin.Context) {
	// 1. Upgrade: Change the connection from HTTP to WebSocket protocol.
	conn, err := h.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Log(fmt.Sprintf("CRITICAL: Failed to upgrade connection: %v", err))
		return
	}
	// 2. Ensure the connection closes when this function finishes.
	defer conn.Close()

	// 3. Welcome Message: Send an initial message to the client.
	conn.WriteMessage(websocket.TextMessage, []byte("Connection Established: Started taking order updates..."))

	// 4. Wrap & Store: Wrap the raw connection in our Service and add it to our Map.
	connection := service.NewWebSocketConnection(conn)
	
	// We use "pizza" as a hardcoded ID for now. 
	// In a real app, you'd get the UserID from a Token or URL.
	h.addConnection("pizza", connection)

	// 5. Keep Alive: This loop keeps the connection open.
	// Without this loop, the function would end and the connection would close.
	for {
		// We read messages here if we expect the client to talk back.
		_, _, err := conn.ReadMessage()
		if err != nil {
			logger.Log("Client disconnected or error occurred")
			break // Exit the loop to trigger the defer conn.Close()
		}
	}
}

// addConnection safely puts a new user into our "Address Book" (Map).
func (h *WebSocketHandler) addConnection(clientId string, connection service.IWebSocketConnection) {
	// Lock the map before writing so two users connecting at once don't crash the server.
	h.mutex.Lock()
	defer h.mutex.Unlock()

	(*h.connection)[clientId] = connection
	logger.Log(fmt.Sprintf("User [%s] added to active connections", clientId))
}

// GetConnectionMap returns the pointer to our address book.
// This is used by the MessageProcessor to find users to send alerts to.
func (h *WebSocketHandler) GetConnectionMap() *map[string]service.IWebSocketConnection {
	return h.connection
}

// GetNewWebSocketHandler is the Constructor to set up the receptionist service.
func GetNewWebSocketHandler() *WebSocketHandler {
	// Initialize the map (make sure it's not nil!)
	connection := make(map[string]service.IWebSocketConnection)
	
	return &WebSocketHandler{
		connection: &connection,
		upgrader: websocket.Upgrader{
			// CheckOrigin: true allows any website to connect to your socket.
			// In production, you would restrict this to your specific domain.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}