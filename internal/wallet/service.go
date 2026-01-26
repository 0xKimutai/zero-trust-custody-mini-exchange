package wallet

import (
	"database/sql"
	"time"
)

type Blockchain struct {
	ID                    string
	Network               string
	ConfirmationThreshold int
}

type Asset struct {
	ID              string
	BlockchainID    string
	Symbol          string
	Decimals        int
	IsToken         bool
	ContractAddress sql.NullString
}

type Wallet struct {
	ID        string
	Type      string // 'hot', 'cold'
	AssetID   string
	Address   string
	Balance   float64 // Using float64 for simplicity in struct, but DB uses NUMERIC. Be careful with precision.
	CreatedAt time.Time
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// CreateBlockchain adds a new supported blockchain
func (s *Service) CreateBlockchain(id, network string, confirmations int) error {
	_, err := s.db.Exec(`INSERT INTO blockchains (id, network, confirmation_threshold) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`,
		id, network, confirmations)
	return err
}

// CreateAsset adds a new asset support
func (s *Service) CreateAsset(id, chainID, symbol string, decimals int, isToken bool, contractAddr string) error {
	var cAddr sql.NullString
	if contractAddr != "" {
		cAddr = sql.NullString{String: contractAddr, Valid: true}
	}
	_, err := s.db.Exec(`INSERT INTO assets (id, blockchain_id, symbol, decimals, is_token, contract_address) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING`,
		id, chainID, symbol, decimals, isToken, cAddr)
	return err
}

// CreateWallet adds a hot/cold wallet for the exchange
func (s *Service) CreateWallet(wType, assetID, address string) error {
	_, err := s.db.Exec(`INSERT INTO wallets (type, asset_id, address) VALUES ($1, $2, $3)`,
		wType, assetID, address)
	return err
}

// GetAsset retrieves asset details
func (s *Service) GetAsset(id string) (*Asset, error) {
	var a Asset
	err := s.db.QueryRow(`SELECT id, blockchain_id, symbol, decimals, is_token, contract_address FROM assets WHERE id = $1`, id).
		Scan(&a.ID, &a.BlockchainID, &a.Symbol, &a.Decimals, &a.IsToken, &a.ContractAddress)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
