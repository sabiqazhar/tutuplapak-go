package routes

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tutuplapak-go/repository"
	"tutuplapak-go/utils"

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
	Category string `json:"category" binding:"required,oneof=Food Beverage Clothes Furniture Tools"`
	Qty      int32  `json:"qty" binding:"required,min=1"`
	Price    int32  `json:"price" binding:"required,min=100"`
	Sku      string `json:"sku" binding:"required,max=32"`
	FileID   string `json:"fileId" binding:"required"`
}

// Response DTO
type ProductResponse struct {
	ProductID        string    `json:"productId"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Qty              int32     `json:"qty"`
	Price            int32     `json:"price"`
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
		utils.Logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Additional guard: SKU must not be empty or whitespace-only
	if strings.TrimSpace(req.Sku) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate file ID
	fileIDInt, err := strconv.Atoi(req.FileID)
	if err != nil {
		utils.Logger.Error().Err(err).Msg("Invalid file id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is not valid"})
		return
	}

	// Check if file exists
	file, err := h.Queries.GetFileByID(c, int32(fileIDInt))
	if err != nil {
		utils.Logger.Error().Err(err).Msg("File not found")
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is not valid"})
		return
	}

	// Get category ID from category name
	categoryID, err := h.Queries.GetProductCategoryByName(c, sql.NullString{String: req.Category, Valid: true})
	if err != nil {
		utils.Logger.Error().Err(err).Msg("Invalid category")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.Logger.Error().Err(err).Msg("User ID from context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if SKU already exists for this user
	if req.Sku != "" {
		_, err = h.Queries.GetProductBySKUAndUserID(c, repository.GetProductBySKUAndUserIDParams{
			Sku:    sql.NullString{String: req.Sku, Valid: true},
			UserID: sql.NullInt32{Int32: userID, Valid: true},
		})
		if err == nil {
			utils.Logger.Error().Msg("SKU already exists for this user")
			c.JSON(http.StatusConflict, gin.H{"error": "sku already exists (per account basis)"})
			return
		}
	}

	// Create product
	product, err := h.Queries.CreateProduct(c, repository.CreateProductParams{
		UserID:   sql.NullInt32{Int32: userID, Valid: true},
		Name:     sql.NullString{String: req.Name, Valid: true},
		Category: sql.NullInt32{Int32: categoryID, Valid: true},
		Qty:      sql.NullInt32{Int32: req.Qty, Valid: true},
		Price:    sql.NullString{String: fmt.Sprintf("%d", req.Price), Valid: true},
		Sku:      sql.NullString{String: req.Sku, Valid: true},
		FileID:   sql.NullInt32{Int32: int32(fileIDInt), Valid: true},
	})
	if err != nil {
		utils.Logger.Error().Err(err).Msg("Failed to create product")
		// Check if it's a unique constraint violation for SKU
		if err.Error() != "" && (strings.Contains(err.Error(), "unique_sku_per_user") || strings.Contains(err.Error(), "duplicate key")) {
			c.JSON(http.StatusConflict, gin.H{"error": "sku already exists (per account basis)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Build response
	response := ProductResponse{
		ProductID:        strconv.FormatInt(int64(product.ProductID), 10),
		Name:             product.Name.String,
		Category:         req.Category,
		Qty:              product.Qty.Int32,
		Price:            req.Price,
		Sku:              product.Sku.String,
		FileID:           strconv.FormatInt(int64(product.FileID.Int32), 10),
		FileURI:          file.FileUri,
		FileThumbnailURI: file.FileThumnailUri.String,
		CreatedAt:        product.CreatedAt.Time,
		UpdatedAt:        product.UpdatedAt.Time,
	}

	c.JSON(http.StatusCreated, response)
}

// Response struct for a single product
type GetProductResponse struct {
	ProductID        string    `json:"productId"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Qty              int32     `json:"qty"`
	Price            int32     `json:"price"`
	Sku              string    `json:"sku"`
	FileID           string    `json:"fileId"`
	FileURI          string    `json:"fileUri"`
	FileThumbnailURI string    `json:"fileThumbnailUri"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// GET /v1/product
func (h *ProductHandler) GetProducts(c *gin.Context) {
	// Default values for limit and offset
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if err != nil || limit < 0 {
		limit = 5
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// productId filter
	var productID sql.NullInt32
	productIDStr := c.Query("productId")
	if productIDStr != "" {
		id, err := strconv.Atoi(productIDStr)
		if err == nil {
			productID = sql.NullInt32{Int32: int32(id), Valid: true}
		}
	}

	// sku filter
	var sku sql.NullString
	skuStr := c.Query("sku")
	if skuStr != "" {
		// Validate that SKU is not purely numeric
		if _, err := strconv.Atoi(skuStr); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SKU format"})
			return
		}
		sku = sql.NullString{String: skuStr, Valid: true}
	}

	// category filter
	var category sql.NullString
	categoryStr := c.Query("category")
	if categoryStr != "" {
		// Validate that the category exists in the database
		_, err := h.Queries.GetProductCategoryByName(c, sql.NullString{String: categoryStr, Valid: true})
		if err != nil {
			// If no rows are returned, the category is invalid
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
				return
			}
			// Handle other potential database errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error while validating category"})
			return
		}
		category = sql.NullString{String: categoryStr, Valid: true}
	}

	// sortBy filter
	var sortBy sql.NullString
	sortByStr := c.Query("sortBy")
	validSorts := []string{"newest", "oldest", "cheapest", "expensive"}
	for _, s := range validSorts {
		if sortByStr == s {
			sortBy = sql.NullString{String: sortByStr, Valid: true}
			break
		}
	}

	// Call repository to get products
	products, err := h.Queries.ListProducts(c, repository.ListProductsParams{
		ProductID: productID,
		Sku:       sku,
		Category:  category,
		SortBy:    sortBy,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Build response
	var response []GetProductResponse
	for _, p := range products {
		fileURI, fileThumbnailURI, err := utils.GetFileInfo(h.Queries, c, p.FileID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error fetching file info"})
			return
		}

		priceInt, _ := strconv.Atoi(utils.NullStringToString(p.Price))
		response = append(response, GetProductResponse{
			ProductID:        fmt.Sprintf("%d", p.ProductID),
			Name:             utils.NullStringToString(p.Name),
			Category:         utils.NullStringToString(p.CategoryName),
			Qty:              p.Qty.Int32,
			Price:            int32(priceInt),
			Sku:              utils.NullStringToString(p.Sku),
			FileID:           utils.NullInt32ToString(p.FileID),
			FileURI:          fileURI,
			FileThumbnailURI: fileThumbnailURI,
			CreatedAt:        p.CreatedAt.Time,
			UpdatedAt:        p.UpdatedAt.Time,
		})
	}

	// If no products found, return empty array instead of null
	if response == nil {
		response = make([]GetProductResponse, 0)
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "productId is not found"})
		return
	}

	// Check if product exists and belongs to user
	existingProduct, err := h.Queries.GetProductByID(c, int32(productID))
	if err != nil || existingProduct.UserID.Int32 != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "productId is not found"})
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Additional guard: SKU must not be empty or whitespace-only
	if strings.TrimSpace(req.Sku) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate file ID
	fileIDInt, err := strconv.Atoi(req.FileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is not valid"})
		return
	}
	file, err := h.Queries.GetFileByID(c, int32(fileIDInt))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is not valid / exists"})
		return
	}

	// Get category ID from category name
	categoryID, err := h.Queries.GetProductCategoryByName(c, sql.NullString{String: req.Category, Valid: true})
	if err != nil {
		utils.Logger.Error().Err(err).Msg("Invalid category")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	// Check if SKU already exists for this user (excluding current product)
	if req.Sku != "" && req.Sku != existingProduct.Sku.String {
		existingSku, err := h.Queries.GetProductBySKUAndUserID(c, repository.GetProductBySKUAndUserIDParams{
			Sku:    sql.NullString{String: req.Sku, Valid: true},
			UserID: sql.NullInt32{Int32: userID, Valid: true},
		})
		if err == nil && existingSku.ProductID != int32(productID) {
			utils.Logger.Error().Msg("SKU already exists for this user")
			c.JSON(http.StatusConflict, gin.H{"error": "sku already exists (per account basis)"})
			return
		}
	}

	updatedProduct, err := h.Queries.UpdateProduct(c, repository.UpdateProductParams{
		ProductID: int32(productID),
		Name:      sql.NullString{String: req.Name, Valid: true},
		Category:  sql.NullInt32{Int32: categoryID, Valid: true},
		Qty:       sql.NullInt32{Int32: req.Qty, Valid: true},
		Price:     sql.NullString{String: fmt.Sprintf("%d", req.Price), Valid: true},
		Sku:       sql.NullString{String: req.Sku, Valid: true},
		FileID:    sql.NullInt32{Int32: int32(fileIDInt), Valid: true},
		UserID:    sql.NullInt32{Int32: userID, Valid: true},
	})
	if err != nil {
		// Check if it's a unique constraint violation for SKU
		if err.Error() != "" && (strings.Contains(err.Error(), "unique_sku_per_user") || strings.Contains(err.Error(), "duplicate key")) {
			c.JSON(http.StatusConflict, gin.H{"error": "sku already exists (per account basis)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error while updating"})
		return
	}

	// Build response
	response := ProductResponse{
		ProductID:        strconv.FormatInt(int64(updatedProduct.ProductID), 10),
		Name:             updatedProduct.Name.String,
		Category:         req.Category,
		Qty:              updatedProduct.Qty.Int32,
		Price:            req.Price,
		Sku:              updatedProduct.Sku.String,
		FileID:           strconv.FormatInt(int64(updatedProduct.FileID.Int32), 10),
		FileURI:          file.FileUri,
		FileThumbnailURI: file.FileThumnailUri.String,
		CreatedAt:        updatedProduct.CreatedAt.Time,
		UpdatedAt:        updatedProduct.UpdatedAt.Time,
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "productId is not found"})
		return
	}

	product, err := h.Queries.GetProductByID(c, int32(productID))
	if err != nil || product.UserID.Int32 != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "productId is not found"})
		return
	}

	err = h.Queries.DeleteProduct(c, repository.DeleteProductParams{
		ProductID: int32(productID),
		UserID:    sql.NullInt32{Int32: userID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.Status(http.StatusOK)
}
