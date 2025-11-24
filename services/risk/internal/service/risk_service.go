package service

import (
	"context"
	"fmt"
	"log"

	"github.com/vnykmshr/nivo/services/risk/internal/models"
	"github.com/vnykmshr/nivo/services/risk/internal/repository"
	"github.com/vnykmshr/nivo/shared/errors"
)

// RiskService handles risk evaluation logic
type RiskService struct {
	ruleRepo  *repository.RiskRuleRepository
	eventRepo *repository.RiskEventRepository
}

// NewRiskService creates a new risk service
func NewRiskService(ruleRepo *repository.RiskRuleRepository, eventRepo *repository.RiskEventRepository) *RiskService {
	return &RiskService{
		ruleRepo:  ruleRepo,
		eventRepo: eventRepo,
	}
}

// EvaluateTransaction evaluates a transaction against all enabled risk rules
func (s *RiskService) EvaluateTransaction(ctx context.Context, req *models.EvaluationRequest) (*models.EvaluationResult, *errors.Error) {
	// Get all enabled rules
	rules, err := s.ruleRepo.GetAll(ctx, true)
	if err != nil {
		return nil, err
	}

	// Initialize result
	result := &models.EvaluationResult{
		Allowed:        true,
		Action:         models.RiskActionAllow,
		RiskScore:      0,
		Reason:         "No risk rules triggered",
		TriggeredRules: []string{},
	}

	// Evaluate each rule
	for _, rule := range rules {
		triggered, score, reason, evalErr := s.evaluateRule(ctx, rule, req)
		if evalErr != nil {
			log.Printf("[risk] Error evaluating rule %s: %v", rule.ID, evalErr)
			continue
		}

		if triggered {
			result.TriggeredRules = append(result.TriggeredRules, rule.ID)

			// Update risk score (use highest score)
			if score > result.RiskScore {
				result.RiskScore = score
			}

			// Determine action (block takes precedence)
			if rule.Action == models.RiskActionBlock {
				result.Allowed = false
				result.Action = models.RiskActionBlock
				result.Reason = reason
			} else if rule.Action == models.RiskActionFlag && result.Action != models.RiskActionBlock {
				result.Action = models.RiskActionFlag
				result.Reason = reason
			}
		}
	}

	// Create risk event for audit trail
	event := &models.RiskEvent{
		TransactionID: req.TransactionID,
		UserID:        req.UserID,
		RiskScore:     result.RiskScore,
		Action:        result.Action,
		Reason:        result.Reason,
		Metadata: map[string]interface{}{
			"amount":           req.Amount,
			"currency":         req.Currency,
			"transaction_type": req.TransactionType,
			"from_wallet_id":   req.FromWalletID,
			"to_wallet_id":     req.ToWalletID,
		},
	}

	// If rules were triggered, set rule ID and type
	if len(result.TriggeredRules) > 0 {
		// Use the first triggered rule for event
		firstRule, ruleErr := s.ruleRepo.GetByID(ctx, result.TriggeredRules[0])
		if ruleErr == nil {
			event.RuleID = &firstRule.ID
			event.RuleType = &firstRule.RuleType
		}
	}

	// Save event
	if createErr := s.eventRepo.Create(ctx, event); createErr != nil {
		log.Printf("[risk] Failed to create risk event: %v", createErr)
	} else {
		result.EventID = event.ID
	}

	return result, nil
}

// evaluateRule evaluates a single rule
func (s *RiskService) evaluateRule(ctx context.Context, rule *models.RiskRule, req *models.EvaluationRequest) (triggered bool, score int, reason string, err *errors.Error) {
	switch rule.RuleType {
	case models.RuleTypeVelocity:
		return s.evaluateVelocityRule(ctx, rule, req)
	case models.RuleTypeDailyLimit:
		return s.evaluateDailyLimitRule(ctx, rule, req)
	case models.RuleTypeThreshold:
		return s.evaluateThresholdRule(ctx, rule, req)
	default:
		return false, 0, "", errors.Internal(fmt.Sprintf("unknown rule type: %s", rule.RuleType))
	}
}

// evaluateVelocityRule checks transaction velocity
func (s *RiskService) evaluateVelocityRule(ctx context.Context, rule *models.RiskRule, req *models.EvaluationRequest) (bool, int, string, *errors.Error) {
	var params models.VelocityRuleParams
	if err := rule.UnmarshalParameters(&params); err != nil {
		return false, 0, "", errors.Internal("failed to unmarshal velocity params")
	}

	// Count recent transactions
	count, err := s.eventRepo.CountUserTransactions(ctx, req.UserID, params.TimeWindowMins)
	if err != nil {
		return false, 0, "", err
	}

	// Check if velocity limit exceeded
	if count >= params.MaxTransactions {
		score := 70 + (count-params.MaxTransactions)*5 // Increase score with excess
		if score > 100 {
			score = 100
		}

		reason := fmt.Sprintf("Velocity limit exceeded: %d transactions in last %d minutes (max: %d)",
			count+1, params.TimeWindowMins, params.MaxTransactions)

		return true, score, reason, nil
	}

	return false, 0, "", nil
}

// evaluateDailyLimitRule checks daily transaction limit
func (s *RiskService) evaluateDailyLimitRule(ctx context.Context, rule *models.RiskRule, req *models.EvaluationRequest) (bool, int, string, *errors.Error) {
	var params models.DailyLimitParams
	if err := rule.UnmarshalParameters(&params); err != nil {
		return false, 0, "", errors.Internal("failed to unmarshal daily limit params")
	}

	// Check currency matches
	if params.Currency != req.Currency {
		return false, 0, "", nil
	}

	// Get user's daily total
	dailyTotal, err := s.eventRepo.GetUserDailyTotal(ctx, req.UserID)
	if err != nil {
		return false, 0, "", err
	}

	// Check if adding this transaction would exceed limit
	newTotal := dailyTotal + req.Amount
	if newTotal > params.MaxAmount {
		score := 80
		percentOver := float64(newTotal-params.MaxAmount) / float64(params.MaxAmount) * 100
		score += int(percentOver / 10)
		if score > 100 {
			score = 100
		}

		reason := fmt.Sprintf("Daily limit exceeded: %d %s today + %d %s = %d %s (max: %d %s)",
			dailyTotal, req.Currency, req.Amount, req.Currency, newTotal, req.Currency, params.MaxAmount, req.Currency)

		return true, score, reason, nil
	}

	return false, 0, "", nil
}

// evaluateThresholdRule checks transaction amount threshold
func (s *RiskService) evaluateThresholdRule(ctx context.Context, rule *models.RiskRule, req *models.EvaluationRequest) (bool, int, string, *errors.Error) {
	var params models.ThresholdParams
	if err := rule.UnmarshalParameters(&params); err != nil {
		return false, 0, "", errors.Internal("failed to unmarshal threshold params")
	}

	// Check currency matches
	if params.Currency != req.Currency {
		return false, 0, "", nil
	}

	// Check if amount is above threshold
	if req.Amount > params.MaxAmount {
		score := 60
		percentOver := float64(req.Amount-params.MaxAmount) / float64(params.MaxAmount) * 100
		score += int(percentOver / 20)
		if score > 100 {
			score = 100
		}

		reason := fmt.Sprintf("Large transaction: %d %s exceeds threshold of %d %s",
			req.Amount, req.Currency, params.MaxAmount, req.Currency)

		return true, score, reason, nil
	}

	return false, 0, "", nil
}

// GetRuleByID retrieves a risk rule by ID
func (s *RiskService) GetRuleByID(ctx context.Context, id string) (*models.RiskRule, *errors.Error) {
	return s.ruleRepo.GetByID(ctx, id)
}

// GetAllRules retrieves all risk rules
func (s *RiskService) GetAllRules(ctx context.Context, enabledOnly bool) ([]*models.RiskRule, *errors.Error) {
	return s.ruleRepo.GetAll(ctx, enabledOnly)
}

// CreateRule creates a new risk rule
func (s *RiskService) CreateRule(ctx context.Context, rule *models.RiskRule) *errors.Error {
	return s.ruleRepo.Create(ctx, rule)
}

// UpdateRule updates a risk rule
func (s *RiskService) UpdateRule(ctx context.Context, rule *models.RiskRule) *errors.Error {
	return s.ruleRepo.Update(ctx, rule)
}

// DeleteRule deletes a risk rule
func (s *RiskService) DeleteRule(ctx context.Context, id string) *errors.Error {
	return s.ruleRepo.Delete(ctx, id)
}

// GetEventByID retrieves a risk event by ID
func (s *RiskService) GetEventByID(ctx context.Context, id string) (*models.RiskEvent, *errors.Error) {
	return s.eventRepo.GetByID(ctx, id)
}

// GetEventsByTransactionID retrieves risk events for a transaction
func (s *RiskService) GetEventsByTransactionID(ctx context.Context, transactionID string) ([]*models.RiskEvent, *errors.Error) {
	return s.eventRepo.GetByTransactionID(ctx, transactionID)
}

// GetEventsByUserID retrieves risk events for a user
func (s *RiskService) GetEventsByUserID(ctx context.Context, userID string, limit int) ([]*models.RiskEvent, *errors.Error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}
	return s.eventRepo.GetByUserID(ctx, userID, limit)
}
