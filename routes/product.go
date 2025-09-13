package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"tutuplapak-go/repository"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	Queries *repository.Queries
}

func NewProductHandler(queries *repository.Queries) *ProductHandler {
	return &ProductHandler{Queries: queries}
}

// Request DTO
type CreateProductRequest struct {
	Name     string `json:"name" binding:"required,min=4,max=32"`
	Category int32  `json:"category" binding:"required"`
	Qty      int32  `json:"qty" binding:"required,min=1"`
	Price    int32  `json:"price" binding:"required,min=100"`
	Sku      string `json:"sku" binding:"required,min=0,max=32"`
	FileID   string `json:"fileId" binding:"required"`
}

// Response DTO
type ProductResponse struct {
	ProductID        string    `json:"productId"`
	Name             string    `json:"name"`
	Category         int32     `json:"category"`
	Qty              int32     `json:"qty"`
	Price            float64   `json:"price"`
	Sku              string    `json:"sku"`
	FileID           string    `json:"fileId"`
	FileURI          string    `json:"fileUri"`
	FileThumbnailURI string    `json:"fileThumbnailUri"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}


// POST /v1/product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate file ID
	fileIDInt, err := strconv.Atoi(req.FileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is not valid"})
		return
	}

	// Check if file exists
	file, err := h.Queries.GetFileByID(c, int32(fileIDInt))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is not valid"})
		return
	}

	// Check if SKU already exists for this user
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	_, err = h.Queries.GetProductBySKUAndUserID(c, repository.GetProductBySKUAndUserIDParams{
		Sku:    sql.NullString{String: req.Sku, Valid: true},
		UserID: sql.NullInt32{Int32: userID, Valid: true},
	})
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "sku already exists"})
		return
	}

	// Create product
	product, err := h.Queries.CreateProduct(c, repository.CreateProductParams{
		UserID:   sql.NullInt32{Int32: userID, Valid: true},
		Name:     sql.NullString{String: req.Name, Valid: true},
		Category: sql.NullInt32{Int32: req.Category, Valid: true},
		Qty:      sql.NullInt32{Int32: req.Qty, Valid: true},
		Price:    sql.NullString{String: fmt.Sprintf("%d", req.Price), Valid: true},
		Sku:      sql.NullString{String: req.Sku, Valid: true},
		FileID:   sql.NullInt32{Int32: int32(fileIDInt), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Build response
	response := ProductResponse{
		ProductID:        strconv.FormatInt(int64(product.ProductID), 10),
		Name:             product.Name.String,
		Category:         product.Category.Int32,
		Qty:              product.Qty.Int32,
		Price:            float64(req.Price),
		Sku:              product.Sku.String,
		FileID:           strconv.FormatInt(int64(product.FileID.Int32), 10),
		FileURI:          file.FileUri,
		FileThumbnailURI: file.FileThumnailUri.String,
		CreatedAt:        product.CreatedAt.Time,
		UpdatedAt:        product.UpdatedAt.Time,
	}

	c.JSON(http.StatusCreated, response)
}
