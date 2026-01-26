package seed

import (
	"log"
	"mini-exchange/internal/wallet"
)

func Seed(ws *wallet.Service) {
	// Chains
	if err := ws.CreateBlockchain("bitcoin", "mainnet", 2); err != nil {
		log.Printf("Seed error: %v", err)
	}
	if err := ws.CreateBlockchain("ethereum", "mainnet", 12); err != nil {
		log.Printf("Seed error: %v", err)
	}

	// Assets
	if err := ws.CreateAsset("BTC", "bitcoin", "BTC", 8, false, ""); err != nil {
		log.Printf("Seed error: %v", err)
	}
	if err := ws.CreateAsset("ETH", "ethereum", "ETH", 18, false, ""); err != nil {
		log.Printf("Seed error: %v", err)
	}
	if err := ws.CreateAsset("USDT", "ethereum", "USDT", 6, true, "0xdac17f958d2ee523a2206206994597c13d831ec7"); err != nil {
		log.Printf("Seed error: %v", err)
	}

	// Wallets (Hot) - In real life these would be real addresses
	if err := ws.CreateWallet("hot", "BTC", "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"); err != nil {
		log.Printf("Seed error: %v", err)
	}
	if err := ws.CreateWallet("hot", "ETH", "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"); err != nil {
		log.Printf("Seed error: %v", err)
	}
}
