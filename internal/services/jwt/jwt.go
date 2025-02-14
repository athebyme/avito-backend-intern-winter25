package jwt

type JWT interface {
	GenerateToken(userID int64, username string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}
