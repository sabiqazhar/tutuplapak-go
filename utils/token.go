package utils

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Simple in-memory token store (use Redis or DB in production)
type TokenStore struct {
	tokens map[string]TokenData
	mutex  sync.RWMutex
}

type TokenData struct {
	UserID    int32
	CreatedAt time.Time
	ExpiresAt time.Time
}

var GlobalTokenStore = &TokenStore{
	tokens: make(map[string]TokenData),
}

func GenerateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (ts *TokenStore) StoreToken(token string, userID int32) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.tokens[token] = TokenData{
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Token expires in 24 hours
	}
}

func (ts *TokenStore) GetUserID(token string) (int32, bool) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	tokenData, exists := ts.tokens[token]
	if !exists {
		return 0, false
	}
	// Check if token is expired
	if time.Now().After(tokenData.ExpiresAt) {
		// Clean up expired token
		delete(ts.tokens, token)
		return 0, false
	}
	return tokenData.UserID, true
}

func (ts *TokenStore) DeleteToken(token string) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	delete(ts.tokens, token)
}

// Cleanup expired tokens periodically
func (ts *TokenStore) CleanupExpiredTokens() {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	now := time.Now()
	for token, tokenData := range ts.tokens {
		if now.After(tokenData.ExpiresAt) {
			delete(ts.tokens, token)
		}
	}
}

// Start cleanup goroutine
func (ts *TokenStore) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			ts.CleanupExpiredTokens()
		}
	}()
}

func GenerateJWTToken(userID uint) (string, error) {
	tokenString := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 2).Unix(),
	})

	token, err := tokenString.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}
	return token, nil
}
