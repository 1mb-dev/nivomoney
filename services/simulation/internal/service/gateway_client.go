package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// GatewayClient makes API calls to the Nivo Gateway
type GatewayClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string // Admin token for simulations
}

// NewGatewayClient creates a new gateway client
func NewGatewayClient(baseURL, authToken string) *GatewayClient {
	return &GatewayClient{
		baseURL:   baseURL,
		authToken: authToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DepositRequest represents a deposit transaction request
type DepositRequest struct {
	WalletID    string `json:"wallet_id"`
	AmountPaise int64  `json:"amount_paise"`
	Description string `json:"description"`
}

// TransferRequest represents a transfer transaction request
type TransferRequest struct {
	SourceWalletID      string `json:"source_wallet_id"`
	DestinationWalletID string `json:"destination_wallet_id"`
	AmountPaise         int64  `json:"amount_paise"`
	Description         string `json:"description"`
}

// WithdrawalRequest represents a withdrawal transaction request
type WithdrawalRequest struct {
	WalletID    string `json:"wallet_id"`
	AmountPaise int64  `json:"amount_paise"`
	Description string `json:"description"`
}

// CreateDeposit creates a deposit transaction
func (c *GatewayClient) CreateDeposit(walletID string, amountPaise int64, description string) error {
	req := DepositRequest{
		WalletID:    walletID,
		AmountPaise: amountPaise,
		Description: description,
	}

	return c.makeRequest("POST", "/api/v1/transaction/transactions/deposit", req)
}

// CreateTransfer creates a transfer transaction
func (c *GatewayClient) CreateTransfer(sourceWalletID, destWalletID string, amountPaise int64, description string) error {
	req := TransferRequest{
		SourceWalletID:      sourceWalletID,
		DestinationWalletID: destWalletID,
		AmountPaise:         amountPaise,
		Description:         description,
	}

	return c.makeRequest("POST", "/api/v1/transaction/transactions/transfer", req)
}

// CreateWithdrawal creates a withdrawal transaction
func (c *GatewayClient) CreateWithdrawal(walletID string, amountPaise int64, description string) error {
	req := WithdrawalRequest{
		WalletID:    walletID,
		AmountPaise: amountPaise,
		Description: description,
	}

	return c.makeRequest("POST", "/api/v1/transaction/transactions/withdrawal", req)
}

// makeRequest is a helper to make HTTP requests
func (c *GatewayClient) makeRequest(method, path string, body interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		log.Printf("[simulation] API error %d: %s", resp.StatusCode, string(responseBody))
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	log.Printf("[simulation] Transaction created successfully: %s %s", method, path)
	return nil
}
