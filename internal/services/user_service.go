package services

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/services/jwt"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserNotFound    = errors.New("user not found")
)

type UserService struct {
	userRepo   storage.UserRepository
	jwtService jwt.JWT
}

func NewUserService(userRepo storage.UserRepository, jwtService jwt.JWT) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *UserService) Login(ctx context.Context, username, password string) (*domain.User, error) {
	tx, err := s.userRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
				log.Printf("rollback error: %v", err)
			}
			panic(p)
		} else if err != nil {
			if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
				log.Printf("rollback error: %v", err)
			}
		} else {
			if err := tx.Commit(); err != nil {
				log.Printf("commit tx error: %v", err)
				return
			}
		}
	}()

	user, err := s.userRepo.FindByUsername(ctx, username)
	if errors.Is(err, storage.ErrUserNotFound) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user = &domain.User{
			Username:     username,
			PasswordHash: string(hashedPassword),
			Coins:        1000,
		}
		if err := s.userRepo.Create(ctx, tx, user); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return user, nil
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GenerateToken(user *domain.User) (string, error) {
	return s.jwtService.GenerateToken(user.ID, user.Username)
}

func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUserCoins(ctx context.Context, userID int64, coins int) error {
	tx, err := s.userRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Coins = coins
	return s.userRepo.Update(ctx, tx, user)
}

func (s *UserService) ValidateUserBalance(ctx context.Context, userID int64, amount int) (bool, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.Coins >= amount, nil
}

func (s *UserService) GetUserBalance(ctx context.Context, userID int64) (int, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.Coins, nil
}
