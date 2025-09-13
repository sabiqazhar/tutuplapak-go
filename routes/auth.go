// routes/auth.go
package routes

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"tutuplapak-go/repository" // sqlc generated code
	"tutuplapak-go/utils"      // token management

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Queries *repository.Queries // <-- sqlc Queries struct
}

func NewAuthHandler(queries *repository.Queries) *AuthHandler {
	return &AuthHandler{Queries: queries}
}

// Request structs
type RegisterEmailRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type RegisterPhoneRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type LoginEmailRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type LoginPhoneRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

// Response struct
type AuthResponse struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
	Token string `json:"token"`
}

// Utility functions
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func validatePhone(phone string) bool {
	// Phone should begin with + and international calling number
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

// Email Registration - POST /v1/register/email
func (h *AuthHandler) RegisterEmail(c *gin.Context) {
	var req RegisterEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Check if user already exists
	_, err := h.Queries.GetUserByEmail(c, req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Create user
	user, err := h.Queries.CreateUserWithEmail(c, repository.CreateUserWithEmailParams{
		Email:    req.Email,
		Password: string(hashedPassword),
		Phone:    "", // Empty phone initially
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Generate token
	token := generateToken()

	// Store token (you might want to save this in DB or cache)
	// For now, we'll just generate it

	c.JSON(http.StatusCreated, AuthResponse{
		Email: user.Email,
		Phone: "", // Empty string if first registering
		Token: token,
	})
}

// Phone Registration - POST /v1/register/phone
func (h *AuthHandler) RegisterPhone(c *gin.Context) {
	var req RegisterPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate phone format
	if !validatePhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Check if phone already exists
	_, err := h.Queries.GetUserByPhone(c, req.Phone)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Phone already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Create user with phone
	user, err := h.Queries.CreateUserWithPhone(c, repository.CreateUserWithPhoneParams{
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Email:    "", // Empty email initially
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Generate token
	token := generateToken()

	c.JSON(http.StatusCreated, AuthResponse{
		Email: "", // Empty string if first registering
		Phone: user.Phone,
		Token: token,
	})
}

// Email Login - POST /v1/login/email
func (h *AuthHandler) LoginEmail(c *gin.Context) {
	var req LoginEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Find user by email
	user, err := h.Queries.GetUserByEmail(c, req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		return
	}

	// Generate token
	token := generateToken()

	c.JSON(http.StatusOK, AuthResponse{
		Email: user.Email,
		Phone: user.Phone, // Could be empty if not linked
		Token: token,
	})
}

// Phone Login - POST /v1/login/phone
func (h *AuthHandler) LoginPhone(c *gin.Context) {
	var req LoginPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate phone format
	if !validatePhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Find user by phone
	user, err := h.Queries.GetUserByPhone(c, req.Phone)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Phone not found"})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Phone not found"})
		return
	}

	// Generate token
	token := generateToken()

	// Store token with user ID
	utils.GlobalTokenStore.StoreToken(token, user.ID)

	c.JSON(http.StatusOK, AuthResponse{
		Email: user.Email, // Could be empty if not linked
		Phone: user.Phone,
		Token: token,
	})
}
