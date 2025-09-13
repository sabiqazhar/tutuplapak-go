// middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"tutuplapak-go/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates Bearer tokens and sets user_id in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Get user ID from token store
		userID, exists := utils.GlobalTokenStore.GetUserID(token)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Set user ID in context for use in handlers
		c.Set("user_id", userID)
		c.Next()
	}
}
