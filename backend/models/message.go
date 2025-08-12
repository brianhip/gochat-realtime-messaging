package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message represents a chat message in the system
// This struct defines how messages are stored in MongoDB and transmitted via WebSocket
type Message struct {
	// ID is the unique identifier for each message in MongoDB
	// primitive.ObjectID is MongoDB's native ID type
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	
	// RoomName identifies which room this message belongs to
	// Messages are scoped to specific rooms for proper message routing
	RoomName string `bson:"room_name" json:"room_name"`
	
	// SenderUsername identifies who sent the message
	// This is stored as username rather than user ID for easier display
	SenderUsername string `bson:"sender_username" json:"sender_username"`
	
	// Content is the actual text content of the message
	// This is what users see in the chat interface
	Content string `bson:"message_content" json:"content"`
	
	// Timestamp tracks when the message was sent
	// Used for message ordering and display purposes
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// NewMessage creates a new Message instance with the current timestamp
// This constructor ensures consistency in message creation across the application
func NewMessage(roomName, senderUsername, content string) *Message {
	return &Message{
		RoomName:       roomName,
		SenderUsername: senderUsername,
		Content:        content,
		Timestamp:      time.Now(),
	}
}