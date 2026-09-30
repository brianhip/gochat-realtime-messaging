package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"gochat/models"
	"gochat/protocol"
	"gochat/services"
)

// maxIncomingMessageSize is the largest WebSocket frame (in bytes) the server
// will read from a client before closing the connection
const maxIncomingMessageSize = 8192

// WebSocketHandler handles WebSocket connections for real-time chat
// This handler manages client connections, message routing, and room operations
type WebSocketHandler struct {
	chatService    *services.ChatService
	authService    *services.AuthService
	upgrader       websocket.Upgrader
	allowedOrigins map[string]bool
}

// NewWebSocketHandler creates a new WebSocket handler
// allowedOrigins is the exact-match set of origins permitted to connect
func NewWebSocketHandler(chatService *services.ChatService, authService *services.AuthService, allowedOrigins map[string]bool) *WebSocketHandler {
	h := &WebSocketHandler{
		chatService:    chatService,
		authService:    authService,
		allowedOrigins: allowedOrigins,
	}

	h.upgrader = websocket.Upgrader{
		CheckOrigin:     h.checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return h
}

// checkOrigin validates the WebSocket handshake's Origin header against the
// configured allowlist using an exact match (no substring/prefix matching,
// which could be bypassed with a crafted origin).
func (h *WebSocketHandler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}

	if h.allowedOrigins[origin] {
		return true
	}

	log.Printf("[debug] rejected WebSocket connection from disallowed origin: %s", origin)
	return false
}

// HandleWebSocket upgrades HTTP connections to WebSocket and manages client communication
// GET /ws?token=<jwt_token>
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract and validate JWT token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter", http.StatusUnauthorized)
		return
	}

	// Validate the token and get user information
	user, err := h.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// All writes go through the client's single writer goroutine, which
	// also sends pings; closing the client stops it and closes the connection
	client := services.NewClient(user.Username, conn)
	go client.WritePump()
	defer client.Close()

	log.Printf("WebSocket connection established for user: %s", user.Username)

	// Set up connection parameters. The read limit must fit a maximum-length
	// message (1000 bytes of content) after JSON escaping, which can grow
	// each byte to as many as 6 (\u00XX), plus the envelope around it
	conn.SetReadLimit(maxIncomingMessageSize)
	client.PrepareRead()

	// Track current room for cleanup
	var currentRoom string

	// Main message handling loop
	for {
		// Read message from client
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", user.Username, err)
			}
			break
		}

		// Parse WebSocket message
		var wsMsg protocol.WebSocketMessage
		if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
			log.Printf("Invalid message format from user %s: %v", user.Username, err)
			h.sendErrorMessage(client, "INVALID_MESSAGE_FORMAT", "Invalid message format")
			continue
		}

		// Handle message based on type
		switch wsMsg.Type {
		case protocol.MessageTypeJoin:
			// A client is in at most one room; leave the current one first
			if currentRoom != "" {
				h.chatService.UnregisterClient(currentRoom, client)
				currentRoom = ""
			}
			currentRoom = h.handleJoinRoom(client, user, wsMsg.Payload)
		case protocol.MessageTypeLeave:
			if currentRoom != "" {
				h.handleLeaveRoom(client, user, currentRoom)
				currentRoom = ""
			}
		case protocol.MessageTypeMessage:
			if currentRoom != "" {
				h.handleChatMessage(client, user, currentRoom, wsMsg.Payload)
			} else {
				h.sendErrorMessage(client, "NOT_IN_ROOM", "You must join a room before sending messages")
			}
		case protocol.MessageTypePing:
			h.handlePing(client, wsMsg.Payload)
		default:
			h.sendErrorMessage(client, "UNKNOWN_MESSAGE_TYPE", "Unknown message type")
		}
	}

	// Clean up when connection closes
	if currentRoom != "" {
		h.chatService.UnregisterClient(currentRoom, client)
		log.Printf("User %s disconnected from room %s", user.Username, currentRoom)
	}

	log.Printf("WebSocket connection closed for user: %s", user.Username)
}

// handleJoinRoom processes a room join request
func (h *WebSocketHandler) handleJoinRoom(client *services.Client, user *models.User, payload json.RawMessage) string {
	var joinPayload protocol.JoinRoomPayload
	if err := json.Unmarshal(payload, &joinPayload); err != nil {
		h.sendErrorMessage(client, "INVALID_JOIN_PAYLOAD", "Invalid join room payload")
		return ""
	}

	// Validate that the username matches the authenticated user
	if joinPayload.Username != user.Username {
		h.sendErrorMessage(client, "USERNAME_MISMATCH", "Username in payload doesn't match authenticated user")
		return ""
	}

	// Create room in database if it doesn't exist
	_, err := h.chatService.GetRoom(joinPayload.RoomName)
	if err != nil {
		// Room doesn't exist, create it
		_, err = h.chatService.CreateRoom(joinPayload.RoomName, user.ID)
		if err != nil {
			log.Printf("Error creating room %s: %v", joinPayload.RoomName, err)
			h.sendErrorMessage(client, "ROOM_CREATION_FAILED", "Failed to create room")
			return ""
		}
	}

	// Register client with chat service
	h.chatService.RegisterClient(joinPayload.RoomName, client)

	// Send room history to new client
	history, err := h.chatService.GetRoomHistory(joinPayload.RoomName, 50)
	if err != nil {
		log.Printf("Error fetching room history: %v", err)
	} else {
		h.sendRoomHistory(client, joinPayload.RoomName, history)
	}

	// Send join confirmation
	joinedPayload := &protocol.JoinedRoomPayload{
		RoomName: joinPayload.RoomName,
		Username: user.Username,
		Success:  true,
		Message:  "Successfully joined room",
	}

	msg, err := protocol.NewWebSocketMessage(protocol.MessageTypeJoinedRoom, joinedPayload)
	if err == nil {
		client.Send(msg)
	}

	return joinPayload.RoomName
}

// handleLeaveRoom processes a room leave request
func (h *WebSocketHandler) handleLeaveRoom(client *services.Client, user *models.User, roomName string) {
	// Unregister client from chat service
	h.chatService.UnregisterClient(roomName, client)

	// Send leave confirmation
	leftPayload := &protocol.LeftRoomPayload{
		RoomName: roomName,
		Username: user.Username,
		Success:  true,
	}

	msg, err := protocol.NewWebSocketMessage(protocol.MessageTypeLeftRoom, leftPayload)
	if err == nil {
		client.Send(msg)
	}

	log.Printf("User %s left room %s", user.Username, roomName)
}

// handleChatMessage processes a chat message from a client
func (h *WebSocketHandler) handleChatMessage(client *services.Client, user *models.User, roomName string, payload json.RawMessage) {
	var msgPayload protocol.MessagePayload
	if err := json.Unmarshal(payload, &msgPayload); err != nil {
		h.sendErrorMessage(client, "INVALID_MESSAGE_PAYLOAD", "Invalid message payload")
		return
	}

	// Validate that the username and room match
	if msgPayload.Username != user.Username {
		h.sendErrorMessage(client, "USERNAME_MISMATCH", "Username in payload doesn't match authenticated user")
		return
	}
	if msgPayload.RoomName != roomName {
		h.sendErrorMessage(client, "ROOM_MISMATCH", "Room in payload doesn't match current room")
		return
	}

	// Validate message content
	if len(msgPayload.Content) == 0 {
		h.sendErrorMessage(client, "EMPTY_MESSAGE", "Message content cannot be empty")
		return
	}
	if len(msgPayload.Content) > 1000 {
		h.sendErrorMessage(client, "MESSAGE_TOO_LONG", "Message content is too long")
		return
	}

	// Create message object
	message := models.NewMessage(roomName, user.Username, msgPayload.Content)

	// Save message to database
	err := h.chatService.SaveMessage(message)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		h.sendErrorMessage(client, "MESSAGE_SAVE_FAILED", "Failed to save message")
		return
	}

	// Create broadcast payload
	newMsgPayload := &protocol.NewMessagePayload{
		Message:  message,
		RoomName: roomName,
	}

	// Broadcast message to all clients in the room (including sender)
	broadcastMsg, err := protocol.NewWebSocketMessage(protocol.MessageTypeNewMessage, newMsgPayload)
	if err != nil {
		log.Printf("Error creating broadcast message: %v", err)
		return
	}

	// Broadcast to all clients including the sender (don't exclude sender)
	h.chatService.BroadcastMessage(roomName, broadcastMsg, "")

	log.Printf("Message from %s in room %s: %s", user.Username, roomName, msgPayload.Content)
}

// handlePing processes a ping message for connection health
func (h *WebSocketHandler) handlePing(client *services.Client, payload json.RawMessage) {
	var pingPayload protocol.PingPayload
	if err := json.Unmarshal(payload, &pingPayload); err != nil {
		h.sendErrorMessage(client, "INVALID_PING_PAYLOAD", "Invalid ping payload")
		return
	}

	// Create pong response
	pongPayload := &protocol.PongPayload{
		ClientTimestamp: pingPayload.ClientTimestamp,
		ServerTimestamp: time.Now(),
	}

	msg, err := protocol.NewWebSocketMessage(protocol.MessageTypePong, pongPayload)
	if err != nil {
		log.Printf("Error creating pong message: %v", err)
		return
	}

	client.Send(msg)
}

// sendErrorMessage sends an error message to the client
func (h *WebSocketHandler) sendErrorMessage(client *services.Client, code, message string) {
	errorPayload := &protocol.ErrorPayload{
		Code:    code,
		Message: message,
	}

	msg, err := protocol.NewWebSocketMessage(protocol.MessageTypeError, errorPayload)
	if err != nil {
		log.Printf("Error creating error message: %v", err)
		return
	}

	client.Send(msg)
}

// sendRoomHistory sends historical messages to a client
func (h *WebSocketHandler) sendRoomHistory(client *services.Client, roomName string, messages []*models.Message) {
	historyPayload := &protocol.RoomHistoryPayload{
		RoomName: roomName,
		Messages: messages,
		Count:    len(messages),
		HasMore:  false, // For now, we don't implement pagination
	}

	msg, err := protocol.NewWebSocketMessage(protocol.MessageTypeRoomHistory, historyPayload)
	if err != nil {
		log.Printf("Error creating room history message: %v", err)
		return
	}

	client.Send(msg)
}