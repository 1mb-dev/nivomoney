# Models Package

Common domain types for Nivo services.

## Overview

The `models` package provides fundamental domain types used across all Nivo services, with a focus on fintech-specific types like Money and Currency that require precision and correctness.

## Features

- **Money**: Precise monetary amounts using integer arithmetic (no float precision issues)
- **Currency**: ISO 4217 currency codes with validation
- **Timestamp**: Custom timestamp type with consistent JSON/database serialization

## Money Type

The `Money` type stores monetary amounts in the smallest currency unit (cents) to avoid floating-point precision issues.

### Basic Usage

```go
import "github.com/vnykmshr/nivo/shared/models"

// Create from cents
money := models.NewMoney(1050, models.USD) // $10.50

// Create from float
money := models.NewMoneyFromFloat(10.50, models.USD)

// String representation
fmt.Println(money) // Output: "10.50 USD"

// Convert to float
amount := money.ToFloat() // 10.50
```

### Arithmetic Operations

```go
price := models.NewMoney(1000, models.USD)    // $10.00
tax := models.NewMoney(150, models.USD)       // $1.50

// Addition
total, err := price.Add(tax) // $11.50
if err != nil {
    // Handle currency mismatch error
}

// Subtraction
discount := models.NewMoney(200, models.USD)
final, err := total.Subtract(discount) // $9.50

// Multiplication
double := price.Multiply(2) // $20.00

// Division
half := price.Divide(2) // $5.00
```

### Comparisons

```go
balance := models.NewMoney(10000, models.USD)  // $100.00
amount := models.NewMoney(5000, models.USD)    // $50.00

if balance.GreaterThan(amount) {
    fmt.Println("Sufficient funds")
}

if amount.LessThanOrEqual(balance) {
    // Process transaction
}

if money1.Equal(money2) {
    // Same amount and currency
}
```

### Validation

```go
money := models.NewMoney(1000, models.USD)

if money.IsZero() {
    // Handle zero amount
}

if money.IsPositive() {
    // Process positive amount
}

if money.IsNegative() {
    // Handle negative balance
}

// Validate currency
if err := money.Validate(); err != nil {
    // Handle invalid currency
}
```

### JSON Serialization

```go
money := models.NewMoney(1050, models.USD)

// Marshal to JSON
data, _ := json.Marshal(money)
// {"amount":1050,"currency":"USD"}

// Unmarshal from JSON
var decoded models.Money
json.Unmarshal(data, &decoded)
```

## Currency Type

The `Currency` type represents ISO 4217 currency codes.

### Supported Currencies

- USD - US Dollar ($)
- EUR - Euro (€)
- GBP - British Pound (£)
- JPY - Japanese Yen (¥)
- CNY - Chinese Yuan (¥)
- INR - Indian Rupee (₹)
- CAD - Canadian Dollar (C$)
- AUD - Australian Dollar (A$)
- CHF - Swiss Franc (CHF)
- SGD - Singapore Dollar (S$)

### Usage

```go
// Using constants
currency := models.USD

// Parse from string
currency, err := models.ParseCurrency("USD")
if err != nil {
    // Handle invalid currency
}

// Case-insensitive parsing
currency, _ := models.ParseCurrency("usd") // Returns USD

// Validation
if err := currency.Validate(); err != nil {
    // Handle invalid currency
}

// Check support
if currency.IsSupported() {
    // Currency is valid
}

// Get all supported currencies
currencies := models.GetSupportedCurrencies()
```

### Currency Properties

```go
currency := models.USD

// Get symbol
symbol := currency.GetSymbol() // "$"

// Get decimal places
places := currency.GetDecimalPlaces() // 2 (most currencies)
// Note: JPY returns 0 (no decimal places)

// String representation
name := currency.String() // "USD"
```

## Timestamp Type

Custom timestamp type with consistent JSON and database serialization.

### Usage

```go
// Create from time.Time
t := time.Now()
ts := models.NewTimestamp(t)

// Get current timestamp
ts := models.Now()

// String representation (ISO 8601)
str := ts.String() // "2025-01-15T10:30:00Z"
```

### Comparisons

```go
ts1 := models.Now()
time.Sleep(1 * time.Second)
ts2 := models.Now()

if ts2.After(ts1) {
    fmt.Println("ts2 is after ts1")
}

if ts1.Before(ts2) {
    fmt.Println("ts1 is before ts2")
}

if ts1.Equal(ts1) {
    fmt.Println("Equal timestamps")
}
```

### JSON Serialization

```go
ts := models.Now()

// Marshal to JSON (ISO 8601 format)
data, _ := json.Marshal(ts)
// "2025-01-15T10:30:00Z"

// Unmarshal from JSON
var decoded models.Timestamp
json.Unmarshal(data, &decoded)

// Zero timestamps serialize as null
zero := models.Timestamp{}
data, _ := json.Marshal(zero) // null
```

### Database Integration

The Timestamp type implements `sql.Scanner` and `driver.Valuer` for seamless database integration:

```go
// Scan from database
var ts models.Timestamp
err := db.QueryRow("SELECT created_at FROM users WHERE id = $1", userID).Scan(&ts)

// Store in database
_, err := db.Exec("INSERT INTO users (name, created_at) VALUES ($1, $2)",
    name, models.Now())
```

## Complete Example: Transfer

```go
package main

import (
    "fmt"
    "github.com/vnykmshr/nivo/shared/models"
)

func Transfer(from, to models.Money, amount models.Money) error {
    // Validate currencies match
    if from.Currency != amount.Currency {
        return fmt.Errorf("currency mismatch")
    }

    // Check sufficient funds
    if from.LessThan(amount) {
        return fmt.Errorf("insufficient funds")
    }

    // Validate amount is positive
    if !amount.IsPositive() {
        return fmt.Errorf("amount must be positive")
    }

    // Perform transfer
    newFrom, _ := from.Subtract(amount)
    newTo, _ := to.Add(amount)

    fmt.Printf("From: %s -> %s\n", from, newFrom)
    fmt.Printf("To:   %s -> %s\n", to, newTo)

    return nil
}

func main() {
    sender := models.NewMoney(10000, models.USD)     // $100.00
    receiver := models.NewMoney(5000, models.USD)    // $50.00
    amount := models.NewMoney(2500, models.USD)      // $25.00

    if err := Transfer(sender, receiver, amount); err != nil {
        fmt.Printf("Transfer failed: %v\n", err)
        return
    }

    fmt.Println("Transfer successful!")
}
```

## Best Practices

### Money

1. **Always use integer storage**: Store amounts in cents to avoid float precision issues
   ```go
   // Good
   money := models.NewMoney(1050, models.USD)

   // Avoid
   amount := 10.50 // float64 has precision issues
   ```

2. **Validate currency compatibility**: Always check currencies match before operations
   ```go
   result, err := money1.Add(money2)
   if err != nil {
       // Handle currency mismatch
   }
   ```

3. **Use comparison methods**: Don't compare amounts directly
   ```go
   // Good
   if balance.GreaterThan(amount) { }

   // Avoid
   if balance.Amount > amount.Amount { } // Ignores currency
   ```

4. **Handle zero division**: Check divisor before division
   ```go
   if divisor != 0 {
       result := money.Divide(divisor)
   }
   ```

### Currency

1. **Use constants**: Prefer predefined currency constants
   ```go
   // Good
   currency := models.USD

   // Less ideal
   currency := models.Currency("USD")
   ```

2. **Always validate**: Validate currency before use
   ```go
   currency, err := models.ParseCurrency(input)
   if err != nil {
       return err
   }
   ```

### Timestamp

1. **Use UTC**: Always work with UTC timestamps
   ```go
   ts := models.NewTimestamp(time.Now().UTC())
   ```

2. **Check for zero**: Check if timestamp is set before use
   ```go
   if !ts.IsZero() {
       // Use timestamp
   }
   ```

## Testing

```bash
go test ./shared/models/...
go test -cover ./shared/models/...
```

## Related Packages

- [shared/errors](../errors/README.md) - For validation errors
- [shared/database](../database/README.md) - For database persistence
