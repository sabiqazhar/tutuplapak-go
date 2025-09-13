// routes/profile.go
package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"tutuplapak-go/repository"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	Queries *repository.Queries
}

func NewProfileHandler(queries *repository.Queries) *ProfileHandler {
	return &ProfileHandler{Queries: queries}
}

// Request structs
type UpdateProfileRequest struct {
	FileID            string `json:"fileId"`
	BankAccountName   string `json:"bankAccountName" binding:"required,min=4,max=32"`
	BankAccountHolder string `json:"bankAccountHolder" binding:"required,min=4,max=32"`
	BankAccountNumber string `json:"bankAccountNumber" binding:"required,min=4,max=32"`
}

type LinkPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type LinkEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Response struct
type ProfileResponse struct {
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	FileID            string `json:"fileId"`
	FileURI           string `json:"fileUri"`
	FileThumbnailURI  string `json:"fileThumbnailUri"`
	BankAccountName   string `json:"bankAccountName"`
	BankAccountHolder string `json:"bankAccountHolder"`
	BankAccountNumber string `json:"bankAccountNumber"`
}

// Helper function to get user ID from gin context
func getUserIDFromContext(c *gin.Context) (int32, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user not found in context")
	}
	return userID.(int32), nil
}

// GET /v1/user
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.Queries.GetUserByID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Get file information if fileId exists
	var fileURI, fileThumbnailURI string
	if user.FileID.Valid && user.FileID.Int32 > 0 {
		file, err := h.Queries.GetFileByID(c, user.FileID.Int32)
		if err == nil {
			fileURI = file.FileUri
			if file.FileThumnailUri.Valid {
				fileThumbnailURI = file.FileThumnailUri.String
			}
		}
	}

	response := ProfileResponse{
		Email:             user.Email,
		Phone:             user.Phone,
		FileID:            getStringFromNullInt32(user.FileID),
		FileURI:           fileURI,
		FileThumbnailURI:  fileThumbnailURI,
		BankAccountName:   getStringFromNullString(user.BankAccountName),
		BankAccountHolder: getStringFromNullString(user.BankAccountHolder),
		BankAccountNumber: getStringFromNullString(user.BankAccountNumber),
	}

	c.JSON(http.StatusOK, response)
}

// PUT /v1/user
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate fileId if provided
	var fileID sql.NullInt32
	if req.FileID != "" {
		// Convert string fileId to int32
		fileIDInt, err := strconv.Atoi(req.FileID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is not valid"})
			return
		}

		// Check if file exists
		_, err = h.Queries.GetFileByID(c, int32(fileIDInt))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is not valid"})
			return
		}
		fileID = sql.NullInt32{Int32: int32(fileIDInt), Valid: true}
	}

	// Update user profile
	updatedUser, err := h.Queries.UpdateUserProfile(c, repository.UpdateUserProfileParams{
		ID:                userID,
		FileID:            fileID,
		BankAccountName:   sql.NullString{String: req.BankAccountName, Valid: true},
		BankAccountHolder: sql.NullString{String: req.BankAccountHolder, Valid: true},
		BankAccountNumber: sql.NullString{String: req.BankAccountNumber, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Get file information if fileId exists
	var fileURI, fileThumbnailURI string
	if updatedUser.FileID.Valid && updatedUser.FileID.Int32 > 0 {
		file, err := h.Queries.GetFileByID(c, updatedUser.FileID.Int32)
		if err == nil {
			fileURI = file.FileUri
			if file.FileThumnailUri.Valid {
				fileThumbnailURI = file.FileThumnailUri.String
			}
		}
	}

	response := ProfileResponse{
		Email:             updatedUser.Email,
		Phone:             updatedUser.Phone,
		FileID:            getStringFromNullInt32(updatedUser.FileID),
		FileURI:           fileURI,
		FileThumbnailURI:  fileThumbnailURI,
		BankAccountName:   getStringFromNullString(updatedUser.BankAccountName),
		BankAccountHolder: getStringFromNullString(updatedUser.BankAccountHolder),
		BankAccountNumber: getStringFromNullString(updatedUser.BankAccountNumber),
	}

	c.JSON(http.StatusOK, response)
}

// POST /v1/user/link/phone
func (h *ProfileHandler) LinkPhone(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req LinkPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate phone format
	if !validatePhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Check if phone is already taken
	_, err = h.Queries.GetUserByPhone(c, req.Phone)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Phone is taken"})
		return
	}

	// Link phone to user
	updatedUser, err := h.Queries.LinkPhoneToUser(c, repository.LinkPhoneToUserParams{
		ID:    userID,
		Phone: req.Phone,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Get file information if fileId exists
	var fileURI, fileThumbnailURI string
	if updatedUser.FileID.Valid && updatedUser.FileID.Int32 > 0 {
		file, err := h.Queries.GetFileByID(c, updatedUser.FileID.Int32)
		if err == nil {
			fileURI = file.FileUri
			if file.FileThumnailUri.Valid {
				fileThumbnailURI = file.FileThumnailUri.String
			}
		}
	}

	response := ProfileResponse{
		Email:             updatedUser.Email,
		Phone:             updatedUser.Phone,
		FileID:            getStringFromNullInt32(updatedUser.FileID),
		FileURI:           fileURI,
		FileThumbnailURI:  fileThumbnailURI,
		BankAccountName:   getStringFromNullString(updatedUser.BankAccountName),
		BankAccountHolder: getStringFromNullString(updatedUser.BankAccountHolder),
		BankAccountNumber: getStringFromNullString(updatedUser.BankAccountNumber),
	}

	c.JSON(http.StatusOK, response)
}

// POST /v1/user/link/email
func (h *ProfileHandler) LinkEmail(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req LinkEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Check if email is already taken
	_, err = h.Queries.GetUserByEmail(c, req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email is taken"})
		return
	}

	// Link email to user
	updatedUser, err := h.Queries.LinkEmailToUser(c, repository.LinkEmailToUserParams{
		ID:    userID,
		Email: req.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Get file information if fileId exists
	var fileURI, fileThumbnailURI string
	if updatedUser.FileID.Valid && updatedUser.FileID.Int32 > 0 {
		file, err := h.Queries.GetFileByID(c, updatedUser.FileID.Int32)
		if err == nil {
			fileURI = file.FileUri
			if file.FileThumnailUri.Valid {
				fileThumbnailURI = file.FileThumnailUri.String
			}
		}
	}

	response := ProfileResponse{
		Email:             updatedUser.Email,
		Phone:             updatedUser.Phone,
		FileID:            getStringFromNullInt32(updatedUser.FileID),
		FileURI:           fileURI,
		FileThumbnailURI:  fileThumbnailURI,
		BankAccountName:   getStringFromNullString(updatedUser.BankAccountName),
		BankAccountHolder: getStringFromNullString(updatedUser.BankAccountHolder),
		BankAccountNumber: getStringFromNullString(updatedUser.BankAccountNumber),
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions
func validatePhone(phone string) bool {
	// Phone should begin with + and international calling number
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

func getStringFromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func getStringFromNullInt32(ni sql.NullInt32) string {
	if ni.Valid {
		return strconv.FormatInt(int64(ni.Int32), 10)
	}
	return ""
}
