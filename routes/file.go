// routes/file.go
package routes

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"tutuplapak-go/repository"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	Queries *repository.Queries
}

func NewFileHandler(queries *repository.Queries) *FileHandler {
	return &FileHandler{Queries: queries}
}

type FileUploadResponse struct {
	FileID           string `json:"fileId"`
	FileURI          string `json:"fileUri"`
	FileThumbnailURI string `json:"fileThumbnailUri"`
}

// POST /v1/file
func (h *FileHandler) UploadFile(c *gin.Context) {
	// Get the file from form data
	file, header, err := c.Request.FormFile("flle") // Note: typo in requirement "flle" not "file"
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidImageType(header) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate file size (max 100KiB = 102400 bytes)
	if header.Size > 102400 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Generate unique file ID
	fileID := generateFileID()

	// Get file extension
	ext := filepath.Ext(header.Filename)

	// Create file paths
	fileName := fileID + ext
	filePath := filepath.Join("uploads", fileName)
	thumbnailName := fileID + "_thumb" + ext
	thumbnailPath := filepath.Join("uploads", thumbnailName)

	// Ensure uploads directory exists
	os.MkdirAll("uploads", 0o755)

	// Save original file
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Create thumbnail (simplified - in production use image processing library)
	err = createThumbnail(filePath, thumbnailPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// In production, upload to AWS S3 and get actual URIs
	// For now, we'll use local paths
	fileURI := fmt.Sprintf("http://localhost:8080/uploads/%s", fileName)
	thumbnailURI := fmt.Sprintf("http://localhost:8080/uploads/%s", thumbnailName)

	// Save file info to database
	savedFile, err := h.Queries.CreateFile(c, repository.CreateFileParams{
		FileUri: fileURI,
		FileThumnailUri: sql.NullString{
			String: thumbnailURI,
			Valid:  true,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	response := FileUploadResponse{
		FileID:           fmt.Sprintf("%d", savedFile.ID), // Convert ID to string
		FileURI:          savedFile.FileUri,
		FileThumbnailURI: savedFile.FileThumnailUri.String,
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions
func generateFileID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func isValidImageType(header *multipart.FileHeader) bool {
	contentType := header.Header.Get("Content-Type")
	validTypes := []string{"image/jpeg", "image/jpg", "image/png"}

	if slices.Contains(validTypes, contentType) {
		return true
	}

	// Also check file extension as fallback
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := []string{".jpg", ".jpeg", ".png"}

	return slices.Contains(validExts, ext)
}

func createThumbnail(sourcePath, thumbnailPath string) error {
	// Simplified thumbnail creation - just copy the file
	// In production, use image processing library like imaging or resize
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(thumbnailPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
