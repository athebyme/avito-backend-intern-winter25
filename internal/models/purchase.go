package models

import "time"

type Purchase struct {
	ID           int64
	UserID       int64
	Item         string
	Price        int
	PurchaseDate time.Time
}
