package main

import (
	"log"

	"tutuplapak-go/middleware"
	"tutuplapak-go/provider"
	"tutuplapak-go/repository"
	"tutuplapak-go/routes"
	"tutuplapak-go/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	dataSourceName := "host=localhost user=user password=password dbname=tutuplapak port=5433 sslmode=disable TimeZone=Asia/Jakarta"

	// Init DB
	db := provider.InitDB(dataSourceName)

	// Init sqlc Queries
	queries := repository.New(db)

	// Init Handlers
	authHandler := routes.NewAuthHandler(queries)
	profileHandler := routes.NewProfileHandler(queries)
	fileHandler := routes.NewFileHandler(queries)
	productHandler := routes.NewProductHandler(queries)
	purchaseHandler := routes.NewPurchaseHandler(queries, db)

	// Start token cleanup routine
	utils.GlobalTokenStore.StartCleanupRoutine()

	// Setup Gin
	r := gin.Default()

	// Serve static files (for uploaded files)
	r.Static("/uploads", "./uploads")

	// V1 API Routes according to requirement
	v1 := r.Group("/v1")
	{
		// Public routes (no authentication required)
		v1.POST("/login/email", authHandler.LoginEmail)
		v1.POST("/login/phone", authHandler.LoginPhone)
		v1.POST("/register/email", authHandler.RegisterEmail)
		v1.POST("/register/phone", authHandler.RegisterPhone)
		v1.POST("/file", fileHandler.UploadFile)
		v1.GET("/product", productHandler.GetProducts)

		// Protected routes (require authentication)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User profile routes
			protected.GET("/user", profileHandler.GetProfile)
			protected.PUT("/user", profileHandler.UpdateProfile)
			protected.POST("/user/link/phone", profileHandler.LinkPhone)
			protected.POST("/user/link/email", profileHandler.LinkEmail)
			protected.POST("/product", productHandler.CreateProduct)
			protected.PUT("/product", productHandler.UpdateProduct)
			protected.DELETE("/product", productHandler.DeleteProduct)
			protected.POST("/purchase", purchaseHandler.CreatePurchase)
		}
	}

	// Run server
	port := ":8080"
	log.Printf("🚀 Server running on http://localhost%s", port)
	log.Fatal(r.Run(port))
}
