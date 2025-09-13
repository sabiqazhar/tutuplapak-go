// cmd/main.go
package main

import (
	"log"

	"tutuplapak-go/provider"
	"tutuplapak-go/repository" // sqlc generated
	"tutuplapak-go/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	dataSourceName := "host=localhost user=postgres password=yourpassword dbname=yourdbname port=5432 sslmode=disable TimeZone=Asia/Jakarta"

	// Init DB
	db := provider.InitDB(dataSourceName)

	// Init sqlc Queries
	queries := repository.New(db)

	// Init Auth Handler
	authHandler := routes.NewAuthHandler(queries)

	// Setup Gin
	r := gin.Default()

	// Routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Jalankan server
	port := ":8080"
	log.Printf("🚀 Server running on http://localhost%s", port)
	log.Fatal(r.Run(port))
}