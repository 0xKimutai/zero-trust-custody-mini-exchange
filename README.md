# Mini Centralized Custody Backend

## Overview

This project is a high-integrity, centralized cryptocurrency custody backend designed to simulate the core financial operations of a custodial exchange. It prioritizes financial correctness, security, and auditability over performance or scalability.

The system is built in Go and utilizes a strict double-entry ledger system to ensure that every satoshi or wei is accounted for. It assumes a zero-trust environment where no balance can be modified without a corresponding ledger entry.

## Architecture & Design Principles

### 1. Double-Entry Ledger
At the heart of the system is the Ledger Service. Unlike typical applications that might store a simple `balance` column on a user table, this system records every transaction as a pair of entries (debit and credit) that must sum to zero. This ensures:
- **Immutability**: Balances are derived, not stored statically.
- **Auditability**: Every change in balance can be traced back to a specific transaction.
- **Error Prevention**: It is mathematically impossible to destroy or create assets unintentionally.

### 2. Service-Oriented (Modular Monolith)
The application is structured into distinct domains (Auth, Wallet, Ledger, Deposit, Withdrawal, Proof of Reserves) within a single binary. This enforces separation of concerns while maintaining the simplicity of deployment.
- **Handler Layer**: Manages HTTP requests, input validation, and rate limiting.
- **Service Layer**: Contains business logic and transaction management.
- **Data Layer**: Handles direct database interactions.

### 3. Zero-Trust Withdrawal Flow
Withdrawals function as a state machine to prevent race conditions and unauthorized fund movement:
1.  **Request**: User initiates withdrawal.
2.  **Ledger Hold**: Funds are immediately debited from the user and credited to a system liability hold account.
3.  **Risk Check**: logic checks (e.g., max amounts) are applied.
4.  **Batch Processing**: Administrators trigger batch processing to broadcast transactions (simulated).
5.  **Completion**: On simulated success, the hold is released and the system asset is reduced.

## Features

-   **Authentication**: JWT-based secure user sessions with bcrypt password hashing.
-   **Asset Management**: Support for multiple blockchains (BTC, ETH) and assets (ERC-20 tokens).
-   **Wallet Management**: Abstraction for hot and cold exchange wallets.
-   **Deposit Processing**: Idempotent processing of simulation blockchain webhooks.
-   **Proof of Reserves (PoR)**: Generation of Merkle Trees from user liabilities to allow for public verification of solvency.
-   **API Versioning**: Validated simulation endpoints under `/api/v1`.
-   **Rate Limiting**: IP-based rate limiting to prevent abuse.

## Prerequisites

-   **Go**: Version 1.23 or higher.
-   **PostgreSQL**: A running Postgres instance.

## Installation & Setup

1.  **Clone the Repository**
    ```bash
    git clone <repository-url>
    cd mini-exchange
    ```

2.  **Configuration**
    Copy the example environment file and configure your database credentials.
    ```bash
    cp .env.example .env
    # Edit .env with your favorite editor
    ```

3.  **Build**
    Download dependencies and build the binary.
    ```bash
    go mod tidy
    go build -o api ./cmd/api
    ```

4.  **Run**
    Start the server. The application will automatically apply the database schema on the first run.
    ```bash
    ./api
    ```

## API Documentation

The API is served at `http://localhost:8080/api/v1`.

### Authentication
-   `POST /api/v1/auth/register`: Create a new user account.
-   `POST /api/v1/auth/login`: Authenticate and receive a JWT.

### User Operations
-   `GET /api/v1/me`: Get current user ID.
-   `GET /api/v1/deposit/address?asset=BTC`: Generate or retrieve a deposit address.
-   `POST /api/v1/withdraw`: Request a withdrawal `{ "asset_id": "BTC", "amount": 0.5, "to_address": "..." }`.
-   `GET /api/v1/withdrawals`: View withdrawal history.

### Verification (Proof of Reserves)
-   `POST /admin/por/generate?asset=BTC`: Generate a Merkle Tree snapshot of all user liabilities for an asset.

### Simulation / Admin
-   `POST /admin/deposit/webhook`: Simulate an incoming blockchain deposit.
-   `POST /admin/withdrawal/process`: Trigger the batch processor for pending withdrawals.

## Testing & Verification

A generic flow verification script is included to simulate a user lifecycle (Register -> Deposit -> Withdraw -> Verify).

```bash
# Ensure the API is running, then run:
go run cmd/verify_flow/main.go
```

## Security Considerations

-   **Secrets**: All secrets (DB passwords, JWT keys) are loaded from environment variables.
-   **Input Validation**: All API inputs are bound and validated before processing.
-   **Rate Limiting**: A token-bucket rate limiter protects public and protected endpoints.
-   **SQL Injection**: Parameterized queries are used exclusively for database interactions.

## License
Proprietary / Private API.
