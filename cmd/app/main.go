package main

import (
	"avito-backend-intern-winter25/config"
	"avito-backend-intern-winter25/internal/services"
	"avito-backend-intern-winter25/internal/storage/postgres"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

const (
	configLocation = "config/config.yaml"
	migrationDir   = "file:./migrations"
)

func main() {
	cfg := &config.PostgresConfig{}

	rawCfg, err := cfg.Init(configLocation)
	if err != nil {
		log.Fatalf("error initilizating config : %v", err)
	}

	cfg, ok := rawCfg.(*config.PostgresConfig)
	if !ok {
		log.Fatalf("unexpected configuration file.")
	}

	connectionString := cfg.GetConnectionString()
	fmt.Println("Connection String:", connectionString)

	m, err := migrate.New(migrationDir, connectionString)
	if err != nil {
		log.Fatalf("Ошибка инициализации миграций: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}
	log.Println("Миграции успешно применены!")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := postgres.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	login, err := userService.Login("rizz_god", "hello world !")
	login2, err := userService.Login("sigma", "hello world !")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Hello, %s ! %d  %d %s", login.Username, login.ID, login.Coins, login.CreatedAt)

	transactionRepo := postgres.NewTransactionRepository(db)
	transactionService := services.NewTransactionService(
		db,
		userRepo,
		transactionRepo,
	)

	err = transactionService.TransferCoins(login.ID, login2.ID, 1000)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("OK")

	purchaseRepo := postgres.NewPurchaseRepository(db)
	merchRepo := postgres.NewMerchRepository(db)
	merchService := services.NewMerchService(
		merchRepo,
		purchaseRepo,
		userRepo,
		db,
	)

	merchService.PurchaseItem(login2.ID, "")
}
