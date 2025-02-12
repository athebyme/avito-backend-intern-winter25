package services

import (
	"avito-backend-intern-winter25/internal/models"
	"avito-backend-intern-winter25/internal/storage"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

type UserService struct {
	userRepo storage.UserRepository
}

func NewUserService(userRepo storage.UserRepository) *UserService {
	return &UserService{userRepo}
}

func (s *UserService) Login(username, password string) (*models.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if errors.Is(err, storage.ErrUserNotFound) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user = &models.User{
			Username:     username,
			PasswordHash: string(hashedPassword),
			Coins:        1000,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
		return user, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}
