package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"gochat/db"
	"gochat/models"
	"gochat/utils"
)

// AuthService handles user authentication operations
// This service manages user registration, login, and token validation
type AuthService struct {
	userCollection *mongo.Collection
}

// NewAuthService creates a new authentication service instance
// It initializes the MongoDB collection for user operations
func NewAuthService() *AuthService {
	return &AuthService{
		userCollection: db.Database.Collection("users"),
	}
}

// RegisterUser creates a new user account with hashed password
// Returns the created user and JWT token, or error if registration fails
func (s *AuthService) RegisterUser(username, password string) (*models.User, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if username already exists
	var existingUser models.User
	err := s.userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&existingUser)
	if err == nil {
		return nil, "", errors.New("username already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, "", err
	}

	// Hash the password with bcrypt
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	// Create new user
	user := models.NewUser(username, passwordHash)

	// Insert user into database
	result, err := s.userCollection.InsertOne(ctx, user)
	if err != nil {
		return nil, "", err
	}

	// Set the generated ID
	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate JWT token for the new user
	token, err := utils.GenerateJWT(user.Username, user.ID.Hex())
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// LoginUser authenticates a user with username and password
// Returns user info and JWT token if credentials are valid
func (s *AuthService) LoginUser(username, password string) (*models.User, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by username
	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, "", errors.New("invalid username or password")
		}
		return nil, "", err
	}

	// Verify password
	if !utils.VerifyPassword(password, user.PasswordHash) {
		return nil, "", errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.Username, user.ID.Hex())
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

// ValidateToken verifies a JWT token and returns user information
// This is used to authenticate requests and WebSocket connections
func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	// Validate JWT token
	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		return nil, err
	}

	// Convert user ID from hex string to ObjectID
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	// Fetch user from database to ensure they still exist
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err = s.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// RefreshToken creates a new token with extended expiration
// This allows users to stay logged in without re-entering credentials
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	// Validate current token first
	_, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate new token with extended expiration
	return utils.RefreshJWT(tokenString)
}

// GetUserByID retrieves a user by their ObjectID
// This is useful for internal operations that need user details
func (s *AuthService) GetUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by their username
// This is useful for operations that work with usernames instead of IDs
func (s *AuthService) GetUserByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// UserExists checks if a user with the given username exists
// This is useful for validation without fetching full user data
func (s *AuthService) UserExists(username string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := s.userCollection.CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}