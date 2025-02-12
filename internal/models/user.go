package models

import "time"

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Coins        int
	CreatedAt    time.Time
}
