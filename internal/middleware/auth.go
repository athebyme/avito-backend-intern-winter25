package middleware

import (
	jwtservice "avito-backend-intern-winter25/internal/services/jwt"
	"github.com/gin-gonic/gin"
	"strings"
)

const (
	userIDKey    = "userID"
	bearerSchema = "Bearer "
)

func AuthMiddleware(jwtService *jwtservice.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"errors": "authorization header is required"})
			return
		}

		if !strings.HasPrefix(authHeader, bearerSchema) {
			c.AbortWithStatusJSON(401, gin.H{"errors": "invalid authorization header format"})
			return
		}

		token := strings.TrimPrefix(authHeader, bearerSchema)
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"errors": "invalid token"})
			return
		}

		c.Set(userIDKey, claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) int64 {
	userID, _ := c.Get(userIDKey)
	return userID.(int64)
}
