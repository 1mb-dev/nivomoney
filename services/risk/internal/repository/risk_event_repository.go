package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vnykmshr/nivo/services/risk/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// RiskEventRepository handles database operations for risk events
type RiskEventRepository struct {
	db *sql.DB
}

// NewRiskEventRepository creates a new risk event repository
func NewRiskEventRepository(db *sql.DB) *RiskEventRepository {
	return &RiskEventRepository{db: db}
}

// Create creates a new risk event
func (r *RiskEventRepository) Create(ctx context.Context, event *models.RiskEvent) *errors.Error {
	var metadataJSON []byte
	var err error

	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata")
		}
	}

	query := `
		INSERT INTO risk_events (transaction_id, user_id, rule_id, rule_type, risk_score, action, reason, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	err = r.db.QueryRowContext(ctx, query,
		event.TransactionID,
		event.UserID,
		event.RuleID,
		event.RuleType,
		event.RiskScore,
		event.Action,
		event.Reason,
		metadataJSON,
	).Scan(&event.ID, &event.CreatedAt)

	if err != nil {
		return errors.DatabaseWrap(err, "failed to create risk event")
	}

	return nil
}

// GetByID retrieves a risk event by ID
func (r *RiskEventRepository) GetByID(ctx context.Context, id string) (*models.RiskEvent, *errors.Error) {
	event := &models.RiskEvent{}
	var metadataJSON []byte

	query := `
		SELECT id, transaction_id, user_id, rule_id, rule_type, risk_score, action, reason, metadata, created_at
		FROM risk_events
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.TransactionID,
		&event.UserID,
		&event.RuleID,
		&event.RuleType,
		&event.RiskScore,
		&event.Action,
		&event.Reason,
		&metadataJSON,
		&event.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("risk event not found")
	}
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get risk event")
	}

	// Unmarshal metadata if present
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, errors.Internal("failed to unmarshal metadata")
		}
	}

	return event, nil
}

// GetByTransactionID retrieves risk events for a transaction
func (r *RiskEventRepository) GetByTransactionID(ctx context.Context, transactionID string) ([]*models.RiskEvent, *errors.Error) {
	query := `
		SELECT id, transaction_id, user_id, rule_id, rule_type, risk_score, action, reason, metadata, created_at
		FROM risk_events
		WHERE transaction_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get risk events by transaction")
	}
	defer func() { _ = rows.Close() }()

	var events []*models.RiskEvent
	for rows.Next() {
		event := &models.RiskEvent{}
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.TransactionID,
			&event.UserID,
			&event.RuleID,
			&event.RuleType,
			&event.RiskScore,
			&event.Action,
			&event.Reason,
			&metadataJSON,
			&event.CreatedAt,
		)

		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan risk event")
		}

		// Unmarshal metadata if present
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, errors.Internal("failed to unmarshal metadata")
			}
		}

		events = append(events, event)
	}

	return events, nil
}

// GetByUserID retrieves risk events for a user
func (r *RiskEventRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*models.RiskEvent, *errors.Error) {
	query := `
		SELECT id, transaction_id, user_id, rule_id, rule_type, risk_score, action, reason, metadata, created_at
		FROM risk_events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get risk events by user")
	}
	defer func() { _ = rows.Close() }()

	var events []*models.RiskEvent
	for rows.Next() {
		event := &models.RiskEvent{}
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.TransactionID,
			&event.UserID,
			&event.RuleID,
			&event.RuleType,
			&event.RiskScore,
			&event.Action,
			&event.Reason,
			&metadataJSON,
			&event.CreatedAt,
		)

		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan risk event")
		}

		// Unmarshal metadata if present
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, errors.Internal("failed to unmarshal metadata")
			}
		}

		events = append(events, event)
	}

	return events, nil
}

// CountUserTransactions counts user transactions in a time window
func (r *RiskEventRepository) CountUserTransactions(ctx context.Context, userID string, minutesAgo int) (int, *errors.Error) {
	query := `
		SELECT COUNT(DISTINCT transaction_id)
		FROM risk_events
		WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '1 minute' * $2
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, minutesAgo).Scan(&count)
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to count user transactions")
	}

	return count, nil
}

// GetUserDailyTotal calculates total amount for user today
// Note: This requires metadata to contain amount information
func (r *RiskEventRepository) GetUserDailyTotal(ctx context.Context, userID string) (int64, *errors.Error) {
	query := `
		SELECT COALESCE(SUM((metadata->>'amount')::bigint), 0)
		FROM risk_events
		WHERE user_id = $1
		  AND created_at >= CURRENT_DATE
		  AND action != 'block'
		  AND metadata->>'amount' IS NOT NULL
	`

	var total int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to get user daily total")
	}

	return total, nil
}
