package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/1mb-dev/nivomoney/services/risk/internal/models"
	"github.com/1mb-dev/nivomoney/services/risk/internal/service"
	"github.com/1mb-dev/nivomoney/shared/errors"
	"github.com/1mb-dev/nivomoney/shared/response"
)

// RiskHandler handles HTTP requests for risk evaluation
type RiskHandler struct {
	riskService *service.RiskService
}

// NewRiskHandler creates a new risk handler
func NewRiskHandler(riskService *service.RiskService) *RiskHandler {
	return &RiskHandler{
		riskService: riskService,
	}
}

// EvaluateTransaction handles POST /api/v1/risk/evaluate
func (h *RiskHandler) EvaluateTransaction(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse request
	var req models.EvaluationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Validate request
	if req.TransactionID == "" {
		response.Error(w, errors.Validation("transaction_id is required"))
		return
	}
	if req.UserID == "" {
		response.Error(w, errors.Validation("user_id is required"))
		return
	}
	if req.Amount <= 0 {
		response.Error(w, errors.Validation("amount must be greater than 0"))
		return
	}
	if req.Currency == "" {
		response.Error(w, errors.Validation("currency is required"))
		return
	}

	// Evaluate transaction
	result, svcErr := h.riskService.EvaluateTransaction(r.Context(), &req)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, result)
}

// GetRuleByID handles GET /api/v1/risk/rules/:id
func (h *RiskHandler) GetRuleByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, errors.BadRequest("rule ID is required"))
		return
	}

	rule, err := h.riskService.GetRuleByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, rule)
}

// GetAllRules handles GET /api/v1/risk/rules
func (h *RiskHandler) GetAllRules(w http.ResponseWriter, r *http.Request) {
	enabledOnly := r.URL.Query().Get("enabled") == "true"

	rules, err := h.riskService.GetAllRules(r.Context(), enabledOnly)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, rules)
}

// CreateRule handles POST /api/v1/risk/rules
func (h *RiskHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse request
	var rule models.RiskRule
	if err := json.Unmarshal(body, &rule); err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Validate request
	if rule.RuleType == "" {
		response.Error(w, errors.Validation("rule_type is required"))
		return
	}
	if rule.Name == "" {
		response.Error(w, errors.Validation("name is required"))
		return
	}
	if rule.Parameters == nil {
		response.Error(w, errors.Validation("parameters are required"))
		return
	}

	// Create rule
	if svcErr := h.riskService.CreateRule(r.Context(), &rule); svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.Created(w, rule)
}

// UpdateRule handles PUT /api/v1/risk/rules/:id
func (h *RiskHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, errors.BadRequest("rule ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse request
	var rule models.RiskRule
	if err := json.Unmarshal(body, &rule); err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	rule.ID = id

	// Update rule
	if svcErr := h.riskService.UpdateRule(r.Context(), &rule); svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, rule)
}

// DeleteRule handles DELETE /api/v1/risk/rules/:id
func (h *RiskHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, errors.BadRequest("rule ID is required"))
		return
	}

	if err := h.riskService.DeleteRule(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.NoContent(w)
}

// GetEventByID handles GET /api/v1/risk/events/:id
func (h *RiskHandler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, errors.BadRequest("event ID is required"))
		return
	}

	event, err := h.riskService.GetEventByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, event)
}

// GetEventsByTransactionID handles GET /api/v1/risk/transactions/:transactionId/events
func (h *RiskHandler) GetEventsByTransactionID(w http.ResponseWriter, r *http.Request) {
	transactionID := r.PathValue("transactionId")
	if transactionID == "" {
		response.Error(w, errors.BadRequest("transaction ID is required"))
		return
	}

	events, err := h.riskService.GetEventsByTransactionID(r.Context(), transactionID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, events)
}

// GetEventsByUserID handles GET /api/v1/risk/users/:userId/events
func (h *RiskHandler) GetEventsByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.Error(w, errors.BadRequest("user ID is required"))
		return
	}

	// Parse limit from query param (default: 100, max: 1000)
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var parsedLimit int
		if _, err := fmt.Sscanf(limitStr, "%d", &parsedLimit); err == nil {
			limit = parsedLimit
		}
	}

	events, err := h.riskService.GetEventsByUserID(r.Context(), userID, limit)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, events)
}
