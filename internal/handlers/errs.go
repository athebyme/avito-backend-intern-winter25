package handlers

import (
	"avito-backend-intern-winter25/pkg/apperrors"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			status, response := apperrors.ToResponse(err)
			c.JSON(status, response)
		}
	}
}
