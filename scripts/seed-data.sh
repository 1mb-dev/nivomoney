#!/bin/bash

# Nivo Wallet - Seed Data Script
# This script populates the system with realistic test data using different personas

set -e

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8000/api/v1}"
ADMIN_TOKEN=""
USER_TOKENS=()

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Wait for services to be ready
wait_for_services() {
    log_info "Waiting for services to be ready..."
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$API_BASE_URL/../health" > /dev/null 2>&1; then
            log_success "Services are ready!"
            return 0
        fi
        log_info "Attempt $attempt/$max_attempts - Services not ready yet..."
        sleep 2
        ((attempt++))
    done

    log_error "Services did not become ready in time"
    exit 1
}

# Create a user and return the auth token
create_user() {
    local name="$1"
    local email="$2"
    local password="$3"
    local phone="$4"

    log_info "Creating user: $name ($email)"

    # Register the user
    local reg_response=$(curl -s -X POST "$API_BASE_URL/identity/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"full_name\": \"$name\",
            \"email\": \"$email\",
            \"password\": \"$password\",
            \"phone\": \"$phone\"
        }")

    local user_id=$(echo "$reg_response" | jq -r '.data.id // empty')
    local error_code=$(echo "$reg_response" | jq -r '.error.code // empty')

    # If user already exists, try to login
    if [ "$error_code" = "CONFLICT" ]; then
        log_warn "User $name already exists, attempting login..."
    elif [ -z "$user_id" ] || [ "$user_id" = "null" ]; then
        log_error "Failed to register user $name: $reg_response"
        return 1
    else
        log_success "Registered user: $name (ID: $user_id)"
    fi

    # Login to get the token
    local login_response=$(curl -s -X POST "$API_BASE_URL/identity/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$email\",
            \"password\": \"$password\"
        }")

    local token=$(echo "$login_response" | jq -r '.data.token // empty')
    user_id=$(echo "$login_response" | jq -r '.data.user.id // empty')

    if [ -z "$token" ] || [ "$token" = "null" ]; then
        log_error "Failed to login user $name: $login_response"
        return 1
    fi

    log_success "Logged in user: $name (ID: $user_id)"
    echo "$token|$user_id"
}

# Assign role to user
assign_role() {
    local user_id="$1"
    local role_name="$2"
    local admin_token="$3"

    log_info "Assigning role '$role_name' to user $user_id"

    # First, get the role ID
    local role_id=$(curl -s -X GET "$API_BASE_URL/rbac/roles" \
        -H "Authorization: Bearer $admin_token" \
        | jq -r ".roles[] | select(.name == \"$role_name\") | .id // empty")

    if [ -z "$role_id" ] || [ "$role_id" = "null" ]; then
        log_warn "Role '$role_name' not found, skipping assignment"
        return 0
    fi

    local response=$(curl -s -X POST "$API_BASE_URL/rbac/users/$user_id/roles" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "{
            \"role_id\": \"$role_id\"
        }")

    log_success "Assigned role '$role_name' to user $user_id"
}

# Create a wallet
create_wallet() {
    local user_id="$1"
    local currency="$2"
    local token="$3"

    log_info "Creating $currency wallet for user $user_id"

    local response=$(curl -s -X POST "$API_BASE_URL/wallet/wallets" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "{
            \"user_id\": \"$user_id\",
            \"currency\": \"$currency\"
        }")

    local wallet_id=$(echo "$response" | jq -r '.data.id // empty')

    if [ -z "$wallet_id" ] || [ "$wallet_id" = "null" ]; then
        log_error "Failed to create wallet: $response"
        return 1
    fi

    log_success "Created $currency wallet: $wallet_id"
    echo "$wallet_id"
}

# Make a deposit
make_deposit() {
    local wallet_id="$1"
    local amount="$2"
    local token="$3"
    local description="${4:-Initial deposit}"

    log_info "Depositing ₹$amount to wallet $wallet_id"

    local response=$(curl -s -X POST "$API_BASE_URL/transaction/transactions/deposit" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "{
            \"wallet_id\": \"$wallet_id\",
            \"amount_paise\": $amount,
            \"description\": \"$description\"
        }")

    local txn_id=$(echo "$response" | jq -r '.data.id // empty')

    if [ -z "$txn_id" ] || [ "$txn_id" = "null" ]; then
        log_error "Failed to make deposit: $response"
        return 1
    fi

    log_success "Deposit completed: ₹$amount (Transaction: $txn_id)"
    echo "$txn_id"
}

# Make a transfer
make_transfer() {
    local from_wallet="$1"
    local to_wallet="$2"
    local amount="$3"
    local token="$4"
    local description="${5:-Transfer}"

    log_info "Transferring ₹$amount from $from_wallet to $to_wallet"

    local response=$(curl -s -X POST "$API_BASE_URL/transaction/transactions/transfer" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "{
            \"source_wallet_id\": \"$from_wallet\",
            \"destination_wallet_id\": \"$to_wallet\",
            \"amount_paise\": $amount,
            \"currency\": \"INR\",
            \"description\": \"$description\"
        }")

    local txn_id=$(echo "$response" | jq -r '.data.id // empty')

    if [ -z "$txn_id" ] || [ "$txn_id" = "null" ]; then
        log_error "Failed to make transfer: $response"
        return 1
    fi

    log_success "Transfer completed: ₹$amount (Transaction: $txn_id)"
    echo "$txn_id"
}

# Main execution
main() {
    echo "========================================"
    echo "   Nivo Wallet - Seed Data Script"
    echo "========================================"
    echo ""

    # Wait for services
    wait_for_services
    echo ""

    # ========================================
    # PERSONA 1: System Administrator
    # ========================================
    log_info "========== Creating System Administrator =========="
    admin_data=$(create_user "Admin User" "admin@nivo.local" "admin123" "+919876543210")
    admin_token=$(echo "$admin_data" | cut -d'|' -f1)
    admin_id=$(echo "$admin_data" | cut -d'|' -f2)
    admin_wallet=$(create_wallet "$admin_id" "INR" "$admin_token")
    echo ""

    # ========================================
    # PERSONA 2: Regular User - Raj Kumar
    # ========================================
    log_info "========== Creating Regular User: Raj Kumar =========="
    raj_data=$(create_user "Raj Kumar" "raj.kumar@gmail.com" "raj123" "+919876543211")
    raj_token=$(echo "$raj_data" | cut -d'|' -f1)
    raj_id=$(echo "$raj_data" | cut -d'|' -f2)
    raj_wallet=$(create_wallet "$raj_id" "INR" "$raj_token")
    make_deposit "$raj_wallet" 50000 "$raj_token" "Initial balance - Salary credit"
    echo ""

    # ========================================
    # PERSONA 3: Merchant - Priya's Electronics
    # ========================================
    log_info "========== Creating Merchant: Priya's Electronics =========="
    priya_data=$(create_user "Priya Sharma" "priya.electronics@business.com" "priya123" "+919876543212")
    priya_token=$(echo "$priya_data" | cut -d'|' -f1)
    priya_id=$(echo "$priya_data" | cut -d'|' -f2)
    priya_wallet=$(create_wallet "$priya_id" "INR" "$priya_token")
    make_deposit "$priya_wallet" 100000 "$priya_token" "Business account opening balance"
    echo ""

    # ========================================
    # PERSONA 4: Freelancer - Arjun Designer
    # ========================================
    log_info "========== Creating Freelancer: Arjun Designer =========="
    arjun_data=$(create_user "Arjun Patel" "arjun.design@freelance.com" "arjun123" "+919876543213")
    arjun_token=$(echo "$arjun_data" | cut -d'|' -f1)
    arjun_id=$(echo "$arjun_data" | cut -d'|' -f2)
    arjun_wallet=$(create_wallet "$arjun_id" "INR" "$arjun_token")
    make_deposit "$arjun_wallet" 25000 "$arjun_token" "Freelance payment received"
    echo ""

    # ========================================
    # PERSONA 5: Student - Neha Student
    # ========================================
    log_info "========== Creating Student: Neha =========="
    neha_data=$(create_user "Neha Singh" "neha.singh@student.com" "neha123" "+919876543214")
    neha_token=$(echo "$neha_data" | cut -d'|' -f1)
    neha_id=$(echo "$neha_data" | cut -d'|' -f2)
    neha_wallet=$(create_wallet "$neha_id" "INR" "$neha_token")
    make_deposit "$neha_wallet" 5000 "$neha_token" "Pocket money from parents"
    echo ""

    # ========================================
    # PERSONA 6: Premium User - Vikram Business
    # ========================================
    log_info "========== Creating Premium User: Vikram =========="
    vikram_data=$(create_user "Vikram Malhotra" "vikram.m@corporate.com" "vikram123" "+919876543215")
    vikram_token=$(echo "$vikram_data" | cut -d'|' -f1)
    vikram_id=$(echo "$vikram_data" | cut -d'|' -f2)
    vikram_wallet=$(create_wallet "$vikram_id" "INR" "$vikram_token")
    make_deposit "$vikram_wallet" 500000 "$vikram_token" "Corporate account funding"
    echo ""

    # ========================================
    # Create Realistic Transactions
    # ========================================
    log_info "========== Creating Realistic Transactions =========="
    echo ""

    # Scenario 1: Raj buys electronics from Priya's store
    log_info "Scenario 1: Raj purchases laptop from Priya's Electronics"
    make_transfer "$raj_wallet" "$priya_wallet" 45000 "$raj_token" "Laptop purchase - Dell Inspiron"
    sleep 1

    # Scenario 2: Vikram hires Arjun for design work
    log_info "Scenario 2: Vikram pays Arjun for website design"
    make_transfer "$vikram_wallet" "$arjun_wallet" 75000 "$vikram_token" "Website redesign project payment"
    sleep 1

    # Scenario 3: Priya orders supplies
    log_info "Scenario 3: Priya pays supplier"
    make_transfer "$priya_wallet" "$vikram_wallet" 30000 "$priya_token" "Electronics components purchase"
    sleep 1

    # Scenario 4: Neha receives scholarship
    log_info "Scenario 4: Neha receives scholarship payment"
    make_deposit "$neha_wallet" 15000 "$neha_token" "University scholarship"
    sleep 1

    # Scenario 5: Neha pays Arjun for tutoring
    log_info "Scenario 5: Neha pays Arjun for graphic design tutoring"
    make_transfer "$neha_wallet" "$arjun_wallet" 5000 "$neha_token" "3 graphic design lessons"
    sleep 1

    # Scenario 6: Raj sends money to Neha (sibling transfer)
    log_info "Scenario 6: Raj sends money to Neha"
    make_transfer "$raj_wallet" "$neha_wallet" 3000 "$raj_token" "Monthly allowance"
    sleep 1

    # Scenario 7: Vikram makes bulk purchase from Priya
    log_info "Scenario 7: Vikram bulk purchases office equipment"
    make_transfer "$vikram_wallet" "$priya_wallet" 150000 "$vikram_token" "50 laptops for office"
    sleep 1

    # Scenario 8: More deposits for active users
    log_info "Scenario 8: Regular income/deposits"
    make_deposit "$arjun_wallet" 35000 "$arjun_token" "Client payment - Logo design"
    make_deposit "$priya_wallet" 80000 "$priya_token" "Product sales revenue"
    sleep 1

    echo ""
    log_info "========== Seed Data Summary =========="
    echo ""
    echo "Created 6 user personas:"
    echo "  1. Admin User (admin@nivo.local) - System Administrator"
    echo "  2. Raj Kumar (raj.kumar@gmail.com) - Regular User"
    echo "  3. Priya Sharma (priya.electronics@business.com) - Merchant"
    echo "  4. Arjun Patel (arjun.design@freelance.com) - Freelancer"
    echo "  5. Neha Singh (neha.singh@student.com) - Student"
    echo "  6. Vikram Malhotra (vikram.m@corporate.com) - Premium User"
    echo ""
    echo "All users have password: [respective_name]123"
    echo "Example: admin123, raj123, priya123, etc."
    echo ""
    echo "Created wallets and transactions representing realistic usage patterns:"
    echo "  - E-commerce purchases"
    echo "  - Freelance payments"
    echo "  - B2B transactions"
    echo "  - Personal transfers"
    echo "  - Deposits and withdrawals"
    echo ""
    log_success "Seed data creation completed successfully!"
    echo ""
    echo "You can now:"
    echo "  - Login with any of the created users"
    echo "  - View transactions and wallet balances"
    echo "  - Test the API with realistic data"
    echo ""
    echo "Gateway URL: $API_BASE_URL"
    echo "========================================"
}

# Run main function
main "$@"
