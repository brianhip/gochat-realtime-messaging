package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Room represents a chat room in the system
// This struct defines how rooms are stored in MongoDB and managed in memory
type Room struct {
	// ID is the unique identifier for each room in MongoDB
	// primitive.ObjectID is MongoDB's native ID type
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	
	// Name is the unique display name for the room
	// Users join rooms by specifying this name
	Name string `bson:"name" json:"name"`
	
	// CreatedAt tracks when the room was created
	// Useful for room management and analytics
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	
	// CreatorID is the user who created this room
	// Links to the User collection via ObjectID
	CreatorID primitive.ObjectID `bson:"creator_id" json:"creator_id"`
}

// NewRoom creates a new Room instance with the current timestamp
// This constructor ensures consistency in room creation across the application
func NewRoom(name string, creatorID primitive.ObjectID) *Room {
	return &Room{
		Name:      name,
		CreatedAt: time.Now(),
		CreatorID: creatorID,
	}
}

// ClientConnection represents an active WebSocket connection in a room
// This is used for in-memory tracking of connected clients, not stored in MongoDB
type ClientConnection struct {
	// Username identifies the connected user
	// Used for message routing and user lists
	Username string
	
	// Connection is the WebSocket connection object
	// Used to send messages directly to this client
	Connection interface{} // We'll use *websocket.Conn when importing websocket package
	
	// JoinedAt tracks when the user joined this room session
	// Used for connection management and debugging
	JoinedAt time.Time
}

// ActiveRoom represents a room with its currently connected clients
// This is used for in-memory room state management, not stored in MongoDB
type ActiveRoom struct {
	// RoomName is the name of the room
	Name string
	
	// Clients is a map of username to client connection
	// Allows quick lookup and message broadcasting to room members
	Clients map[string]*ClientConnection
	
	// CreatedAt tracks when this active room session was created
	// Used for room lifecycle management
	CreatedAt time.Time
}

// NewActiveRoom creates a new ActiveRoom instance for managing connected clients
// This is used when the first user joins a room
func NewActiveRoom(name string) *ActiveRoom {
	return &ActiveRoom{
		Name:      name,
		Clients:   make(map[string]*ClientConnection),
		CreatedAt: time.Now(),
	}
}

// AddClient adds a new client connection to the room
// Returns true if the client was added, false if they were already in the room
func (r *ActiveRoom) AddClient(username string, connection interface{}) bool {
	// Check if user is already in the room
	if _, exists := r.Clients[username]; exists {
		return false
	}
	
	// Add the new client connection
	r.Clients[username] = &ClientConnection{
		Username:   username,
		Connection: connection,
		JoinedAt:   time.Now(),
	}
	return true
}

// RemoveClient removes a client connection from the room
// Returns true if the client was removed, false if they weren't in the room
func (r *ActiveRoom) RemoveClient(username string) bool {
	if _, exists := r.Clients[username]; !exists {
		return false
	}
	
	delete(r.Clients, username)
	return true
}

// GetClientList returns a list of usernames currently in the room
// Used for sending user list updates to clients
func (r *ActiveRoom) GetClientList() []string {
	usernames := make([]string, 0, len(r.Clients))
	for username := range r.Clients {
		usernames = append(usernames, username)
	}
	return usernames
}

// IsEmpty returns true if no clients are connected to the room
// Used to determine when to clean up inactive rooms
func (r *ActiveRoom) IsEmpty() bool {
	return len(r.Clients) == 0
}