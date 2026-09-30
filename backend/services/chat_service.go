package services

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gochat/db"
	"gochat/models"
	"gochat/protocol"
)

// ChatService handles real-time chat operations and room management
// This service manages active rooms, message broadcasting, and message persistence
type ChatService struct {
	// MongoDB collections for persistent storage
	messageCollection *mongo.Collection
	roomCollection    *mongo.Collection
	
	// In-memory state for active rooms and connections
	activeRooms map[string]*models.ActiveRoom
	roomsMutex  sync.RWMutex // Protects activeRooms map from concurrent access
	
	// Channels for message broadcasting
	broadcast chan *BroadcastMessage
	register  chan *ClientRegistration
	unregister chan *ClientUnregistration
	
	// Service lifecycle management
	done chan bool
}

// BroadcastMessage represents a message that needs to be sent to room members
type BroadcastMessage struct {
	RoomName string
	Message  *protocol.WebSocketMessage
	SenderUsername string // Used to avoid echoing back to sender
}

// ClientRegistration represents a new client joining a room
type ClientRegistration struct {
	RoomName string
	Username string
	Client   *Client
}

// ClientUnregistration represents a client leaving a room
type ClientUnregistration struct {
	RoomName string
	Username string
	Client   *Client // Only this connection is removed, not a newer one for the same user
}

// NewChatService creates a new chat service instance
// It initializes MongoDB collections and starts the message broadcasting goroutine
func NewChatService() *ChatService {
	service := &ChatService{
		messageCollection: db.Database.Collection("messages"),
		roomCollection:    db.Database.Collection("rooms"),
		activeRooms:       make(map[string]*models.ActiveRoom),
		broadcast:         make(chan *BroadcastMessage, 256),
		register:          make(chan *ClientRegistration, 256),
		unregister:        make(chan *ClientUnregistration, 256),
		done:              make(chan bool),
	}

	// Start the message broadcasting goroutine
	go service.run()

	return service
}

// run handles the main message broadcasting loop
// This goroutine processes client registrations, unregistrations, and message broadcasts
func (s *ChatService) run() {
	for {
		select {
		case registration := <-s.register:
			s.handleClientRegistration(registration)
		case unregistration := <-s.unregister:
			s.handleClientUnregistration(unregistration)
		case broadcastMsg := <-s.broadcast:
			s.handleBroadcast(broadcastMsg)
		case <-s.done:
			return
		}
	}
}

// handleClientRegistration processes a new client joining a room
func (s *ChatService) handleClientRegistration(reg *ClientRegistration) {
	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	// Create room if it doesn't exist
	if _, exists := s.activeRooms[reg.RoomName]; !exists {
		s.activeRooms[reg.RoomName] = models.NewActiveRoom(reg.RoomName)
		log.Printf("Created new active room: %s", reg.RoomName)
	}

	room := s.activeRooms[reg.RoomName]
	
	// Add client to room
	if room.AddClient(reg.Username, reg.Client) {
		log.Printf("User %s joined room %s", reg.Username, reg.RoomName)
		
		// Send the updated user list to everyone in the room, including the new client
		s.sendUserListToRoom(room)
		
		// Notify other clients about new user
		userJoinedPayload := &protocol.UserJoinedPayload{
			Username: reg.Username,
			RoomName: reg.RoomName,
			JoinedAt: time.Now(),
		}
		
		msg, err := protocol.NewWebSocketMessage(protocol.MessageTypeUserJoined, userJoinedPayload)
		if err == nil {
			s.broadcastToRoom(reg.RoomName, msg, reg.Username)
		}
	}
}

// handleClientUnregistration processes a client leaving a room
func (s *ChatService) handleClientUnregistration(unreg *ClientUnregistration) {
	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	room, exists := s.activeRooms[unreg.RoomName]
	if !exists {
		return
	}

	// Ignore stale unregistrations from a connection the user has since replaced
	if client, ok := room.Clients[unreg.Username]; !ok || client.Connection != unreg.Client {
		return
	}

	// Remove client from room
	if room.RemoveClient(unreg.Username) {
		log.Printf("User %s left room %s", unreg.Username, unreg.RoomName)
		
		// Notify other clients about user leaving
		userLeftPayload := &protocol.UserLeftPayload{
			Username: unreg.Username,
			RoomName: unreg.RoomName,
			LeftAt:   time.Now(),
		}
		
		msg, err := protocol.NewWebSocketMessage(protocol.MessageTypeUserLeft, userLeftPayload)
		if err == nil {
			s.broadcastToRoom(unreg.RoomName, msg, unreg.Username)
		}
		
		// Clean up empty room, or send the remaining members the updated user list
		if room.IsEmpty() {
			delete(s.activeRooms, unreg.RoomName)
			log.Printf("Cleaned up empty room: %s", unreg.RoomName)
		} else {
			s.sendUserListToRoom(room)
		}
	}
}

// handleBroadcast sends a message to all clients in a room
func (s *ChatService) handleBroadcast(broadcastMsg *BroadcastMessage) {
	s.roomsMutex.RLock()
	room, exists := s.activeRooms[broadcastMsg.RoomName]
	s.roomsMutex.RUnlock()

	if !exists {
		return
	}

	// Send message to all clients in the room except the sender
	for username, client := range room.Clients {
		if username == broadcastMsg.SenderUsername {
			continue // Don't echo back to sender
		}
		
		c, ok := client.Connection.(*Client)
		if !ok {
			continue
		}
		
		// Queue without blocking; a slow or broken client is closed by its own
		// write pump, and its read loop then unregisters it
		c.Send(broadcastMsg.Message)
	}
}

// RegisterClient adds a new client to a room
func (s *ChatService) RegisterClient(roomName string, client *Client) {
	s.register <- &ClientRegistration{
		RoomName: roomName,
		Username: client.Username,
		Client:   client,
	}
}

// UnregisterClient removes a client from a room
func (s *ChatService) UnregisterClient(roomName string, client *Client) {
	s.unregister <- &ClientUnregistration{
		RoomName: roomName,
		Username: client.Username,
		Client:   client,
	}
}

// BroadcastMessage sends a message to all clients in a room
func (s *ChatService) BroadcastMessage(roomName string, message *protocol.WebSocketMessage, senderUsername string) {
	s.broadcast <- &BroadcastMessage{
		RoomName:       roomName,
		Message:        message,
		SenderUsername: senderUsername,
	}
}

// SaveMessage persists a message to the database
func (s *ChatService) SaveMessage(message *models.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.messageCollection.InsertOne(ctx, message)
	return err
}

// GetRoomHistory retrieves recent messages for a room
func (s *ChatService) GetRoomHistory(roomName string, limit int) ([]*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find messages for the room, sorted by timestamp descending
	opts := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetLimit(int64(limit))

	cursor, err := s.messageCollection.Find(ctx, bson.M{"room_name": roomName}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	err = cursor.All(ctx, &messages)
	if err != nil {
		return nil, err
	}

	// Reverse the slice to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CreateRoom creates a new room in the database
func (s *ChatService) CreateRoom(roomName string, creatorID primitive.ObjectID) (*models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if room already exists
	var existingRoom models.Room
	err := s.roomCollection.FindOne(ctx, bson.M{"name": roomName}).Decode(&existingRoom)
	if err == nil {
		return nil, errors.New("room already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Create new room
	room := models.NewRoom(roomName, creatorID)

	// Insert room into database
	result, err := s.roomCollection.InsertOne(ctx, room)
	if err != nil {
		return nil, err
	}

	// Set the generated ID
	room.ID = result.InsertedID.(primitive.ObjectID)

	return room, nil
}

// GetRoom retrieves room information from database
func (s *ChatService) GetRoom(roomName string) (*models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var room models.Room
	err := s.roomCollection.FindOne(ctx, bson.M{"name": roomName}).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("room not found")
		}
		return nil, err
	}

	return &room, nil
}

// GetActiveRoomUsers returns the list of users currently in a room
func (s *ChatService) GetActiveRoomUsers(roomName string) []string {
	s.roomsMutex.RLock()
	defer s.roomsMutex.RUnlock()

	room, exists := s.activeRooms[roomName]
	if !exists {
		return []string{}
	}

	return room.GetClientList()
}

// sendUserListToRoom sends the room's current user list to every client in it
// Must be called from the run goroutine with roomsMutex held. It queues on each
// client directly instead of going through the broadcast channel, which only
// this goroutine drains and so could deadlock if full
func (s *ChatService) sendUserListToRoom(room *models.ActiveRoom) {
	users := room.GetClientList()
	userListPayload := &protocol.UserListPayload{
		RoomName: room.Name,
		Users:    users,
		Count:    len(users),
	}

	msg, err := protocol.NewWebSocketMessage(protocol.MessageTypeUserList, userListPayload)
	if err != nil {
		log.Printf("Error creating user list message: %v", err)
		return
	}

	for _, client := range room.Clients {
		if c, ok := client.Connection.(*Client); ok {
			c.Send(msg)
		}
	}
}

// broadcastToRoom sends a message to all clients in a room
func (s *ChatService) broadcastToRoom(roomName string, message *protocol.WebSocketMessage, excludeUsername string) {
	s.broadcast <- &BroadcastMessage{
		RoomName:       roomName,
		Message:        message,
		SenderUsername: excludeUsername,
	}
}

// Stop gracefully shuts down the chat service
func (s *ChatService) Stop() {
	close(s.done)
}