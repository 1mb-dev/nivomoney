package handler

import (
	"io"
	"net/http"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/ledger/internal/models"
	"github.com/vnykmshr/nivo/services/ledger/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// LedgerHandler handles HTTP requests for ledger operations.
type LedgerHandler struct {
	ledgerService *service.LedgerService
}

// NewLedgerHandler creates a new ledger handler.
func NewLedgerHandler(ledgerService *service.LedgerService) *LedgerHandler {
	return &LedgerHandler{
		ledgerService: ledgerService,
	}
}

// CreateAccount creates a new ledger account.
// POST /api/v1/accounts
func (h *LedgerHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request (gopantic v1.2.0+ supports json.RawMessage)
	req, err := model.ParseInto[models.CreateAccountRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Create account
	account, svcErr := h.ledgerService.CreateAccount(r.Context(), &req)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.Created(w, account)
}

// GetAccount retrieves an account by ID.
// GET /api/v1/accounts/:id
func (h *LedgerHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		response.Error(w, errors.BadRequest("account ID is required"))
		return
	}

	account, svcErr := h.ledgerService.GetAccount(r.Context(), accountID)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, account)
}

// GetAccountByCode retrieves an account by code.
// GET /api/v1/accounts/code/:code
func (h *LedgerHandler) GetAccountByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		response.Error(w, errors.BadRequest("account code is required"))
		return
	}

	account, svcErr := h.ledgerService.GetAccountByCode(r.Context(), code)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, account)
}

// ListAccounts retrieves accounts with optional filters.
// GET /api/v1/accounts?type=asset&status=active&limit=50&offset=0
func (h *LedgerHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	var accountType *models.AccountType
	if typeParam := r.URL.Query().Get("type"); typeParam != "" {
		t := models.AccountType(typeParam)
		accountType = &t
	}

	var status *models.AccountStatus
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		s := models.AccountStatus(statusParam)
		status = &s
	}

	limit := 50 // default
	offset := 0 // default

	// List accounts
	accounts, svcErr := h.ledgerService.ListAccounts(r.Context(), accountType, status, limit, offset)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, accounts)
}

// UpdateAccount updates an account.
// PUT /api/v1/accounts/:id
func (h *LedgerHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		response.Error(w, errors.BadRequest("account ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request
	req, err := model.ParseInto[models.UpdateAccountRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Update account
	account, svcErr := h.ledgerService.UpdateAccount(r.Context(), accountID, &req)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, account)
}

// GetAccountBalance retrieves the current balance of an account.
// GET /api/v1/accounts/:id/balance
func (h *LedgerHandler) GetAccountBalance(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		response.Error(w, errors.BadRequest("account ID is required"))
		return
	}

	balance, svcErr := h.ledgerService.GetAccountBalance(r.Context(), accountID)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, map[string]interface{}{
		"account_id": accountID,
		"balance":    balance,
	})
}

// CreateJournalEntry creates a new journal entry.
// POST /api/v1/journal-entries
func (h *LedgerHandler) CreateJournalEntry(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request
	req, err := model.ParseInto[models.CreateJournalEntryRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Create journal entry
	entry, svcErr := h.ledgerService.CreateJournalEntry(r.Context(), &req)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.Created(w, entry)
}

// GetJournalEntry retrieves a journal entry with its lines.
// GET /api/v1/journal-entries/:id
func (h *LedgerHandler) GetJournalEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("id")
	if entryID == "" {
		response.Error(w, errors.BadRequest("entry ID is required"))
		return
	}

	entry, svcErr := h.ledgerService.GetJournalEntry(r.Context(), entryID)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, entry)
}

// ListJournalEntries retrieves journal entries with optional filters.
// GET /api/v1/journal-entries?status=posted&limit=50&offset=0
func (h *LedgerHandler) ListJournalEntries(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	var status *models.EntryStatus
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		s := models.EntryStatus(statusParam)
		status = &s
	}

	limit := 50 // default
	offset := 0 // default

	// List journal entries
	entries, svcErr := h.ledgerService.ListJournalEntries(r.Context(), status, limit, offset)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, entries)
}

// PostJournalEntry posts a draft journal entry.
// POST /api/v1/journal-entries/:id/post
func (h *LedgerHandler) PostJournalEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("id")
	if entryID == "" {
		response.Error(w, errors.BadRequest("entry ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request
	req, err := model.ParseInto[models.PostJournalEntryRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Validate entry ID matches
	if req.EntryID != entryID {
		response.Error(w, errors.BadRequest("entry ID mismatch"))
		return
	}

	// Post entry
	entry, svcErr := h.ledgerService.PostJournalEntry(r.Context(), entryID, req.PostedBy)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, entry)
}

// VoidJournalEntry voids a posted journal entry.
// POST /api/v1/journal-entries/:id/void
func (h *LedgerHandler) VoidJournalEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("id")
	if entryID == "" {
		response.Error(w, errors.BadRequest("entry ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request
	req, err := model.ParseInto[models.VoidJournalEntryRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Validate entry ID matches
	if req.EntryID != entryID {
		response.Error(w, errors.BadRequest("entry ID mismatch"))
		return
	}

	// Void entry
	entry, svcErr := h.ledgerService.VoidJournalEntry(r.Context(), entryID, req.VoidedBy, req.VoidReason)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, entry)
}

// ReverseJournalEntry creates a reversing entry.
// POST /api/v1/journal-entries/:id/reverse
func (h *LedgerHandler) ReverseJournalEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("id")
	if entryID == "" {
		response.Error(w, errors.BadRequest("entry ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse request
	type ReverseRequest struct {
		ReversedBy string `json:"reversed_by" validate:"required,uuid"`
		Reason     string `json:"reason" validate:"required,min:10,max:500"`
	}

	req, err := model.ParseInto[ReverseRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Reverse entry
	reversalEntry, svcErr := h.ledgerService.ReverseJournalEntry(r.Context(), entryID, req.ReversedBy, req.Reason)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.Created(w, reversalEntry)
}

// CreateAccountInternal creates a new ledger account (internal endpoint).
// POST /internal/v1/accounts
// This is an internal endpoint for service-to-service communication (no authentication required).
func (h *LedgerHandler) CreateAccountInternal(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request (gopantic v1.2.0+ supports json.RawMessage)
	req, err := model.ParseInto[models.CreateAccountRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Create account
	account, svcErr := h.ledgerService.CreateAccount(r.Context(), &req)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.Created(w, account)
}
