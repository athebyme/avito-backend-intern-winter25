package handlers

import (
	"avito-backend-intern-winter25/internal/middleware"
	"avito-backend-intern-winter25/internal/models/http/request"
	"avito-backend-intern-winter25/internal/models/http/response"
	"avito-backend-intern-winter25/internal/services"
	"avito-backend-intern-winter25/internal/services/jwt"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct {
	userService        *services.UserService
	merchService       *services.MerchService
	transactionService *services.TransactionService
	logger             zap.Logger
}

func NewHandler(
	userService *services.UserService,
	merchService *services.MerchService,
	transactionService *services.TransactionService,
	writer zap.Logger,
) *Handler {
	return &Handler{
		userService:        userService,
		merchService:       merchService,
		transactionService: transactionService,
		logger:             writer,
	}
}

func (h *Handler) SetupRoutes(r *gin.Engine, jwtService *jwt.Service) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Recovered from panic in SetupRoutes:", err)
		}
	}()

	api := r.Group("/api")
	{
		api.POST("/auth", h.Auth)

		secured := api.Group("/")
		secured.Use(middleware.AuthMiddleware(jwtService))
		{
			secured.GET("/info", h.GetInfo)
			secured.GET("/balance", h.Balance)
			secured.POST("/sendCoin", h.SendCoin)
			secured.GET("/merch/list", h.ListMerch)
			secured.GET("/buy/:item", h.BuyItem)
		}
	}

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/health", h.Health)
	h.logger.Info("Routes setup complete.")
}

func (h *Handler) Auth(c *gin.Context) {
	var req request.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Errors: "invalid request body"})
		return
	}

	user, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Errors: err.Error()})
		return
	}

	token, err := h.userService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Errors: "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) GetInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)

	purchases, err := h.merchService.GetPurchasesByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Errors: err.Error()})
		return
	}

	sentTx, err := h.transactionService.GetSentTransactions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Errors: err.Error()})
		return
	}
	receivedTx, err := h.transactionService.GetReceivedTransactions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Errors: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"purchases":         purchases,
		"sent_transactions": sentTx,
		"received_tx":       receivedTx,
	})
}

func (h *Handler) SendCoin(c *gin.Context) {
	var req request.SendCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Errors: "invalid request format"})
		return
	}

	fromUserID := middleware.GetUserID(c)
	toUser, err := h.userService.GetUserByUsername(c, req.ToUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Errors: "recipient user not found"})
		return
	}

	err = h.transactionService.TransferCoins(c, fromUserID, toUser.ID, req.Amount)
	if err != nil {
		switch err {
		case services.ErrInvalidAmount:
			c.JSON(http.StatusBadRequest, response.ErrorResponse{Errors: "invalid amount"})
		case services.ErrLackOfFundsOnAccount:
			c.JSON(http.StatusBadRequest, response.ErrorResponse{Errors: "insufficient funds"})
		default:
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Errors: "failed to transfer coins"})
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) BuyItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	itemName := c.Param("item")

	err := h.merchService.PurchaseItem(c, userID, itemName)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientCoins):
			c.JSON(http.StatusBadRequest, response.ErrorResponse{Errors: "insufficient coins"})
		default:
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Errors: "failed to purchase item"})
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) ListMerch(c *gin.Context) {
	merch, err := h.merchService.GetAllAvailableMerch(c)

	resp := make([]*response.MerchResponse, len(merch))
	for i, v := range merch {
		resp[i] = response.MerchResponseFromModel(v)
	}
	if err != nil {
		switch {
		default:
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Errors: "failed to purchase item"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Balance(c *gin.Context) {
	userID := middleware.GetUserID(c)
	balance, err := h.userService.GetUserBalance(c, userID)
	if err != nil {
		switch {
		default:
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Errors: "failed to purchase item"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}
