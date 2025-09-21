package routes

import (
	"database/sql"
	"net/http"

	"tutuplapak-go/repository"
	"tutuplapak-go/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Queries *repository.Queries
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

// Email Registration - POST /v1/register/email
func (h *AuthHandler) RegisterEmail(c *gin.Context) {
	var req RegisterEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Check if user already exists
	_, err := h.Queries.GetUserByEmail(c, sql.NullString{String: req.Email, Valid: true})
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
		Email:    sql.NullString{String: req.Email, Valid: true},
		Password: string(hashedPassword),
		Phone:    sql.NullString{String: "", Valid: false}, // Empty phone initially
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Generate token
	token, err := utils.GenerateJWTToken(uint(user.ID))
	if err != nil {
		utils.Logger.Error().Err(err)
		utils.Logger.Error().Msg("failed to create token")
	}
	c.JSON(http.StatusCreated, AuthResponse{
		Email: user.Email.String,
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
	if !utils.ValidatePhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Check if phone already exists
	_, err := h.Queries.GetUserByPhone(c, sql.NullString{String: req.Phone, Valid: true})
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
		Phone:    sql.NullString{String: req.Phone, Valid: true},
		Password: string(hashedPassword),
		Email:    sql.NullString{String: "", Valid: false}, // Empty email initially
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Generate token
	token, err := utils.GenerateJWTToken(uint(user.ID))
	if err != nil {
		utils.Logger.Error().Err(err)
		utils.Logger.Error().Msg("failed to create token")
	}
	c.JSON(http.StatusCreated, AuthResponse{
		Email: "", // Empty string if first registering
		Phone: user.Phone.String,
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
	user, err := h.Queries.GetUserByEmail(c, sql.NullString{String: req.Email, Valid: true})
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
	token, err := utils.GenerateJWTToken(uint(user.ID))
	if err != nil {
		utils.Logger.Error().Err(err)
		utils.Logger.Error().Msg("failed to create token")
	}
	c.JSON(http.StatusOK, AuthResponse{
		Email: user.Email.String,
		Phone: utils.NullStringToString(user.Phone), // Could be empty if not linked
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
	if !utils.ValidatePhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Find user by phone
	user, err := h.Queries.GetUserByPhone(c, sql.NullString{String: req.Phone, Valid: true})
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
	token, err := utils.GenerateJWTToken(uint(user.ID))
	if err != nil {
		utils.Logger.Error().Err(err)
		utils.Logger.Error().Msg("failed to create token")
	}
	// Store token with user ID
	utils.GlobalTokenStore.StoreToken(token, user.ID)
	c.JSON(http.StatusOK, AuthResponse{
		Email: utils.NullStringToString(user.Email), // Could be empty if not linked
		Phone: user.Phone.String,
		Token: token,
	})
}
