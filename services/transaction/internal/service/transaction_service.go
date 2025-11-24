package service

import (
	"context"
	"fmt"
	"log"

	"github.com/vnykmshr/nivo/services/transaction/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/events"
)

// TransactionRepositoryInterface defines the interface for transaction repository operations.
type TransactionRepositoryInterface interface {
	Create(ctx context.Context, transaction *models.Transaction) *errors.Error
	GetByID(ctx context.Context, id string) (*models.Transaction, *errors.Error)
	ListByWallet(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error)
}

// TransactionService handles business logic for transaction operations.
type TransactionService struct {
	transactionRepo TransactionRepositoryInterface
	riskClient      *RiskClient
	eventPublisher  *events.Publisher
}

// NewTransactionService creates a new transaction service.
func NewTransactionService(transactionRepo TransactionRepositoryInterface, riskClient *RiskClient, eventPublisher *events.Publisher) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		riskClient:      riskClient,
		eventPublisher:  eventPublisher,
	}
}

// CreateTransfer creates a transfer transaction between wallets.
func (s *TransactionService) CreateTransfer(ctx context.Context, req *models.CreateTransferRequest) (*models.Transaction, *errors.Error) {
	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	// Validate source and destination are different
	if req.SourceWalletID == req.DestinationWalletID {
		return nil, errors.BadRequest("source and destination wallets must be different")
	}

	// Create transaction
	sourceWalletID := req.SourceWalletID
	destWalletID := req.DestinationWalletID
	var reference *string
	if req.Reference != "" {
		reference = &req.Reference
	}

	transaction := &models.Transaction{
		Type:                models.TransactionTypeTransfer,
		Status:              models.TransactionStatusPending,
		SourceWalletID:      &sourceWalletID,
		DestinationWalletID: &destWalletID,
		Amount:              req.Amount,
		Currency:            req.Currency,
		Description:         req.Description,
		Reference:           reference,
		Metadata:            metadata,
	}

	if createErr := s.transactionRepo.Create(ctx, transaction); createErr != nil {
		return nil, createErr
	}

	// Publish transaction.created event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishTransactionEvent("transaction.created", transaction.ID, map[string]interface{}{
			"type":                  string(transaction.Type),
			"status":                string(transaction.Status),
			"amount":                transaction.Amount,
			"currency":              transaction.Currency,
			"source_wallet_id":      transaction.SourceWalletID,
			"destination_wallet_id": transaction.DestinationWalletID,
			"description":           transaction.Description,
		})
	}

	// Evaluate risk for the transaction
	if evalErr := s.evaluateTransactionRisk(ctx, transaction); evalErr != nil {
		log.Printf("[transaction] Risk evaluation failed for transaction %s: %v", transaction.ID, evalErr)
		// Continue processing even if risk evaluation fails (fail open for now)
	}

	// TODO: In production, trigger async processing:
	// 1. Verify source wallet has sufficient balance
	// 2. Create hold on source wallet
	// 3. Create ledger entry
	// 4. Update wallet balances
	// 5. Mark transaction as completed
	// For now, transaction remains in pending state

	return transaction, nil
}

// CreateDeposit creates a deposit transaction to a wallet.
func (s *TransactionService) CreateDeposit(ctx context.Context, req *models.CreateDepositRequest) (*models.Transaction, *errors.Error) {
	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	destWalletID := req.WalletID
	var reference *string
	if req.Reference != "" {
		reference = &req.Reference
	}

	transaction := &models.Transaction{
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusPending,
		DestinationWalletID: &destWalletID,
		Amount:              req.Amount,
		Currency:            req.Currency,
		Description:         req.Description,
		Reference:           reference,
		Metadata:            metadata,
	}

	if createErr := s.transactionRepo.Create(ctx, transaction); createErr != nil {
		return nil, createErr
	}

	// Publish transaction.created event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishTransactionEvent("transaction.created", transaction.ID, map[string]interface{}{
			"type":                  string(transaction.Type),
			"status":                string(transaction.Status),
			"amount":                transaction.Amount,
			"currency":              transaction.Currency,
			"destination_wallet_id": transaction.DestinationWalletID,
			"description":           transaction.Description,
		})
	}

	// TODO: Trigger async processing for deposit
	// 1. Verify external payment received
	// 2. Create ledger entry
	// 3. Update wallet balance
	// 4. Mark transaction as completed

	return transaction, nil
}

// CreateWithdrawal creates a withdrawal transaction from a wallet.
func (s *TransactionService) CreateWithdrawal(ctx context.Context, req *models.CreateWithdrawalRequest) (*models.Transaction, *errors.Error) {
	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	sourceWalletID := req.WalletID
	var reference *string
	if req.Reference != "" {
		reference = &req.Reference
	}

	transaction := &models.Transaction{
		Type:           models.TransactionTypeWithdrawal,
		Status:         models.TransactionStatusPending,
		SourceWalletID: &sourceWalletID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Description:    req.Description,
		Reference:      reference,
		Metadata:       metadata,
	}

	if createErr := s.transactionRepo.Create(ctx, transaction); createErr != nil {
		return nil, createErr
	}

	// Publish transaction.created event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishTransactionEvent("transaction.created", transaction.ID, map[string]interface{}{
			"type":             string(transaction.Type),
			"status":           string(transaction.Status),
			"amount":           transaction.Amount,
			"currency":         transaction.Currency,
			"source_wallet_id": transaction.SourceWalletID,
			"description":      transaction.Description,
		})
	}

	// TODO: Trigger async processing for withdrawal
	// 1. Verify wallet has sufficient balance
	// 2. Create hold on wallet
	// 3. Initiate external payment
	// 4. Create ledger entry
	// 5. Update wallet balance
	// 6. Mark transaction as completed

	return transaction, nil
}

// GetTransaction retrieves a transaction by ID.
func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*models.Transaction, *errors.Error) {
	return s.transactionRepo.GetByID(ctx, id)
}

// ListWalletTransactions retrieves transactions for a wallet.
func (s *TransactionService) ListWalletTransactions(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error) {
	return s.transactionRepo.ListByWallet(ctx, walletID, filter)
}

// ReverseTransaction reverses a completed transaction.
func (s *TransactionService) ReverseTransaction(ctx context.Context, transactionID, reason string) (*models.Transaction, *errors.Error) {
	// Get original transaction
	originalTx, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Validate transaction can be reversed
	if !originalTx.IsCompleted() {
		return nil, errors.BadRequest("only completed transactions can be reversed")
	}

	if originalTx.Type == models.TransactionTypeReversal {
		return nil, errors.BadRequest("cannot reverse a reversal transaction")
	}

	// Create reversal transaction
	parentID := transactionID
	reversalTx := &models.Transaction{
		Type:                models.TransactionTypeReversal,
		Status:              models.TransactionStatusPending,
		SourceWalletID:      originalTx.DestinationWalletID, // Reverse direction
		DestinationWalletID: originalTx.SourceWalletID,
		Amount:              originalTx.Amount,
		Currency:            originalTx.Currency,
		Description:         "Reversal: " + reason,
		ParentTransactionID: &parentID,
		Metadata:            map[string]string{"reversal_reason": reason},
	}

	if createErr := s.transactionRepo.Create(ctx, reversalTx); createErr != nil {
		return nil, createErr
	}

	// TODO: Trigger async processing for reversal
	// 1. Create reversal ledger entry
	// 2. Update wallet balances
	// 3. Mark reversal as completed
	// 4. Mark original transaction as reversed

	return reversalTx, nil
}

// evaluateTransactionRisk evaluates risk for a transaction using the Risk Service.
func (s *TransactionService) evaluateTransactionRisk(ctx context.Context, transaction *models.Transaction) error {
	if s.riskClient == nil {
		log.Printf("[transaction] Risk client not configured, skipping risk evaluation")
		return nil
	}

	// Extract user ID from wallet ownership (for now, use a placeholder)
	// In production, you would fetch the wallet owner from the wallet service
	userID := "unknown"

	// Prepare risk evaluation request
	riskReq := &RiskEvaluationRequest{
		TransactionID:   transaction.ID,
		UserID:          userID,
		Amount:          transaction.Amount,
		Currency:        string(transaction.Currency),
		TransactionType: string(transaction.Type),
	}

	if transaction.SourceWalletID != nil {
		riskReq.FromWalletID = *transaction.SourceWalletID
	}
	if transaction.DestinationWalletID != nil {
		riskReq.ToWalletID = *transaction.DestinationWalletID
	}

	// Call risk service
	result, err := s.riskClient.EvaluateTransaction(ctx, riskReq)
	if err != nil {
		return err
	}

	// Log risk evaluation result
	log.Printf("[transaction] Risk evaluation for transaction %s: action=%s, score=%d, allowed=%v",
		transaction.ID, result.Action, result.RiskScore, result.Allowed)

	// Store risk information in transaction metadata
	if transaction.Metadata == nil {
		transaction.Metadata = make(map[string]string)
	}
	transaction.Metadata["risk_score"] = fmt.Sprintf("%d", result.RiskScore)
	transaction.Metadata["risk_action"] = result.Action
	transaction.Metadata["risk_event_id"] = result.EventID

	if len(result.TriggeredRules) > 0 {
		transaction.Metadata["risk_triggered_rules"] = fmt.Sprintf("%d", len(result.TriggeredRules))
	}

	// Handle risk actions
	if !result.Allowed {
		log.Printf("[transaction] Transaction %s BLOCKED by risk evaluation: %s", transaction.ID, result.Reason)
		// In production, you would update the transaction status to failed
		// For now, just log the blocking decision
	} else if result.Action == "flag" {
		log.Printf("[transaction] Transaction %s FLAGGED by risk evaluation: %s", transaction.ID, result.Reason)
		// In production, you might notify compliance team or require manual review
	}

	return nil
}
