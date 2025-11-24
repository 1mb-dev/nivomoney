package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vnykmshr/nivo/services/risk/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// RiskRuleRepository handles database operations for risk rules
type RiskRuleRepository struct {
	db *sql.DB
}

// NewRiskRuleRepository creates a new risk rule repository
func NewRiskRuleRepository(db *sql.DB) *RiskRuleRepository {
	return &RiskRuleRepository{db: db}
}

// Create creates a new risk rule
func (r *RiskRuleRepository) Create(ctx context.Context, rule *models.RiskRule) *errors.Error {
	paramsJSON, err := json.Marshal(rule.Parameters)
	if err != nil {
		return errors.Internal("failed to marshal parameters")
	}

	query := `
		INSERT INTO risk_rules (rule_type, name, parameters, action, enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		rule.RuleType,
		rule.Name,
		paramsJSON,
		rule.Action,
		rule.Enabled,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return errors.Conflict("risk rule with this name already exists")
		}
		return errors.DatabaseWrap(err, "failed to create risk rule")
	}

	return nil
}

// GetByID retrieves a risk rule by ID
func (r *RiskRuleRepository) GetByID(ctx context.Context, id string) (*models.RiskRule, *errors.Error) {
	rule := &models.RiskRule{}
	var paramsJSON []byte

	query := `
		SELECT id, rule_type, name, parameters, action, enabled, created_at, updated_at
		FROM risk_rules
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.RuleType,
		&rule.Name,
		&paramsJSON,
		&rule.Action,
		&rule.Enabled,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("risk rule not found")
	}
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get risk rule")
	}

	// Unmarshal parameters
	if err := json.Unmarshal(paramsJSON, &rule.Parameters); err != nil {
		return nil, errors.Internal("failed to unmarshal parameters")
	}

	return rule, nil
}

// GetAll retrieves all risk rules
func (r *RiskRuleRepository) GetAll(ctx context.Context, enabledOnly bool) ([]*models.RiskRule, *errors.Error) {
	query := `
		SELECT id, rule_type, name, parameters, action, enabled, created_at, updated_at
		FROM risk_rules
	`

	var args []interface{}
	if enabledOnly {
		query += " WHERE enabled = $1"
		args = append(args, true)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get risk rules")
	}
	defer rows.Close()

	var rules []*models.RiskRule
	for rows.Next() {
		rule := &models.RiskRule{}
		var paramsJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.RuleType,
			&rule.Name,
			&paramsJSON,
			&rule.Action,
			&rule.Enabled,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)

		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan risk rule")
		}

		// Unmarshal parameters
		if err := json.Unmarshal(paramsJSON, &rule.Parameters); err != nil {
			return nil, errors.Internal("failed to unmarshal parameters")
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// GetByType retrieves all enabled risk rules of a specific type
func (r *RiskRuleRepository) GetByType(ctx context.Context, ruleType models.RuleType) ([]*models.RiskRule, *errors.Error) {
	query := `
		SELECT id, rule_type, name, parameters, action, enabled, created_at, updated_at
		FROM risk_rules
		WHERE rule_type = $1 AND enabled = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, ruleType)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get risk rules by type")
	}
	defer rows.Close()

	var rules []*models.RiskRule
	for rows.Next() {
		rule := &models.RiskRule{}
		var paramsJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.RuleType,
			&rule.Name,
			&paramsJSON,
			&rule.Action,
			&rule.Enabled,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)

		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan risk rule")
		}

		// Unmarshal parameters
		if err := json.Unmarshal(paramsJSON, &rule.Parameters); err != nil {
			return nil, errors.Internal("failed to unmarshal parameters")
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// Update updates a risk rule
func (r *RiskRuleRepository) Update(ctx context.Context, rule *models.RiskRule) *errors.Error {
	paramsJSON, err := json.Marshal(rule.Parameters)
	if err != nil {
		return errors.Internal("failed to marshal parameters")
	}

	query := `
		UPDATE risk_rules
		SET rule_type = $1, name = $2, parameters = $3, action = $4, enabled = $5
		WHERE id = $6
		RETURNING updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		rule.RuleType,
		rule.Name,
		paramsJSON,
		rule.Action,
		rule.Enabled,
		rule.ID,
	).Scan(&rule.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.NotFound("risk rule not found")
	}
	if err != nil {
		if isUniqueViolation(err) {
			return errors.Conflict("risk rule with this name already exists")
		}
		return errors.DatabaseWrap(err, "failed to update risk rule")
	}

	return nil
}

// Delete deletes a risk rule
func (r *RiskRuleRepository) Delete(ctx context.Context, id string) *errors.Error {
	query := `DELETE FROM risk_rules WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete risk rule")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseWrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NotFound("risk rule not found")
	}

	return nil
}

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	// PostgreSQL unique violation error code is 23505
	return err != nil && (err.Error() == "pq: duplicate key value violates unique constraint" ||
		// Also check for the error code
		(len(err.Error()) >= 5 && err.Error()[0:5] == "ERROR"))
}
