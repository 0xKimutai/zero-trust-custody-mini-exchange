package withdrawal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mini-exchange/internal/ledger"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db     *sql.DB
	ledger *ledger.Service
}

func NewService(db *sql.DB, l *ledger.Service) *Service {
	return &Service{
		db:     db,
		ledger: l,
	}
}

// RequestWithdrawal initiates a user withdrawal
func (s *Service) RequestWithdrawal(ctx context.Context, userID, assetID, toAddress string, amount float64) (string, error) {
	if amount <= 0 {
		return "", errors.New("amount must be positive")
	}

	// 1. Check Balance (Optimistic check, Ledger will enforce strictly)
	bal, err := s.ledger.GetUserBalance(userID, assetID)
	if err != nil {
		return "", err
	}
	if bal < amount {
		return "", errors.New("insufficient balance")
	}

	// 2. Risk Checks (Simulated)
	// Example: Max withdrawal limit
	if amount > 1000 { // e.g. 1000 BTC is too much
		return "", errors.New("risk limit exceeded")
	}

	// 3. Start DB Tx for Atomic Operations
	// Note: We need to coordinate with Ledger.
	// Since Ledger.PostTransaction manages its own TX, we do strict ordering:
	// A. Create Withdrawal Record (status=processing)
	// B. Execute Ledger Hold (Debit User, Credit Pending)
	// If B fails, we fail Request. (Rollback via deleting withdrawal or setting status failed)
	
	// Better approach for atomicity:
	// Create Withdrawal with status 'pending_ledger'.
	// Then execute Ledger.
	// Then update Withdrawal to 'requested'.
	
	withdrawalID := uuid.New().String()
	_, err = s.db.Exec(`
		INSERT INTO withdrawals (id, user_id, asset_id, amount, to_address, status)
		VALUES ($1, $2, $3, $4, $5, 'requesting')
	`, withdrawalID, userID, assetID, amount, toAddress)
	if err != nil {
		return "", err
	}

	// 4. Ledger Hold
	// Debit User Liability
	// Credit "Withdrawal Hold" Liability (System Account)
	
	userAcc, err := s.ledger.GetOrCreateAccount(userID, assetID, "liability")
	if err != nil {
		return "", err
	}
	
	// System "Hold" Account (Liability type ?)
	// Yes, we still owe *someone* (the user, but locked).
	// Or we credit a "Pending Withdrawal" account which is a liability.
	holdAcc, err := s.ledger.GetOrCreateAccount("", assetID, "liability_hold")
	if err != nil {
		return "", err
	}

	desc := fmt.Sprintf("Withdrawal Req %s", withdrawalID)
	req := ledger.TransactionRequest{
		Description: desc,
		ReferenceID: withdrawalID,
		Entries: []ledger.Entry{
			{AccountID: userAcc.ID, Direction: "debit", Amount: amount},    // Reduce Available
			{AccountID: holdAcc.ID, Direction: "credit", Amount: amount},   // Increase On-Hold
		},
	}

	if err := s.ledger.PostTransaction(ctx, req); err != nil {
		// Cleanup
		s.db.Exec(`UPDATE withdrawals SET status = 'failed' WHERE id = $1`, withdrawalID)
		return "", fmt.Errorf("ledger lock failed: %w", err)
	}

	// Success
	_, err = s.db.Exec(`UPDATE withdrawals SET status = 'requested' WHERE id = $1`, withdrawalID)
	return withdrawalID, err
}

// ProcessBatch simulates picking up requests and broadcasting them
// This would be run by a background worker or cron, or triggered via Admin API
func (s *Service) ProcessBatch(ctx context.Context) error {
	// 1. Get requested withdrawals
	rows, err := s.db.QueryContext(ctx, `SELECT id, asset_id, amount, to_address FROM withdrawals WHERE status = 'requested' LIMIT 10`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, assetID, toAddr string
		var amount float64
		if err := rows.Scan(&id, &assetID, &amount, &toAddr); err != nil {
			continue
		}

		// 2. Simulate Broadcast
		txHash := fmt.Sprintf("tx-withdraw-%s", uuid.New().String()[:8])
		
		// 3. Finalize Ledger
		// Debit "Hold" Liability
		// Credit "System Hot Wallet" Asset (Reducing Asset)
		
		holdAcc, err := s.ledger.GetOrCreateAccount("", assetID, "liability_hold")
		if err != nil {
			// Log error
			continue
		}
		hotWalletAcc, err := s.ledger.GetOrCreateAccount("", assetID, "asset")
		if err != nil {
			continue
		}

		req := ledger.TransactionRequest{
			Description: fmt.Sprintf("Withdrawal Complete %s", id),
			ReferenceID: id,
			Entries: []ledger.Entry{
				{AccountID: holdAcc.ID, Direction: "debit", Amount: amount},       // Remove Hold
				{AccountID: hotWalletAcc.ID, Direction: "credit", Amount: amount}, // Reduce System Asset
			},
		}

		if err := s.ledger.PostTransaction(ctx, req); err != nil {
			// Log error, maybe retry
			continue
		}

		// 4. Update Status
		s.db.Exec(`UPDATE withdrawals SET status = 'completed', tx_hash = $1 WHERE id = $2`, txHash, id)
	}
	return nil
}

// GetHistory returns user's withdrawals
func (s *Service) GetHistory(userID string) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`SELECT id, asset_id, amount, status, tx_hash, created_at FROM withdrawals WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id, assetID, status string
		var txHash sql.NullString
		var amount float64
		var createdAt time.Time
		if err := rows.Scan(&id, &assetID, &amount, &status, &txHash, &createdAt); err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id": id, "asset": assetID, "amount": amount, "status": status, "tx_hash": txHash.String, "created_at": createdAt,
		})
	}
	return result, nil
}
