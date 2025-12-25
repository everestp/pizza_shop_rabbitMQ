package service

import (
	"sync"
)

// 1. The Interface (The Logic Contract)
// This interface allows the Consumer Service to hand off messages 
// without knowing the internal details of how they are processed.
type IMessageProcessor interface {
	ProcessMessage(message any) error
}

// 2. The Processor Struct
// This is the "brain" of your service. It holds:
// - A Publisher: In case processing a message requires sending a NEW message (e.g., an "Order Confirmed" event).
// - Connections: A map of active WebSocket clients (keyed by UserID or ConnectionID).
type MessageProcessor struct {
	publisher  IMessagePubliser
	connection *map[string]IWebSocketConnection // Map of active users online
	mutext     sync.RWMutex                     // Protects the map from concurrent read/write crashes
}

// 3. The Factory (Constructor)
// This creates the processor and injects the dependencies it needs to work.
func GetMessageProcessorService(publisher IMessagePubliser, connection *map[string]IWebSocketConnection) *MessageProcessor {
	return &MessageProcessor{
		publisher:  publisher,
		connection: connection,
	}
}

// Note: You will need to implement the ProcessMessage method 
// to satisfy the IMessageProcessor interface.