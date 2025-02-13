package middleware

import (
	"avito-backend-intern-winter25/config"
	"avito-backend-intern-winter25/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func Auth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header is required"})
			return
		}

		claims, err := jwt.ParseToken(tokenString, cfg.SecretKey)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
