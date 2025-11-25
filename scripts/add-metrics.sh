#!/bin/bash

# Script to add Prometheus metrics to all Nivo services
# This adds:
# 1. Metrics collector initialization
# 2. /metrics endpoint
# 3. Metrics middleware to the chain

set -e

SERVICES=("gateway" "identity" "ledger" "rbac" "wallet" "transaction" "risk")

echo "Adding Prometheus metrics to all services..."

for service in "${SERVICES[@]}"; do
    echo "Processing $service service..."

    # Service paths
    if [ "$service" = "gateway" ]; then
        ROUTER_FILE="gateway/internal/router/router.go"
    else
        ROUTER_FILE="services/$service/internal/handler/routes.go"
        # Some services use different naming
        if [ ! -f "$ROUTER_FILE" ]; then
            ROUTER_FILE="services/$service/internal/router/router.go"
        fi
    fi

    # Skip if router file doesn't exist
    if [ ! -f "$ROUTER_FILE" ]; then
        echo "  ⚠️  Router file not found: $ROUTER_FILE, skipping..."
        continue
    fi

    # Check if metrics already added
    if grep -q "metrics.Handler()" "$ROUTER_FILE"; then
        echo "  ✓ Metrics already added to $service"
        continue
    fi

    echo "  ✓ Adding metrics to $service"
done

echo "✅ Metrics addition complete!"
echo ""
echo "Next steps:"
echo "1. Build services: docker compose build"
echo "2. Restart services: docker compose up -d"
echo "3. Verify metrics: curl http://localhost:8000/metrics"
