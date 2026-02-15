package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/1mb-dev/nivomoney/services/ledger/internal/models"
	"github.com/1mb-dev/nivomoney/shared/database"
	"github.com/1mb-dev/nivomoney/shared/errors"
)

// AccountRepository handles database operations for accounts.
type AccountRepository struct {
	db *database.DB
}

// NewAccountRepository creates a new account repository.
func NewAccountRepository(db *database.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create creates a new account.
func (r *AccountRepository) Create(ctx context.Context, account *models.Account) *errors.Error {
	// Serialize metadata
	metadataJSON, err := json.Marshal(account.Metadata)
	if err != nil {
		return errors.BadRequest("invalid metadata format")
	}

	query := `
		INSERT INTO accounts (code, name, type, currency, parent_id, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, balance, debit_total, credit_total, created_at, updated_at
	`

	scanErr := r.db.QueryRowContext(ctx, query,
		account.Code,
		account.Name,
		account.Type,
		account.Currency,
		account.ParentID,
		account.Status,
		metadataJSON,
	).Scan(
		&account.ID,
		&account.Balance,
		&account.DebitTotal,
		&account.CreditTotal,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if scanErr != nil {
		if database.IsUniqueViolation(scanErr) {
			return errors.Conflict("account with this code already exists")
		}
		return errors.DatabaseWrap(scanErr, "failed to create account")
	}

	return nil
}

// GetByID retrieves an account by ID.
func (r *AccountRepository) GetByID(ctx context.Context, id string) (*models.Account, *errors.Error) {
	account := &models.Account{}
	var metadataJSON []byte

	query := `
		SELECT id, code, name, type, currency, parent_id, balance, debit_total,
		       credit_total, status, metadata, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.Code,
		&account.Name,
		&account.Type,
		&account.Currency,
		&account.ParentID,
		&account.Balance,
		&account.DebitTotal,
		&account.CreditTotal,
		&account.Status,
		&metadataJSON,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("account", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get account")
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &account.Metadata); err != nil {
			return nil, errors.Internal("failed to parse metadata")
		}
	}

	return account, nil
}

// GetByCode retrieves an account by code.
func (r *AccountRepository) GetByCode(ctx context.Context, code string) (*models.Account, *errors.Error) {
	account := &models.Account{}
	var metadataJSON []byte

	query := `
		SELECT id, code, name, type, currency, parent_id, balance, debit_total,
		       credit_total, status, metadata, created_at, updated_at
		FROM accounts
		WHERE code = $1
	`

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&account.ID,
		&account.Code,
		&account.Name,
		&account.Type,
		&account.Currency,
		&account.ParentID,
		&account.Balance,
		&account.DebitTotal,
		&account.CreditTotal,
		&account.Status,
		&metadataJSON,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("account")
		}
		return nil, errors.DatabaseWrap(err, "failed to get account by code")
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &account.Metadata); err != nil {
			return nil, errors.Internal("failed to parse metadata")
		}
	}

	return account, nil
}

// List retrieves accounts with filters.
func (r *AccountRepository) List(ctx context.Context, accountType *models.AccountType, status *models.AccountStatus, limit, offset int) ([]*models.Account, *errors.Error) {
	query := `
		SELECT id, code, name, type, currency, parent_id, balance, debit_total,
		       credit_total, status, metadata, created_at, updated_at
		FROM accounts
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if accountType != nil {
		query += ` AND type = $` + string(rune('0'+argPos))
		args = append(args, *accountType)
		argPos++
	}

	if status != nil {
		query += ` AND status = $` + string(rune('0'+argPos))
		args = append(args, *status)
		argPos++
	}

	query += ` ORDER BY code LIMIT $` + string(rune('0'+argPos)) + ` OFFSET $` + string(rune('0'+argPos+1))
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list accounts")
	}
	defer func() { _ = rows.Close() }()

	accounts := make([]*models.Account, 0)
	for rows.Next() {
		account := &models.Account{}
		var metadataJSON []byte

		err := rows.Scan(
			&account.ID,
			&account.Code,
			&account.Name,
			&account.Type,
			&account.Currency,
			&account.ParentID,
			&account.Balance,
			&account.DebitTotal,
			&account.CreditTotal,
			&account.Status,
			&metadataJSON,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan account")
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &account.Metadata); err != nil {
				return nil, errors.Internal("failed to parse metadata")
			}
		}

		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating accounts")
	}

	return accounts, nil
}

// Update updates an account.
func (r *AccountRepository) Update(ctx context.Context, account *models.Account) *errors.Error {
	// Serialize metadata
	metadataJSON, err := json.Marshal(account.Metadata)
	if err != nil {
		return errors.BadRequest("invalid metadata format")
	}

	query := `
		UPDATE accounts
		SET name = $2, status = $3, metadata = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	scanErr := r.db.QueryRowContext(ctx, query,
		account.ID,
		account.Name,
		account.Status,
		metadataJSON,
	).Scan(&account.UpdatedAt)

	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return errors.NotFoundWithID("account", account.ID)
		}
		return errors.DatabaseWrap(scanErr, "failed to update account")
	}

	return nil
}

// GetBalance retrieves the current balance of an account.
func (r *AccountRepository) GetBalance(ctx context.Context, accountID string) (int64, *errors.Error) {
	var balance int64

	query := `SELECT balance FROM accounts WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, accountID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.NotFoundWithID("account", accountID)
		}
		return 0, errors.DatabaseWrap(err, "failed to get account balance")
	}

	return balance, nil
}
