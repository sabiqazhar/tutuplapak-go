package routes

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"tutuplapak-go/repository"
	"tutuplapak-go/utils"

	"github.com/gin-gonic/gin"
)

type PurchaseHandler struct {
	Queries *repository.Queries
	DB      *sql.DB
}

func NewPurchaseHandler(queries *repository.Queries, db *sql.DB) *PurchaseHandler {
	return &PurchaseHandler{Queries: queries, DB: db}
}

// Request structs
type PurchasedItemRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Qty       int32  `json:"qty" binding:"required,min=1"`
}

type CreatePurchaseRequest struct {
	PurchasedItems      []PurchasedItemRequest `json:"purchasedItems" binding:"required,min=1"`
	SenderName          string                 `json:"senderName" binding:"required,min=4,max=55"`
	SenderContactType   string                 `json:"senderContactType" binding:"required,oneof=email phone"`
	SenderContactDetail string                 `json:"senderContactDetail" binding:"required"`
}

// Response structs
type PurchasedItemResponse struct {
	ProductID        string `json:"productId"`
	Name             string `json:"name"`
	Category         string `json:"category"`
	Qty              int32  `json:"qty"`
	Price            int    `json:"price"`
	SKU              string `json:"sku"`
	FileID           string `json:"fileId"`
	FileURI          string `json:"fileUri"`
	FileThumbnailURI string `json:"fileThumbnailUri"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type PaymentDetailResponse struct {
	BankAccountName   string  `json:"bankAccountName"`
	BankAccountHolder string  `json:"bankAccountHolder"`
	BankAccountNumber string  `json:"bankAccountNumber"`
	TotalPrice        float64 `json:"totalPrice"`
}

type CreatePurchaseResponse struct {
	PurchaseID     string                  `json:"purchaseId"`
	PurchasedItems []PurchasedItemResponse `json:"purchasedItems"`
	TotalPrice     float64                 `json:"totalPrice"`
	PaymentDetails []PaymentDetailResponse `json:"paymentDetails"`
}

type PaymentConfirmationRequest struct {
	FileIDs []string `json:"fileIds" binding:"required,min=1"`
}

// POST /v1/purchase
func (h *PurchaseHandler) CreatePurchase(c *gin.Context) {
	var req CreatePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
		return
	}

	// Validate sender contact detail
	if req.SenderContactType == "email" {
		if !utils.ValidateEmail(req.SenderContactDetail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}
	} else if req.SenderContactType == "phone" {
		if !utils.ValidatePhone(req.SenderContactDetail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
			return
		}
	}

	ctx := context.Background()

	var productSnapshots []repository.Product
	sellerSubtotals := make(map[int32]float64)
	var overallTotalPrice float64
	var total int32

	for _, item := range req.PurchasedItems {
		// Strictly reject invalid qty (<= 0)
		if item.Qty <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error"})
			return
		}

		productID, err := strconv.Atoi(item.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
			return
		}

		product, err := h.Queries.GetProductForUpdate(ctx, int32(productID))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Product with ID %s not found", item.ProductID)})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
			return
		}

		if product.Qty.Int32 < item.Qty {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Not enough stock for product %s. Available: %d, Requested: %d", product.Name.String, product.Qty.Int32, item.Qty)})
			return
		}

		productSnapshots = append(productSnapshots, product)
		price, _ := strconv.ParseFloat(product.Price.String, 64)
		itemTotal := price * float64(item.Qty)
		overallTotalPrice += itemTotal
		sellerSubtotals[product.UserID.Int32] += itemTotal
		total += int32(itemTotal)
	}

	purchase, err := h.Queries.CreatePurchase(ctx, repository.CreatePurchaseParams{
		SenderName:          sql.NullString{String: req.SenderName, Valid: true},
		SenderContactType:   sql.NullString{String: req.SenderContactType, Valid: true},
		SenderContactDetail: sql.NullString{String: req.SenderContactDetail, Valid: true},
		Total:               sql.NullInt32{Int32: total, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase record"})
		return
	}

	var purchasedItemsResponse []PurchasedItemResponse
	categoryCache := make(map[int32]string) // Cache for category names

	for i, snapshot := range productSnapshots {
		price, _ := strconv.ParseFloat(snapshot.Price.String, 64)
		itemTotal := price * float64(req.PurchasedItems[i].Qty)

		err := h.Queries.CreatePurchaseItem(ctx, repository.CreatePurchaseItemParams{
			PurchaseID: purchase.ID,
			ProductID:  snapshot.ProductID,
			Qty:        sql.NullInt32{Int32: req.PurchasedItems[i].Qty, Valid: true},
			Total:      sql.NullString{String: fmt.Sprintf("%.2f", itemTotal), Valid: true},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase item record"})
			return
		}

		// Get category name using the cache
		var categoryName string
		if name, ok := categoryCache[snapshot.Category.Int32]; ok {
			categoryName = name
		} else {
			catName, err := h.Queries.GetProductCategoryByID(ctx, snapshot.Category.Int32)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve category name"})
				return
			}
			categoryName = utils.NullStringToString(catName)
			categoryCache[snapshot.Category.Int32] = categoryName
		}

		// Build response snapshot
		fileURI, thumbnailURI, _ := utils.GetFileInfo(h.Queries, ctx, snapshot.FileID)
		priceInt, _ := strconv.Atoi(utils.NullStringToString(snapshot.Price))
		purchasedItemsResponse = append(purchasedItemsResponse, PurchasedItemResponse{
			ProductID:        fmt.Sprintf("%d", snapshot.ProductID),
			Name:             utils.NullStringToString(snapshot.Name),
			Category:         categoryName,
			Qty:              req.PurchasedItems[i].Qty, // The quantity bought
			Price:            priceInt,
			SKU:              utils.NullStringToString(snapshot.Sku),
			FileID:           utils.NullInt32ToString(snapshot.FileID),
			FileURI:          fileURI,
			FileThumbnailURI: thumbnailURI,
			CreatedAt:        snapshot.CreatedAt.Time.String(),
			UpdatedAt:        snapshot.UpdatedAt.Time.String(),
		})
	}

	var paymentDetailsResponse []PaymentDetailResponse
	for sellerID, subtotal := range sellerSubtotals {
		bankDetails, err := h.Queries.GetSellerBankDetailsByUserID(ctx, sellerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seller bank details"})
			return
		}
		paymentDetailsResponse = append(paymentDetailsResponse, PaymentDetailResponse{
			BankAccountName:   utils.NullStringToString(bankDetails.BankAccountName),
			BankAccountHolder: utils.NullStringToString(bankDetails.BankAccountHolder),
			BankAccountNumber: utils.NullStringToString(bankDetails.BankAccountNumber),
			TotalPrice:        subtotal,
		})
	}

	c.JSON(http.StatusCreated, CreatePurchaseResponse{
		PurchaseID:     fmt.Sprintf("%d", purchase.ID),
		PurchasedItems: purchasedItemsResponse,
		TotalPrice:     overallTotalPrice,
		PaymentDetails: paymentDetailsResponse,
	})
}

func (h *PurchaseHandler) ConfirmPayment(c *gin.Context) {
	purchaseIDStr := c.Param("purchaseId")
	purchaseID, err := strconv.Atoi(purchaseIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid purchase ID format"})
		return
	}

	var req PaymentConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Begin transaction
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	// Get purchase details
	purchase, err := qtx.GetPurchaseByID(ctx, int32(purchaseID))
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Purchase not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve purchase"})
		return
	}

	// Check if already paid
	if purchase.IsPaid.Valid && purchase.IsPaid.Bool {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Purchase already paid"})
		return
	}

	// Get purchase items with seller information
	purchaseItems, err := qtx.GetPurchaseItemsByPurchaseID(ctx, int32(purchaseID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve purchase items"})
		return
	}

	// Group items by seller
	sellerItems := make(map[int32][]repository.GetPurchaseItemsByPurchaseIDRow)
	for _, item := range purchaseItems {
		if item.UserID.Valid {
			sellerItems[item.UserID.Int32] = append(sellerItems[item.UserID.Int32], item)
		}
	}

	// Validate file IDs count matches number of sellers
	if len(req.FileIDs) != len(sellerItems) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Number of file IDs (%d) must match number of sellers (%d)", len(req.FileIDs), len(sellerItems))})
		return
	}

	// Validate all file IDs exist
	for _, fileIDStr := range req.FileIDs {
		fileID, err := strconv.Atoi(fileIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID format"})
			return
		}

		file, err := qtx.GetFileByID(ctx, int32(fileID))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File with ID %s not found", fileIDStr)})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate file"})
			return
		}
		_ = file
	}

	// Create payment details for each seller
	fileIndex := 0
	for sellerID, items := range sellerItems {
		fileID, err := strconv.Atoi(req.FileIDs[fileIndex])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID format"})
			return
		}

		err = qtx.CreatePaymentDetail(ctx, repository.CreatePaymentDetailParams{
			PurchaseID: int32(purchaseID),
			UserID:     sql.NullInt32{Int32: sellerID, Valid: true},
			FileID:     sql.NullInt32{Int32: int32(fileID), Valid: true},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment detail"})
			return
		}

		// Update product quantities (decrease even if it goes negative)
		for _, item := range items {
			err = qtx.UpdateProductQuantity(ctx, repository.UpdateProductQuantityParams{
				ProductID: item.ProductID,
				Qty:       item.Qty,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product quantity"})
				return
			}
		}

		fileIndex++
	}

	// Update purchase status to paid
	err = qtx.UpdatePurchasePaymentStatus(ctx, int32(purchaseID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Payment confirmed successfully"})
}
