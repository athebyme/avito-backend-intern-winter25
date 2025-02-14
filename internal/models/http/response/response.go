package response

import "avito-backend-intern-winter25/internal/models/domain"

type AuthResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}

type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Item      `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}

type Item struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []ReceivedTransaction `json:"received"`
	Sent     []SentTransaction     `json:"sent"`
}

type ReceivedTransaction struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type SentTransaction struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type MerchResponse struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func MerchResponseFromModel(m *domain.Merch) *MerchResponse {
	if m == nil {
		return nil
	}
	return &MerchResponse{
		Name:  m.Name,
		Price: m.Price,
	}
}
