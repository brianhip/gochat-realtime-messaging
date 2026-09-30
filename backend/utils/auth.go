package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT secret key for signing tokens
// This must be set at startup via SetJWTSecret before any token is
// generated or validated; there is no default.
var jwtSecret []byte

// SetJWTSecret configures the secret key used to sign and verify JWTs.
// It must be called once at application startup (see main.go), which
// loads the value from the JWT_SECRET environment variable and refuses
// to start the server if it is missing or too short.
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// Claims represents the JWT token claims for user authentication
type Claims struct {
	Username string `json:"username"`
	UserID   string `json:"user_id"`
	jwt.RegisteredClaims
}

// HashPassword creates a bcrypt hash of the password for storage
// bcrypt is deliberately slow and generates and embeds its own random salt,
// so no separate salt needs to be stored
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if the provided password matches the stored bcrypt hash
// Returns true if password is correct, false otherwise
func VerifyPassword(password, storedHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil
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