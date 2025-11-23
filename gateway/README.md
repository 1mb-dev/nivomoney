# Nivo API Gateway

The API Gateway serves as the unified entry point for all Nivo microservices, handling request routing, authentication, rate limiting, and observability.

## Features

### Core Functionality
- **Request Routing**: Routes requests to appropriate backend services based on URL patterns
- **JWT Authentication**: Validates JWT tokens locally without calling external services (fast!)
- **Service Discovery**: Static configuration via environment variables
- **Reverse Proxy**: Transparent proxying to backend services

### Security
- **Local JWT Validation**: Validates tokens using shared secret (no network latency)
- **Permission Checking**: Extracts user roles and permissions from JWT claims
- **Rate Limiting**: Gateway-wide rate limiting to prevent abuse
- **CORS**: Configurable CORS policies

### Observability
- **Request Logging**: Logs all incoming requests with method, path, status
- **Request ID Propagation**: Generates and propagates request IDs for distributed tracing
- **Health Checks**: Gateway-level health endpoint
- **Error Handling**: Standardized error responses

### Middleware Chain
1. Recovery (panic handling)
2. Logging (request/response logging)
3. Request ID generation
4. CORS headers
5. Rate limiting
6. JWT authentication (for protected routes)

## Architecture

```
Client → Gateway (Port 8000)
         ├── /api/v1/identity/* → Identity Service (8080)
         ├── /api/v1/ledger/* → Ledger Service (8081)
         ├── /api/v1/rbac/* → RBAC Service (8082)
         ├── /api/v1/transaction/* → Transaction Service (8083)
         └── /api/v1/wallet/* → Wallet Service (8084)
```

## Configuration

The gateway is configured via environment variables:

```bash
# Gateway port
SERVICE_PORT=8000

# JWT secret (must match Identity service)
JWT_SECRET=your-super-secret-jwt-key

# Backend service URLs
IDENTITY_SERVICE_URL=http://identity-service:8080
LEDGER_SERVICE_URL=http://ledger-service:8081
RBAC_SERVICE_URL=http://rbac-service:8082
TRANSACTION_SERVICE_URL=http://transaction-service:8083
WALLET_SERVICE_URL=http://wallet-service:8084
```

## Usage

### Local Development

```bash
# Copy environment variables
cp .env.example .env

# Run the gateway
make run

# Or build and run
make build
./bin/gateway
```

### Docker

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

### Testing

```bash
# Health check
curl http://localhost:8000/health

# Register a new user (public endpoint, no auth required)
curl -X POST http://localhost:8000/api/v1/identity/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "full_name": "John Doe",
    "phone": "+91-9876543210"
  }'

# Login (public endpoint, returns JWT)
curl -X POST http://localhost:8000/api/v1/identity/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'

# Get user profile (protected endpoint, requires JWT)
curl http://localhost:8000/api/v1/identity/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Create wallet (protected endpoint)
curl -X POST http://localhost:8000/api/v1/wallet/wallets \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "SAVINGS",
    "currency": "INR"
  }'
```

## Route Patterns

### Public Routes (No Authentication)
- `POST /api/v1/identity/auth/register` - User registration
- `POST /api/v1/identity/auth/login` - User login
- `GET /health` - Gateway health check

### Protected Routes (JWT Required)
All other `/api/v1/*` routes require JWT authentication in the `Authorization: Bearer <token>` header.

## JWT Validation

The gateway validates JWTs **locally** (no network calls to Identity service):

1. Extracts token from `Authorization: Bearer <token>` header
2. Validates signature using shared JWT secret
3. Checks token expiration
4. Extracts user claims (user_id, email, roles, permissions)
5. Adds user context to request headers (`X-User-ID`, `X-User-Email`)
6. Forwards request to backend service

### Why Local Validation?
- ✅ **Fast**: No network latency (microseconds vs milliseconds)
- ✅ **Resilient**: Works even if Identity service is down
- ✅ **Scalable**: No load on Identity service for every request
- ✅ **Standard**: How JWTs are designed to work

## Request Flow

```
1. Client sends request to gateway
2. Gateway applies middleware chain:
   - Recovery (panic handling)
   - Logging
   - Request ID generation
   - CORS headers
   - Rate limiting
   - JWT validation (if protected route)
3. Gateway proxies request to backend service:
   - Adds X-Forwarded headers
   - Adds X-User-ID header (from JWT)
   - Preserves original path
4. Backend service processes request
5. Gateway returns response to client
```

## Error Handling

The gateway returns standardized error responses:

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid or expired token",
    "details": null
  }
}
```

Common error codes:
- `UNAUTHORIZED` (401): Missing or invalid JWT
- `FORBIDDEN` (403): Insufficient permissions
- `SERVICE_NOT_FOUND` (404): Unknown service in path
- `BAD_GATEWAY` (502): Backend service unavailable
- `SERVICE_UNAVAILABLE` (503): Rate limit exceeded

## Development

### Adding a New Service

1. Update `internal/proxy/registry.go` to add service URL
2. No route changes needed - all `/api/v1/{service}/*` routes are proxied automatically

### Adding Custom Routes

For routes that need special handling, add them to `internal/router/router.go`:

```go
// Example: Custom route for a specific endpoint
mux.HandleFunc("POST /api/v1/special/endpoint", r.customHandler)
```

## Performance

- Request latency: ~1-2ms (gateway overhead)
- JWT validation: <1ms (local validation)
- Throughput: Tested up to 10,000 req/s on standard hardware

## Future Enhancements

- [ ] Health check aggregation (ping all backend services)
- [ ] Circuit breaker for backend service failures
- [ ] Request/response caching
- [ ] Advanced metrics (Prometheus)
- [ ] OpenAPI documentation aggregation
- [ ] WebSocket support
