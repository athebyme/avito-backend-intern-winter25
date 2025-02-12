package models

import "time"

type CoinTransaction struct {
	ID         int64
	FromUserID int64
	ToUserID   int64
	Amount     int
	CreatedAt  time.Time
}
