package handlers

import (
	"avito-backend-intern-winter25/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type AuthHandler struct {
	authService services.AuthService
}

func (h *Handler) Auth(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request"})
		return
	}

	// Поиск или создание пользователя
	user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Генерация токена
	token, err := jwt.GenerateToken(
		user.ID,
		user.Username,
		h.cfg.JWT.SecretKey,
		time.Hour*time.Duration(h.cfg.JWT.ExpirationHours),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: token})
}
