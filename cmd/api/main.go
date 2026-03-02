package main

import (
	"log"
	"os"

	"mini-exchange/config"
	"mini-exchange/db"
	"mini-exchange/internal/api"
	"mini-exchange/internal/auth"
	"mini-exchange/internal/deposit"
	"mini-exchange/internal/ledger"
	"mini-exchange/internal/por"
	"mini-exchange/internal/seed"
	"mini-exchange/internal/wallet"
	"mini-exchange/internal/withdrawal"
)

func main() {
	// Load Config (and .env)
	cfg := config.LoadConfig()

	// Initialize Database
	err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	// Initialize Services
	authService := auth.NewService(db.GetDB(), cfg.JWTSecret)
	walletService := wallet.NewService(db.GetDB())
	ledgerService := ledger.NewService(db.GetDB())
	depositService := deposit.NewService(db.GetDB(), ledgerService)
	withdrawalService := withdrawal.NewService(db.GetDB(), ledgerService)
	porService := por.NewService(db.GetDB())

	// Auto-Migrate
	schema, err := os.ReadFile("db/schema.sql")
	if err == nil {
		_, err = db.GetDB().Exec(string(schema))
		if err != nil {
			log.Printf("Migration warning: %v", err)
		} else {
			log.Println("Database Schema Applied")
		}
	}

	// Run Seeds (Must be after migration)
	seed.Seed(walletService)

	// Initialize Handlers
	depositHandler := deposit.NewHandler(depositService)
	withdrawalHandler := withdrawal.NewHandler(withdrawalService)
	porHandler := por.NewHandler(porService)

	// Setup Router
	r := api.SetupRouter(authService, depositHandler, withdrawalHandler, porHandler)

	port := config.GetEnv("PORT", "8080")
	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
