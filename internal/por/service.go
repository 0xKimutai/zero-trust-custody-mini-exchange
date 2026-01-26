package por

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type Leaf struct {
	UserID  string
	Balance float64
	Hash    string
}

type MerkleNode struct {
	Hash  string
	Left  *MerkleNode
	Right *MerkleNode
}

// GenerateSnapshot creates a PoR snapshot for an asset
func (s *Service) GenerateSnapshot(ctx context.Context, assetID string) (string, error) {
	// 1. Snapshot Balances (Copy to a temporary structure or just fetch)
	// We fetch all user liabilities
	rows, err := s.db.QueryContext(ctx, `
		SELECT user_id, balance 
		FROM ledger_accounts 
		WHERE asset_id = $1 AND type = 'liability' AND balance > 0
		ORDER BY user_id
	`, assetID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// type Leaf is now package level
	var leaves []Leaf
	var totalLiabilities float64

	for rows.Next() {
		var l Leaf
		if err := rows.Scan(&l.UserID, &l.Balance); err != nil {
			continue
		}
		totalLiabilities += l.Balance
		
		// Hash = SHA256(UserID + Balance)
		// Real impl should include nonce
		data := fmt.Sprintf("%s:%.8f", l.UserID, l.Balance)
		hash := sha256.Sum256([]byte(data))
		l.Hash = hex.EncodeToString(hash[:])
		leaves = append(leaves, l)
	}

	if len(leaves) == 0 {
		return "", fmt.Errorf("no liabilities found for %s", assetID)
	}

	// 2. Build Tree
	rootHash := buildMerkleTree(leaves)

	// 3. Store Snapshot
	id := uuid.New().String()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO por_snapshots (id, asset_id, merkle_root, total_liabilities)
		VALUES ($1, $2, $3, $4)
	`, id, assetID, rootHash, totalLiabilities)

	return rootHash, err
}

func buildMerkleTree(leaves []Leaf) string {
	var hashes []string
	for _, l := range leaves {
		hashes = append(hashes, l.Hash)
	}
	
	if len(hashes) == 0 {
		return ""
	}

	return computeRoot(hashes)
}

func computeRoot(hashes []string) string {
	if len(hashes) == 1 {
		return hashes[0]
	}

	var parsed []string
	// If odd, duplicate last
	if len(hashes)%2 != 0 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}

	for i := 0; i < len(hashes); i += 2 {
		left := hashes[i]
		right := hashes[i+1]
		// Concatenate and Hash
		combined := left + right
		sum := sha256.Sum256([]byte(combined))
		parsed = append(parsed, hex.EncodeToString(sum[:]))
	}

	return computeRoot(parsed)
}
