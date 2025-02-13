package handlers

import "avito-backend-intern-winter25/internal/services"

type TransactionHandler struct {
	transactionService services.TransactionService
	userService        services.UserService
}
