package protocol

import (
	"encoding/json"
	"time"

	"gochat/models"
)

// MessageType defines the different types of WebSocket messages
// This enum helps clients and server understand how to process each message
type MessageType string

const (
	// Client-to-Server message types
	MessageTypeJoin    MessageType = "JOIN"    // User wants to join a room
	MessageTypeLeave   MessageType = "LEAVE"   // User wants to leave a room
	MessageTypeMessage MessageType = "MESSAGE" // User is sending a chat message
	MessageTypePing    MessageType = "PING"    // Client ping for connection health
	
	// Server-to-Client message types
	MessageTypeJoinedRoom    MessageType = "JOINED_ROOM"    // Confirmation of room join
	MessageTypeLeftRoom      MessageType = "LEFT_ROOM"      // Confirmation of room leave
	MessageTypeNewMessage    MessageType = "NEW_MESSAGE"    // New chat message broadcast
	MessageTypeUserJoined    MessageType = "USER_JOINED"    // Another user joined the room
	MessageTypeUserLeft      MessageType = "USER_LEFT"      // Another user left the room
	MessageTypeUserList      MessageType = "USER_LIST"      // Current users in room
	MessageTypeError         MessageType = "ERROR"          // Error message
	MessageTypePong          MessageType = "PONG"           // Server pong response
	MessageTypeRoomHistory   MessageType = "ROOM_HISTORY"   // Historical messages for room
)

// WebSocketMessage is the base structure for all WebSocket communications
// All messages between client and server use this format for consistency
type WebSocketMessage struct {
	// Type identifies what kind of message this is
	// Used by both client and server to determine how to handle the message
	Type MessageType `json:"type"`
	
	// Payload contains the actual message data
	// The structure varies based on the message type
	Payload json.RawMessage `json:"payload"`
	
	// Timestamp tracks when the message was created
	// Useful for debugging and message ordering
	Timestamp time.Time `json:"timestamp"`
}

// NewWebSocketMessage creates a new WebSocket message with the current timestamp
// This constructor ensures all messages have consistent formatting
func NewWebSocketMessage(msgType MessageType, payload interface{}) (*WebSocketMessage, error) {
	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	return &WebSocketMessage{
		Type:      msgType,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}, nil
}

// JoinRoomPayload is sent by clients when they want to join a room
type JoinRoomPayload struct {
	// RoomName is the name of the room the user wants to join
	RoomName string `json:"room_name"`
	
	// Username identifies the user joining the room
	// This should match their authenticated username
	Username string `json:"username"`
}

// LeaveRoomPayload is sent by clients when they want to leave a room
type LeaveRoomPayload struct {
	// RoomName is the name of the room the user wants to leave
	RoomName string `json:"room_name"`
	
	// Username identifies the user leaving the room
	Username string `json:"username"`
}

// MessagePayload is sent by clients when they send a chat message
type MessagePayload struct {
	// RoomName is the room where the message should be sent
	RoomName string `json:"room_name"`
	
	// Username identifies who is sending the message
	Username string `json:"username"`
	
	// Content is the actual message text
	Content string `json:"content"`
}

// NewMessagePayload is sent by server to broadcast new messages to room members
type NewMessagePayload struct {
	// Message contains the full message details
	Message *models.Message `json:"message"`
	
	// RoomName is included for client-side filtering (redundant but helpful)
	RoomName string `json:"room_name"`
}

// UserJoinedPayload is sent by server when a new user joins a room
type UserJoinedPayload struct {
	// Username is the user who joined
	Username string `json:"username"`
	
	// RoomName is the room they joined
	RoomName string `json:"room_name"`
	
	// JoinedAt tracks when they joined
	JoinedAt time.Time `json:"joined_at"`
}

// UserLeftPayload is sent by server when a user leaves a room
type UserLeftPayload struct {
	// Username is the user who left
	Username string `json:"username"`
	
	// RoomName is the room they left
	RoomName string `json:"room_name"`
	
	// LeftAt tracks when they left
	LeftAt time.Time `json:"left_at"`
}

// UserListPayload is sent by server to provide the current user list for a room
type UserListPayload struct {
	// RoomName is the room this user list applies to
	RoomName string `json:"room_name"`
	
	// Users is the list of currently connected usernames in the room
	Users []string `json:"users"`
	
	// Count is the total number of users (redundant but convenient for clients)
	Count int `json:"count"`
}

// JoinedRoomPayload is sent by server to confirm a user successfully joined a room
type JoinedRoomPayload struct {
	// RoomName is the room the user joined
	RoomName string `json:"room_name"`
	
	// Username is the user who joined (confirmation)
	Username string `json:"username"`
	
	// Success indicates if the join was successful
	Success bool `json:"success"`
	
	// Message provides additional context or error details
	Message string `json:"message,omitempty"`
}

// LeftRoomPayload is sent by server to confirm a user left a room
type LeftRoomPayload struct {
	// RoomName is the room the user left
	RoomName string `json:"room_name"`
	
	// Username is the user who left (confirmation)
	Username string `json:"username"`
	
	// Success indicates if the leave was successful
	Success bool `json:"success"`
}

// ErrorPayload is sent by server when an error occurs
type ErrorPayload struct {
	// Code is a machine-readable error identifier
	Code string `json:"code"`
	
	// Message is a human-readable error description
	Message string `json:"message"`
	
	// Details provides additional context about the error
	Details map[string]interface{} `json:"details,omitempty"`
}

// RoomHistoryPayload is sent by server to provide historical messages for a room
type RoomHistoryPayload struct {
	// RoomName is the room these messages belong to
	RoomName string `json:"room_name"`
	
	// Messages is the list of historical messages
	Messages []*models.Message `json:"messages"`
	
	// Count is the total number of messages returned
	Count int `json:"count"`
	
	// HasMore indicates if there are more messages available
	HasMore bool `json:"has_more"`
}

// PingPayload is sent by clients for connection health checks
type PingPayload struct {
	// Timestamp from client for round-trip time calculation
	ClientTimestamp time.Time `json:"client_timestamp"`
}

// PongPayload is sent by server in response to ping
type PongPayload struct {
	// ClientTimestamp echoed back from the ping
	ClientTimestamp time.Time `json:"client_timestamp"`
	
	// ServerTimestamp when the pong was sent
	ServerTimestamp time.Time `json:"server_timestamp"`
}