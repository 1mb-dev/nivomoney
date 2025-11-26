# Notification Service

The Notification Service is a **simulation engine** that handles all outbound notifications in Nivo Money. It simulates real-world notification delivery without actually sending messages to external providers.

## Features

- **Multi-channel Support**: SMS, Email, Push Notifications, In-App Messages
- **Template System**: Reusable templates with variable substitution ({{variable}})
- **Simulation Engine**: Realistic delivery simulation with configurable delays and failure rates
- **Lifecycle Tracking**: Queued → Sent → Delivered/Failed with timestamps
- **Retry Logic**: Automatic retries with exponential backoff
- **Priority Handling**: Critical (OTP) messages processed first
- **Idempotency**: Prevents duplicate notifications using correlation_id
- **Admin Dashboard**: View, filter, and replay notifications
- **Statistics**: Success rate, channel breakdown, type distribution

## Architecture

### Components

1. **Notification Repository**: Database operations for notifications
2. **Template Repository**: Template CRUD and retrieval
3. **Template Engine**: Variable substitution ({{var}} → value)
4. **Simulation Engine**: Mimics real notification delivery
5. **Background Worker**: Processes queued notifications asynchronously

### Database Schema

**notifications** table:
- Stores all notification attempts
- Tracks lifecycle status and timestamps
- Supports idempotency via correlation_id
- Indexed for efficient queries

**notification_templates** table:
- Reusable templates with placeholders
- Versioning support
- Channel-specific templates

## API Endpoints

### Notifications

- `POST /v1/notifications/send` - Send a notification
- `GET /v1/notifications/{id}` - Get notification details
- `GET /v1/notifications` - List notifications with filters

### Templates

- `POST /v1/templates` - Create a template
- `GET /v1/templates/{id}` - Get template
- `GET /v1/templates` - List all templates
- `PUT /v1/templates/{id}` - Update template
- `POST /v1/templates/{id}/preview` - Preview with variables

### Admin (RBAC Protected)

- `GET /admin/notifications/stats` - Get statistics
- `POST /admin/notifications/{id}/replay` - Replay notification

## Configuration

Environment variables:

```bash
# Service Configuration
SERVICE_PORT=8087
ENVIRONMENT=development
DATABASE_URL=postgres://...
MIGRATIONS_DIR=./migrations

# Simulation Engine Configuration
SIM_DELIVERY_DELAY_MS=1000          # Delay before marking as 'sent'
SIM_FINAL_DELAY_MS=2000             # Delay before final status
SIM_FAILURE_RATE_PERCENT=10.0       # Percentage that fail (0-100)
SIM_MAX_RETRY_ATTEMPTS=3            # Max retries
SIM_RETRY_DELAY_MS=2000             # Base retry delay
```

## Usage Examples

### Send OTP via SMS

```bash
curl -X POST http://localhost:8087/v1/notifications/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "sms",
    "type": "otp",
    "priority": "critical",
    "recipient": "+919876543210",
    "template_id": "<otp_sms_template_id>",
    "variables": {
      "otp": "123456",
      "validity_minutes": "10"
    },
    "correlation_id": "txn-123-otp"
  }'
```

### Send Transaction Alert Email

```bash
curl -X POST http://localhost:8087/v1/notifications/send \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "email",
    "type": "transaction_alert",
    "priority": "high",
    "recipient": "user@example.com",
    "template_id": "<transaction_alert_email_template_id>",
    "variables": {
      "user_name": "John Doe",
      "transaction_type": "Credit",
      "amount": "5000",
      "date": "2025-11-26",
      "transaction_id": "TXN123456",
      "description": "Payment received",
      "balance": "15000"
    }
  }'
```

### List Notifications with Filters

```bash
# Get all failed SMS notifications
curl "http://localhost:8087/v1/notifications?channel=sms&status=failed&limit=50"

# Get notifications for a specific user
curl "http://localhost:8087/v1/notifications?user_id=<uuid>&limit=20&offset=0"
```

### Get Statistics

```bash
curl http://localhost:8087/admin/notifications/stats
```

Response:
```json
{
  "total_notifications": 1250,
  "by_channel": {
    "sms": 500,
    "email": 450,
    "push": 200,
    "in_app": 100
  },
  "by_status": {
    "queued": 50,
    "sent": 100,
    "delivered": 1000,
    "failed": 100
  },
  "by_type": {
    "otp": 300,
    "transaction_alert": 600,
    "account_alert": 200,
    "welcome": 150
  },
  "success_rate": 90.91,
  "average_retries": 0.25
}
```

## Default Templates

10 pre-seeded templates:

1. `otp_sms` - OTP via SMS
2. `transaction_alert_sms` - Transaction alerts
3. `welcome_email` - Welcome email
4. `transaction_alert_email` - Transaction email
5. `kyc_update_email` - KYC status updates
6. `account_alert_email` - Account alerts
7. `wallet_created_push` - Wallet creation push
8. `transaction_alert_push` - Transaction push
9. `security_alert_push` - Security alerts
10. `welcome_inapp` - Welcome in-app message

## Simulation Behavior

### Status Lifecycle

1. **Queued** (on creation)
   - Notification saved to database
   - Returns notification_id immediately

2. **Sent** (after SIM_DELIVERY_DELAY_MS)
   - Worker picks up notification
   - Simulates network delay
   - Marks as 'sent'

3. **Delivered/Failed** (after SIM_FINAL_DELAY_MS)
   - Random determination based on failure rate
   - Delivered: Success
   - Failed: Random failure reason, retry logic kicks in

### Failure Simulation

- Configurable failure rate (default 10%)
- Channel-specific failure reasons
- Retry logic with exponential backoff
- Max retry attempts (default 3)

## Development

```bash
# Run migrations
migrate -path ./migrations -database "postgres://..." up

# Run service
go run cmd/server/main.go

# Run with custom config
SIM_FAILURE_RATE_PERCENT=25 go run cmd/server/main.go
```

## Testing

```bash
# Run tests
go test ./...

# Test with coverage
go test -cover ./...
```

## Integration

Other services can send notifications by calling the notification service:

```go
// Example: Send OTP after registration
notificationReq := &NotificationRequest{
    Channel:    "sms",
    Type:       "otp",
    Priority:   "critical",
    Recipient:  user.Phone,
    TemplateID: otpTemplateID,
    Variables: map[string]interface{}{
        "otp": generatedOTP,
        "validity_minutes": "10",
    },
    CorrelationID: fmt.Sprintf("user-%s-otp", user.ID),
}
```

## Monitoring

- Health check: `GET /health`
- Ready check: `GET /ready`
- Statistics: `GET /admin/notifications/stats`

## Security

- Admin endpoints protected by RBAC
- Idempotency prevents duplicate sends
- Correlation ID tracking for audit
- All notifications logged and auditable
