package services

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/services/jwt"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserNotFound    = errors.New("user not found")
)

type UserService struct {
	userRepo    storage.UserRepository
	jwtService  jwt.JWT
	bcryptCost  int
	redisClient RedisClient
	cacheTTL    time.Duration
}

func NewUserService(userRepo storage.UserRepository, jwtService jwt.JWT, redisClient RedisClient) *UserService {
	return &UserService{
		userRepo:    userRepo,
		jwtService:  jwtService,
		bcryptCost:  4,
		redisClient: redisClient,
		cacheTTL:    30 * time.Minute,
	}
}

func (s *UserService) Login(ctx context.Context, username, password string) (user *domain.User, err error) {
	cachedUser, err := s.getCachedUser(ctx, username)
	if err == nil && cachedUser != nil {
		if err = bcrypt.CompareHashAndPassword([]byte(cachedUser.PasswordHash), []byte(password)); err == nil {
			return cachedUser, nil
		}
	}

	user, err = s.userRepo.FindByUsername(ctx, username)

	if errors.Is(err, storage.ErrUserNotFound) {
		tx, txErr := s.userRepo.BeginTx(ctx)
		if txErr != nil {
			return nil, txErr
		}

		defer func() {
			if p := recover(); p != nil {
				if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
					log.Printf("rollback error: %v", rbErr)
				}
				panic(p)
			} else if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
					log.Printf("rollback error: %v", rbErr)
				}
			} else {
				if cmErr := tx.Commit(); cmErr != nil {
					log.Printf("commit tx error: %v", cmErr)
					user = nil
					err = cmErr
				}
			}
		}()

		user, err = s.userRepo.FindByUsername(ctx, username)
		if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
			return nil, err
		}

		if user == nil || errors.Is(err, storage.ErrUserNotFound) {
			var hashedPassword []byte
			if hashedPassword, err = bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost); err != nil {
				return nil, err
			}
			user = &domain.User{
				Username:     username,
				PasswordHash: string(hashedPassword),
				Coins:        1000,
			}
			if err = s.userRepo.Create(ctx, tx, user); err != nil {
				return nil, err
			}
			if cacheErr := s.cacheUser(ctx, user); cacheErr != nil {
				log.Printf("Failed to cache new user: %v", cacheErr)
			}
			return user, nil
		}
	} else if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		err = ErrInvalidPassword
		return nil, err
	}

	if cacheErr := s.cacheUser(ctx, user); cacheErr != nil {
		log.Printf("Failed to cache user: %v", cacheErr)
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

func (s *UserService) getCachedUser(ctx context.Context, username string) (*domain.User, error) {
	val, err := s.redisClient.Get(ctx, "user:"+username).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		log.Printf("Redis error: %v", err)
		return nil, err
	}

	var user domain.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		log.Printf("Failed to unmarshal user from cache: %v", err)
		return nil, err
	}

	return &user, nil
}

func (s *UserService) cacheUser(ctx context.Context, user *domain.User) error {
	userData, err := json.Marshal(user)
	if err != nil {
		log.Printf("Failed to marshal user for cache: %v", err)
		return err
	}

	return s.redisClient.Set(ctx, "user:"+user.Username, userData, s.cacheTTL).Err()
}
