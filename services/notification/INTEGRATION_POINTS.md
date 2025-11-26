# Notification Service Integration Points

This document outlines where the notification service should be integrated into existing services.

## Integration Strategy

Each service will have a NotificationClient that makes HTTP calls to the notification service to send notifications for key user actions.

## Integration Points by Service

### 1. Identity Service (`services/identity`)

#### User Registration
- **File**: `services/identity/internal/service/auth_service.go`
- **Function**: `Register()` (line 94)
- **Trigger**: After successful user registration
- **Notification Type**: `welcome`
- **Channel**: `email` + `sms`
- **Variables**:
  - `full_name`: User's full name
  - `email`: User's email address
- **Template**: `welcome_email` / `welcome_sms`

#### KYC Verification
- **File**: `services/identity/internal/service/auth_service.go`
- **Function**: `VerifyKYC()` (line 340)
- **Trigger**: After admin verifies user's KYC
- **Notification Type**: `kyc_status`
- **Channel**: `email` + `sms`
- **Variables**:
  - `full_name`: User's full name
  - `status`: "approved"
- **Template**: `kyc_approved_email` / `kyc_approved_sms`

#### KYC Rejection
- **File**: `services/identity/internal/service/auth_service.go`
- **Function**: `RejectKYC()` (line 371)
- **Trigger**: After admin rejects user's KYC
- **Notification Type**: `kyc_status`
- **Channel**: `email` + `sms`
- **Variables**:
  - `full_name`: User's full name
  - `status`: "rejected"
  - `reason`: Rejection reason
- **Template**: `kyc_rejected_email` / `kyc_rejected_sms`

### 2. Wallet Service (`services/wallet`)

#### Wallet Creation
- **File**: `services/wallet/internal/service/wallet_service.go`
- **Function**: `CreateWallet()` (line 38)
- **Trigger**: After successful wallet creation
- **Notification Type**: `wallet_created`
- **Channel**: `email` + `in_app`
- **Variables**:
  - `wallet_type`: Type of wallet (savings/current/fixed)
  - `currency`: Wallet currency
  - `wallet_id`: Wallet ID
- **Template**: `wallet_created_email` / `wallet_created_inapp`

#### Wallet Activation
- **File**: `services/wallet/internal/service/wallet_service.go`
- **Function**: `ActivateWallet()` (line 121)
- **Trigger**: After wallet is activated (post-KYC)
- **Notification Type**: `wallet_activated`
- **Channel**: `email` + `sms` + `in_app`
- **Variables**:
  - `wallet_type`: Type of wallet
  - `currency`: Wallet currency
  - `wallet_id`: Wallet ID
- **Template**: `wallet_activated_email` / `wallet_activated_sms` / `wallet_activated_inapp`

### 3. Transaction Service (`services/transaction`)

#### Transaction Created
- **File**: `services/transaction/internal/service/transaction_service.go`
- **Functions**: `CreateTransfer()`, `CreateDeposit()`, `CreateWithdrawal()`
- **Trigger**: After transaction is created
- **Notification Type**: `transaction_created`
- **Channel**: `email` + `sms` + `in_app`
- **Priority**: `high` (for large amounts), `normal` (for regular amounts)
- **Variables**:
  - `transaction_type`: Type (transfer/deposit/withdrawal)
  - `amount`: Transaction amount
  - `currency`: Currency
  - `transaction_id`: Transaction ID
  - `description`: Transaction description
- **Template**: `transaction_created_email` / `transaction_created_sms` / `transaction_created_inapp`

#### Transaction Completed
- **Trigger**: When transaction status changes to `completed`
- **Notification Type**: `transaction_completed`
- **Channel**: `email` + `sms` + `in_app`
- **Variables**:
  - `transaction_type`: Type
  - `amount`: Amount
  - `currency`: Currency
  - `transaction_id`: Transaction ID
- **Template**: `transaction_completed_email` / `transaction_completed_sms`

#### Transaction Failed
- **Trigger**: When transaction status changes to `failed`
- **Notification Type**: `transaction_failed`
- **Channel**: `email` + `sms`
- **Priority**: `high`
- **Variables**:
  - `transaction_type`: Type
  - `amount`: Amount
  - `currency`: Currency
  - `transaction_id`: Transaction ID
  - `failure_reason`: Reason for failure
- **Template**: `transaction_failed_email` / `transaction_failed_sms`

## Implementation Plan

### Phase 1: Create Notification Client
Create a reusable HTTP client for calling the notification service:
- File: `shared/clients/notification_client.go`
- Methods: `SendNotification()`, `SendBulkNotifications()`

### Phase 2: Update Service Constructors
Add notification client to each service's constructor:
- `NewAuthService()` in identity service
- `NewWalletService()` in wallet service
- `NewTransactionService()` in transaction service

### Phase 3: Add Notification Triggers
Add notification client calls at each integration point listed above.

### Phase 4: Create Additional Templates
Add missing templates to `services/notification/migrations/003_seed_templates.up.sql`:
- `welcome_email`, `welcome_sms`
- `kyc_approved_email`, `kyc_approved_sms`
- `kyc_rejected_email`, `kyc_rejected_sms`
- `wallet_created_email`, `wallet_created_inapp`
- `wallet_activated_email`, `wallet_activated_sms`, `wallet_activated_inapp`
- `transaction_created_email`, `transaction_created_sms`, `transaction_created_inapp`
- `transaction_completed_email`, `transaction_completed_sms`
- `transaction_failed_email`, `transaction_failed_sms`

### Phase 5: Testing
Test each integration point end-to-end:
1. Register user → Verify welcome notification sent
2. Verify KYC → Verify approval notification sent
3. Create wallet → Verify wallet created notification sent
4. Activate wallet → Verify activation notification sent
5. Create transaction → Verify transaction notification sent

## Configuration

Add notification service URL to each service's environment variables:
```env
NOTIFICATION_SERVICE_URL=http://notification-service:8087
```

## Error Handling

- Notification failures should NOT block primary operations
- Log notification errors but continue with business logic
- Consider retry logic for critical notifications (welcome, KYC status)
