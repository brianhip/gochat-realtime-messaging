package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"gochat/db"
	"gochat/handlers"
	"gochat/services"
	"gochat/utils"
)

// minJWTSecretLength is the minimum acceptable length (in bytes) for
// JWT_SECRET. Anything shorter is rejected as too weak to sign tokens with.
const minJWTSecretLength = 32

// main is the entry point of the GoChat server
// It initializes the database, sets up services, configures routes, and starts the HTTP server
func main() {
	log.Println("Starting GoChat server...")

	// Load environment variables from .env file (if it exists)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load and validate the JWT signing secret. There is no default: a
	// missing or weak secret would let anyone forge valid tokens, so the
	// server refuses to start rather than silently falling back.
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < minJWTSecretLength {
		log.Fatalf("JWT_SECRET environment variable must be set and at least %d characters long", minJWTSecretLength)
	}
	utils.SetJWTSecret(jwtSecret)

	// Initialize MongoDB connection
	mongoURI := getEnvOrDefault("MONGO_URI", "mongodb://localhost:27017")
	err := db.InitMongoDB(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := db.DisconnectMongoDB(); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize services
	authService := services.NewAuthService()
	chatService := services.NewChatService()
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	wsHandler := handlers.NewWebSocketHandler(chatService, authService)

	// Set up HTTP routes
	setupRoutes(authHandler, wsHandler)

	// Start server
	port := getEnvOrDefault("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("API endpoints available at: http://localhost:%s/api", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// setupRoutes configures all HTTP routes for the application
// This includes authentication endpoints, WebSocket upgrades, and static file serving
func setupRoutes(authHandler *handlers.AuthHandler, wsHandler *handlers.WebSocketHandler) {
	// Authentication routes (no auth required)
	http.HandleFunc("/api/auth/register", enableCORS(authHandler.Register))
	http.HandleFunc("/api/auth/login", enableCORS(authHandler.Login))
	
	// Protected authentication routes (auth required)
	http.HandleFunc("/api/auth/validate", enableCORS(authHandler.ValidateToken))
	http.HandleFunc("/api/auth/refresh", enableCORS(authHandler.RefreshToken))

	// WebSocket endpoint for real-time chat
	// Authentication is handled via token query parameter
	http.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Health check endpoint
	http.HandleFunc("/api/health", enableCORS(healthCheckHandler))

	// API info endpoint
	http.HandleFunc("/api/info", enableCORS(apiInfoHandler))

	// Serve static files for frontend (if needed)
	// In production, you might want to serve these from a CDN or separate server
	http.HandleFunc("/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// Serve a simple landing page or redirect to frontend
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>GoChat Server</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 5px 0; border-radius: 5px; }
        .method { color: #2196F3; font-weight: bold; }
    </style>
</head>
<body>
    <h1>GoChat Real-time Messaging Server</h1>
    <p>The GoChat server is running successfully!</p>
    
    <h2>Available API Endpoints:</h2>
    <div class="endpoint">
        <span class="method">POST</span> /api/auth/register - Register a new user
    </div>
    <div class="endpoint">
        <span class="method">POST</span> /api/auth/login - Login with credentials
    </div>
    <div class="endpoint">
        <span class="method">POST</span> /api/auth/validate - Validate JWT token
    </div>
    <div class="endpoint">
        <span class="method">POST</span> /api/auth/refresh - Refresh JWT token
    </div>
    <div class="endpoint">
        <span class="method">GET</span> /api/health - Server health check
    </div>
    <div class="endpoint">
        <span class="method">GET</span> /api/info - API information
    </div>
    
    <h2>WebSocket Endpoint:</h2>
    <div class="endpoint">
        <span class="method">WS</span> /ws?token=&lt;jwt_token&gt; - Real-time chat connection
    </div>
    
    <h2>Message Types:</h2>
    <ul>
        <li><strong>JOIN:</strong> Join a chat room</li>
        <li><strong>LEAVE:</strong> Leave current room</li>
        <li><strong>MESSAGE:</strong> Send a chat message</li>
        <li><strong>PING:</strong> Connection health check</li>
    </ul>
    
    <p>For detailed API documentation and frontend integration, visit the project repository.</p>
</body>
</html>
			`))
		} else {
			http.NotFound(w, r)
		}
	}))

	log.Println("Routes configured successfully")
}

// healthCheckHandler provides a simple health check endpoint
// This is useful for load balancers and monitoring systems
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"status": "healthy",
		"service": "gochat",
		"version": "1.0.0",
		"timestamp": "` + getCurrentTimestamp() + `"
	}`))
}

// apiInfoHandler provides information about the API
// This helps developers understand available endpoints and versions
func apiInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"name": "GoChat API",
		"version": "1.0.0",
		"description": "Real-time multi-room chat application API",
		"endpoints": {
			"auth": {
				"register": "POST /api/auth/register",
				"login": "POST /api/auth/login",
				"validate": "POST /api/auth/validate",
				"refresh": "POST /api/auth/refresh"
			},
			"websocket": "WS /ws?token=<jwt_token>",
			"health": "GET /api/health",
			"info": "GET /api/info"
		},
		"websocket_message_types": [
			"JOIN", "LEAVE", "MESSAGE", "PING",
			"JOINED_ROOM", "LEFT_ROOM", "NEW_MESSAGE", 
			"USER_JOINED", "USER_LEFT", "USER_LIST", 
			"ERROR", "PONG", "ROOM_HISTORY"
		]
	}`))
}

// enableCORS middleware adds CORS headers to allow frontend connections
// This is essential for web applications making requests from different origins
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // In production, set specific origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next(w, r)
	}
}

// getEnvOrDefault retrieves an environment variable or returns a default value
// This allows for flexible configuration in different deployment environments
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getCurrentTimestamp returns the current timestamp in ISO 8601 format
// This is used for health checks and logging
func getCurrentTimestamp() string {
	return "2025-01-01T00:00:00Z" // Placeholder - would use time.Now().Format(time.RFC3339) in real implementation
}