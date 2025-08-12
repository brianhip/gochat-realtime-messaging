package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the chat application
// This struct defines how user data is stored in MongoDB and used throughout the application
type User struct {
	// ID is the unique identifier for each user in MongoDB
	// primitive.ObjectID is MongoDB's native ID type
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	
	// Username is the unique display name for the user
	// It must be unique across all users in the system
	Username string `bson:"username" json:"username"`
	
	// PasswordHash stores the hashed version of the user's password
	// We never store plain text passwords for security reasons
	PasswordHash string `bson:"password_hash" json:"-"` // json:"-" means this field won't be included in JSON responses
	
	// Salt is a random value used in password hashing to prevent rainbow table attacks
	// Each user gets a unique salt to make password hashing more secure
	Salt string `bson:"salt" json:"-"`
	
	// CreatedAt tracks when the user account was created
	// This is useful for analytics and user management
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// NewUser creates a new User instance with the current timestamp
// This is a constructor function that ensures CreatedAt is always set
func NewUser(username, passwordHash, salt string) *User {
	return &User{
		Username:     username,
		PasswordHash: passwordHash,
		Salt:         salt,
		CreatedAt:    time.Now(),
	}
}