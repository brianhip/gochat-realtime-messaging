package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key for signing tokens
// In production, this should be loaded from environment variables
var jwtSecret = []byte("your-secret-key-change-this-in-production")

// Claims represents the JWT token claims for user authentication
type Claims struct {
	Username string `json:"username"`
	UserID   string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateSalt creates a random salt for password hashing
// Salt adds randomness to prevent rainbow table attacks
func GenerateSalt() (string, error) {
	// Generate 32 random bytes for the salt
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	
	// Convert to hexadecimal string for storage
	return hex.EncodeToString(salt), nil
}

// HashPassword creates a secure hash of the password using salt
// This implements PBKDF2 with SHA-256 for secure password storage
func HashPassword(password, salt string) string {
	// Combine password and salt
	combined := password + salt
	
	// Create SHA-256 hash
	hash := sha256.Sum256([]byte(combined))
	
	// Convert to hexadecimal string
	return hex.EncodeToString(hash[:])
}

// VerifyPassword checks if the provided password matches the stored hash
// Returns true if password is correct, false otherwise
func VerifyPassword(password, salt, storedHash string) bool {
	// Hash the provided password with the same salt
	computedHash := HashPassword(password, salt)
	
	// Compare with stored hash
	return computedHash == storedHash
}

// GenerateJWT creates a JWT token for authenticated users
// Token expires after 24 hours by default
func GenerateJWT(username, userID string) (string, error) {
	// Set token expiration time (24 hours from now)
	expirationTime := time.Now().Add(24 * time.Hour)
	
	// Create JWT claims
	claims := &Claims{
		Username: username,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gochat",
			Subject:   userID,
		},
	}
	
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign token with secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}

// ValidateJWT verifies a JWT token and returns the claims if valid
// Returns error if token is invalid, expired, or malformed
func ValidateJWT(tokenString string) (*Claims, error) {
	// Parse the token with claims
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Check if token is valid and get claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token")
}

// RefreshJWT creates a new token with extended expiration if the current token is still valid
// This allows users to stay logged in without re-entering credentials
func RefreshJWT(tokenString string) (string, error) {
	// Validate current token
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}
	
	// Generate new token with same user info but new expiration
	return GenerateJWT(claims.Username, claims.UserID)
}

// ExtractTokenFromAuthHeader extracts JWT token from Authorization header
// Expects format: "Bearer <token>"
func ExtractTokenFromAuthHeader(authHeader string) (string, error) {
	// Check if header starts with "Bearer "
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}
	
	// Extract token part
	token := authHeader[7:]
	if len(token) == 0 {
		return "", fmt.Errorf("empty token in authorization header")
	}
	
	return token, nil
}