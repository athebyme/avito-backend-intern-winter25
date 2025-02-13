package services

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
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

func (s *UserService) Login(ctx context.Context, username, password string) (*domain.User, error) {
	tx, err := s.userRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("rollback error: %v", err)
		}
	}()

	user, err := s.userRepo.FindByUsername(ctx, username)
	if errors.Is(err, storage.ErrUserNotFound) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user = &domain.User{
			Username:     username,
			PasswordHash: string(hashedPassword),
			Coins:        1000,
		}
		if err := s.userRepo.Create(ctx, tx, user); err != nil {
			return nil, err
		}
		return user, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}
