package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"gochat/services"
	"gochat/utils"
)

// AuthHandler handles HTTP authentication endpoints
// This handler manages user registration, login, and token validation
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents the JSON payload for user registration
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents the JSON payload for user login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the JSON response for successful authentication
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
	User    any    `json:"user,omitempty"`
}

// ErrorResponse represents the JSON response for errors
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Register handles user registration requests
// POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Set response headers for JSON
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Method not allowed",
			Code:    "METHOD_NOT_ALLOWED",
		})
		return
	}

	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Invalid JSON payload",
			Code:    "INVALID_JSON",
		})
		return
	}

	// Validate input
	if err := h.validateRegistrationInput(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: err.Error(),
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Attempt to register user
	user, token, err := h.authService.RegisterUser(req.Username, req.Password)
	if err != nil {
		// Check for specific error types
		statusCode := http.StatusInternalServerError
		errorCode := "REGISTRATION_FAILED"
		
		if strings.Contains(err.Error(), "username already exists") {
			statusCode = http.StatusConflict
			errorCode = "USERNAME_EXISTS"
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: err.Error(),
			Code:    errorCode,
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "User registered successfully",
		Token:   token,
		User:    user,
	})
}

// Login handles user login requests
// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Set response headers for JSON
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Method not allowed",
			Code:    "METHOD_NOT_ALLOWED",
		})
		return
	}

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Invalid JSON payload",
			Code:    "INVALID_JSON",
		})
		return
	}

	// Validate input
	if err := h.validateLoginInput(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: err.Error(),
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Attempt to login user
	user, token, err := h.authService.LoginUser(req.Username, req.Password)
	if err != nil {
		// Check for specific error types
		statusCode := http.StatusUnauthorized
		errorCode := "LOGIN_FAILED"
		
		if strings.Contains(err.Error(), "invalid username or password") {
			statusCode = http.StatusUnauthorized
			errorCode = "INVALID_CREDENTIALS"
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: err.Error(),
			Code:    errorCode,
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		User:    user,
	})
}

// ValidateToken handles token validation requests
// POST /api/auth/validate
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// Set response headers for JSON
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Method not allowed",
			Code:    "METHOD_NOT_ALLOWED",
		})
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Authorization header required",
			Code:    "MISSING_AUTH_HEADER",
		})
		return
	}

	// Extract token from "Bearer <token>" format
	token, err := utils.ExtractTokenFromAuthHeader(authHeader)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: err.Error(),
			Code:    "INVALID_AUTH_HEADER",
		})
		return
	}

	// Validate token
	user, err := h.authService.ValidateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Invalid or expired token",
			Code:    "INVALID_TOKEN",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "Token is valid",
		User:    user,
	})
}

// RefreshToken handles token refresh requests
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Set response headers for JSON
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Method not allowed",
			Code:    "METHOD_NOT_ALLOWED",
		})
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Authorization header required",
			Code:    "MISSING_AUTH_HEADER",
		})
		return
	}

	// Extract token from "Bearer <token>" format
	token, err := utils.ExtractTokenFromAuthHeader(authHeader)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: err.Error(),
			Code:    "INVALID_AUTH_HEADER",
		})
		return
	}

	// Refresh token
	newToken, err := h.authService.RefreshToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: false,
			Message: "Unable to refresh token",
			Code:    "REFRESH_FAILED",
		})
		return
	}

	// Return success response with new token
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Token:   newToken,
	})
}

// validateRegistrationInput validates user registration input
func (h *AuthHandler) validateRegistrationInput(req RegisterRequest) error {
	// Check username length
	if len(req.Username) < 3 {
		return &ValidationError{"Username must be at least 3 characters long"}
	}
	if len(req.Username) > 30 {
		return &ValidationError{"Username must be less than 30 characters"}
	}

	// Check for valid username characters (alphanumeric and underscore)
	for _, char := range req.Username {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_') {
			return &ValidationError{"Username can only contain letters, numbers, and underscores"}
		}
	}

	// Check password length
	if len(req.Password) < 6 {
		return &ValidationError{"Password must be at least 6 characters long"}
	}
	if len(req.Password) > 128 {
		return &ValidationError{"Password must be less than 128 characters"}
	}

	return nil
}

// validateLoginInput validates user login input
func (h *AuthHandler) validateLoginInput(req LoginRequest) error {
	// Check if username is provided
	if strings.TrimSpace(req.Username) == "" {
		return &ValidationError{"Username is required"}
	}

	// Check if password is provided
	if strings.TrimSpace(req.Password) == "" {
		return &ValidationError{"Password is required"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// AuthMiddleware provides middleware for protecting routes that require authentication
func (h *AuthHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set response headers for JSON
		w.Header().Set("Content-Type", "application/json")

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Success: false,
				Message: "Authorization header required",
				Code:    "MISSING_AUTH_HEADER",
			})
			return
		}

		// Extract token from "Bearer <token>" format
		token, err := utils.ExtractTokenFromAuthHeader(authHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Success: false,
				Message: err.Error(),
				Code:    "INVALID_AUTH_HEADER",
			})
			return
		}

		// Validate token
		user, err := h.authService.ValidateToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Success: false,
				Message: "Invalid or expired token",
				Code:    "INVALID_TOKEN",
			})
			return
		}

		// Add user info to request context for use in handlers
		// Store user in a way that can be accessed by subsequent handlers
		r.Header.Set("X-User-ID", user.ID.Hex())
		r.Header.Set("X-Username", user.Username)

		// Call the next handler
		next(w, r)
	}
}