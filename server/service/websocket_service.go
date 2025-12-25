package service

import (
    "sync"

    "github.com/gorilla/websocket"
)

// 1. The Interface (Abstraction)
// This allows you to swap the 'gorilla/websocket' library for another 
// one in the future without changing your business logic.
type IWebSocketConnection interface {
    SendMessage(message []byte) error
    ReceivedMessage() ([]byte, error)
    Close() error
}

// 2. The Wrapper Struct
// We wrap the raw *websocket.Conn to add extra safety (Mutex).
type WebSocketConnection struct {
    conn  *websocket.Conn
    mutex sync.Mutex // Vital for thread-safety
}

// SendMessage sends data from the SERVER to the CLIENT (Browser).
func (ws *WebSocketConnection) SendMessage(message []byte) error {
    // WebSockets in Go are not safe for concurrent writes.
    // The Mutex ensures that if two processes try to send a message 
    // at the exact same time, they wait in line instead of crashing.
    ws.mutex.Lock()
    defer ws.mutex.Unlock()
    
    return ws.conn.WriteMessage(websocket.TextMessage, message)
}

// ReceivedMessage listens for data coming from the CLIENT to the SERVER.
func (ws *WebSocketConnection) ReceivedMessage() ([]byte, error) {
    // Note: Usually, Read and Write happen in different loops, 
    // but the Mutex here prevents internal state corruption.
    ws.mutex.Lock()
    defer ws.mutex.Unlock()
    
    _, msg, err := ws.conn.ReadMessage()
    return msg, err
}

// Close cleanly terminates the connection.
func (ws *WebSocketConnection) Close() error {
    return ws.conn.Close()
}

// NewWebSocketConnection is the constructor.
func NewWebSocketConnection(conn *websocket.Conn) *WebSocketConnection {
    return &WebSocketConnection{
        conn: conn,
    }
}