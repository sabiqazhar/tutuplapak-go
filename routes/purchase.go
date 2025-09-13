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
	Price            string `json:"price"`
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

// POST /v1/purchase
func (h *PurchaseHandler) CreatePurchase(c *gin.Context) {
	var req CreatePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)
	ctx := context.Background()

	var productSnapshots []repository.Product
	sellerSubtotals := make(map[int32]float64)
	var overallTotalPrice float64
	var total int32

	for _, item := range req.PurchasedItems {
		productID, err := strconv.Atoi(item.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
			return
		}

		product, err := qtx.GetProductForUpdate(ctx, int32(productID))
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

	purchase, err := qtx.CreatePurchase(ctx, repository.CreatePurchaseParams{
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

		err := qtx.CreatePurchaseItem(ctx, repository.CreatePurchaseItemParams{
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
		purchasedItemsResponse = append(purchasedItemsResponse, PurchasedItemResponse{
			ProductID:        fmt.Sprintf("%d", snapshot.ProductID),
			Name:             utils.NullStringToString(snapshot.Name),
			Category:         categoryName,
			Qty:              req.PurchasedItems[i].Qty, // The quantity bought
			Price:            utils.NullStringToString(snapshot.Price),
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
		bankDetails, err := qtx.GetSellerBankDetailsByUserID(ctx, sellerID)
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

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, CreatePurchaseResponse{
		PurchaseID:     fmt.Sprintf("%d", purchase.ID),
		PurchasedItems: purchasedItemsResponse,
		TotalPrice:     overallTotalPrice,
		PaymentDetails: paymentDetailsResponse,
	})
}
