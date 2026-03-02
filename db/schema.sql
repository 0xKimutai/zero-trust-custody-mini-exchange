-- Users & Auth
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'user', -- user, admin
    status TEXT DEFAULT 'active', -- active, frozen
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Assets & Chains
CREATE TABLE IF NOT EXISTS blockchains (
    id TEXT PRIMARY KEY, -- e.g., 'bitcoin', 'ethereum'
    network TEXT NOT NULL, -- 'mainnet', 'testnet'
    confirmation_threshold INT DEFAULT 1
);

CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY, -- 'BTC', 'ETH', 'USDT'
    blockchain_id TEXT REFERENCES blockchains(id),
    symbol TEXT NOT NULL,
    decimals INT NOT NULL,
    is_token BOOLEAN DEFAULT FALSE,
    contract_address TEXT -- NULL for native assets
);

-- Wallets
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL, -- 'hot', 'cold'
    asset_id TEXT REFERENCES assets(id),
    address TEXT NOT NULL,
    balance NUMERIC NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(address, asset_id)
);

-- Ledger (Double Entry)
CREATE TABLE IF NOT EXISTS ledger_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id), -- NULL for system accounts
    asset_id TEXT REFERENCES assets(id),
    type TEXT NOT NULL, -- 'liability' (user balance), 'asset' (exchange wallet tracking)
    balance NUMERIC NOT NULL DEFAULT 0,
    CONSTRAINT unique_user_asset UNIQUE (user_id, asset_id, type)
);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL, -- Grouping ID for double-entry
    account_id UUID REFERENCES ledger_accounts(id),
    direction TEXT NOT NULL, -- 'debit', 'credit'
    amount NUMERIC NOT NULL CHECK (amount > 0),
    balance_after NUMERIC NOT NULL,
    description TEXT,
    reference_id UUID, -- Link to deposit/withdrawal
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Deposits
CREATE TABLE IF NOT EXISTS deposit_addresses (
    address TEXT PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    asset_id TEXT REFERENCES assets(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    asset_id TEXT REFERENCES assets(id),
    tx_hash TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    confirmations INT DEFAULT 0,
    status TEXT NOT NULL, -- 'pending', 'confirmed', 'credited'
    ledger_tx_id UUID, -- Link to ledger transaction
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tx_hash, asset_id) -- Prevent replay
);

-- Withdrawals
CREATE TABLE IF NOT EXISTS withdrawals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    asset_id TEXT REFERENCES assets(id),
    amount NUMERIC NOT NULL,
    to_address TEXT NOT NULL,
    status TEXT NOT NULL, -- 'requested', 'processing', 'broadcasted', 'completed', 'failed', 'rejected'
    tx_hash TEXT,
    ledger_hold_id UUID, -- ID for the initial hold/debit
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Proof of Reserves
CREATE TABLE IF NOT EXISTS por_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id TEXT REFERENCES assets(id),
    merkle_root TEXT NOT NULL,
    total_liabilities NUMERIC NOT NULL,
    block_height INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
