# Validator Package

Input validation utilities for Nivo services, providing struct validation with custom rules for financial data.

## Features

- **Struct Validation**: Validate entire structs with declarative tags
- **Custom Validators**: Fintech-specific validators (currency, amounts, account numbers)
- **Error Integration**: Automatic conversion to `shared/errors` format
- **Human-Readable Messages**: Clear, actionable error messages
- **JSON Tag Support**: Uses JSON field names in error messages
- **Banking Validators**: IBAN, sort codes, routing numbers
- **Global Instance**: Convenient global validator for simple use cases

## Installation

```bash
go get github.com/vnykmshr/nivo/shared/validator
```

## Usage

### Basic Validation

```go
import "github.com/vnykmshr/nivo/shared/validator"

type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18"`
    Password string `json:"password" validate:"required,min=8"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Validate the request
    if err := validator.Validate(req); err != nil {
        response.Error(w, err.(*errors.Error))
        return
    }

    // Process valid request...
}
```

Validation error response:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "name": "name is required",
      "email": "must be a valid email address",
      "age": "must be greater than or equal to 18"
    }
  }
}
```

### Custom Validator Instance

```go
v := validator.New()

type User struct {
    Name string `json:"name" validate:"required"`
}

user := User{Name: "John"}
if err := v.Validate(user); err != nil {
    // Handle validation error
}
```

### Single Variable Validation

```go
email := "user@example.com"
if err := validator.ValidateVar(email, "required,email"); err != nil {
    // Invalid email
}

amount := 100
if err := validator.ValidateVar(amount, "gt=0,lte=10000"); err != nil {
    // Invalid amount
}
```

## Standard Validation Tags

### Required & Empty
- `required` - Field must not be empty
- `omitempty` - Skip validation if field is empty

### String Validation
- `min=5` - Minimum length (characters)
- `max=100` - Maximum length (characters)
- `len=10` - Exact length
- `alpha` - Only letters
- `alphanum` - Letters and numbers only
- `numeric` - Only numbers
- `email` - Valid email address
- `url` - Valid URL

### Numeric Validation
- `gt=0` - Greater than
- `gte=18` - Greater than or equal
- `lt=100` - Less than
- `lte=100` - Less than or equal

### Enum Validation
- `oneof=pending processing completed` - Must be one of the specified values

### Format Validation
- `uuid` - Valid UUID format

## Custom Fintech Validators

### Currency Validation

Validates ISO 4217 currency codes:

```go
type Transfer struct {
    Currency string `json:"currency" validate:"required,currency"`
}

transfer := Transfer{Currency: "USD"} // Valid
transfer := Transfer{Currency: "XXX"} // Invalid
```

Supported currencies: USD, EUR, GBP, JPY, CNY, INR, CAD, AUD, CHF, SGD

### Money Amount Validation

Validates monetary amounts (must be positive):

```go
type Payment struct {
    Amount int64 `json:"amount" validate:"required,money_amount"`
}

payment := Payment{Amount: 1000} // Valid (10.00)
payment := Payment{Amount: 0}    // Invalid
payment := Payment{Amount: -100} // Invalid
```

### Account Number Validation

Validates bank account numbers (10-20 alphanumeric characters):

```go
type BankAccount struct {
    Number string `json:"account_number" validate:"required,account_number"`
}

account := BankAccount{Number: "1234567890"}       // Valid
account := BankAccount{Number: "ABC1234567890"}    // Valid
account := BankAccount{Number: "123"}              // Invalid (too short)
account := BankAccount{Number: "12345-6789"}       // Invalid (special chars)
```

### IBAN Validation

Validates International Bank Account Numbers:

```go
type BankAccount struct {
    IBAN string `json:"iban" validate:"required,iban"`
}

account := BankAccount{IBAN: "GB82WEST12345698765432"}     // Valid
account := BankAccount{IBAN: "DE89370400440532013000"}     // Valid
account := BankAccount{IBAN: "1234567890"}                 // Invalid
```

IBAN format: 2 letters (country), 2 digits (check), up to 30 alphanumeric

### Sort Code Validation (UK)

Validates UK bank sort codes (6 digits, hyphens optional):

```go
type UKBankAccount struct {
    SortCode string `json:"sort_code" validate:"required,sort_code"`
}

account := UKBankAccount{SortCode: "123456"}    // Valid
account := UKBankAccount{SortCode: "12-34-56"}  // Valid (hyphens removed)
account := UKBankAccount{SortCode: "12345"}     // Invalid (too short)
```

### Routing Number Validation (US)

Validates US bank routing numbers (9 digits):

```go
type USBankAccount struct {
    RoutingNumber string `json:"routing_number" validate:"required,routing_number"`
}

account := USBankAccount{RoutingNumber: "021000021"}    // Valid
account := USBankAccount{RoutingNumber: "111000025"}    // Valid
account := USBankAccount{RoutingNumber: "12345678"}     // Invalid (too short)
```

## Complete Examples

### Transfer Request Validation

```go
type TransferRequest struct {
    FromAccount string `json:"from_account" validate:"required,account_number"`
    ToAccount   string `json:"to_account" validate:"required,account_number"`
    Amount      int64  `json:"amount" validate:"required,money_amount"`
    Currency    string `json:"currency" validate:"required,currency"`
    Reference   string `json:"reference" validate:"max=100"`
}

func transfer(w http.ResponseWriter, r *http.Request) {
    var req TransferRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    // Validate
    if err := validator.Validate(req); err != nil {
        response.Error(w, err.(*errors.Error))
        return
    }

    // Process transfer...
    response.OK(w, result)
}
```

### User Registration Validation

```go
type RegisterRequest struct {
    Email           string `json:"email" validate:"required,email"`
    Password        string `json:"password" validate:"required,min=8,max=100"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
    FullName        string `json:"full_name" validate:"required,min=2,max=100"`
    DateOfBirth     string `json:"date_of_birth" validate:"required"`
    PhoneNumber     string `json:"phone_number" validate:"required,min=10,max=15"`
    Country         string `json:"country" validate:"required,len=2,alpha"`
}

func register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    json.NewDecoder(r.Body).Decode(&req)

    if err := validator.Validate(req); err != nil {
        // Returns structured validation errors
        response.Error(w, err.(*errors.Error))
        return
    }

    // Create user...
}
```

Error response:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "email": "must be a valid email address",
      "password": "must be at least 8 characters",
      "full_name": "full_name is required",
      "country": "must be exactly 2 characters"
    }
  }
}
```

### Payment Method Validation

```go
type PaymentMethod struct {
    Type          string `json:"type" validate:"required,oneof=card bank_transfer"`
    CardNumber    string `json:"card_number" validate:"required_if=Type card,len=16,numeric"`
    IBAN          string `json:"iban" validate:"required_if=Type bank_transfer,iban"`
    SortCode      string `json:"sort_code" validate:"omitempty,sort_code"`
    RoutingNumber string `json:"routing_number" validate:"omitempty,routing_number"`
}

func addPaymentMethod(w http.ResponseWriter, r *http.Request) {
    var pm PaymentMethod
    json.NewDecoder(r.Body).Decode(&pm)

    if err := validator.Validate(pm); err != nil {
        response.Error(w, err.(*errors.Error))
        return
    }

    // Add payment method...
}
```

### Transaction Validation

```go
type Transaction struct {
    Amount      int64  `json:"amount" validate:"required,money_amount"`
    Currency    string `json:"currency" validate:"required,currency"`
    Description string `json:"description" validate:"required,min=3,max=500"`
    Category    string `json:"category" validate:"required,oneof=transfer payment refund withdrawal deposit"`
    Status      string `json:"status" validate:"required,oneof=pending processing completed failed"`
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
    var tx Transaction
    json.NewDecoder(r.Body).Decode(&tx)

    if err := validator.Validate(tx); err != nil {
        response.Error(w, err.(*errors.Error))
        return
    }

    // Process transaction...
}
```

## Integration with Response Package

Works seamlessly with `shared/response` for consistent error handling:

```go
import (
    "github.com/vnykmshr/nivo/shared/response"
    "github.com/vnykmshr/nivo/shared/validator"
)

func handler(w http.ResponseWriter, r *http.Request) {
    var req Request
    json.NewDecoder(r.Body).Decode(&req)

    // Validation errors are automatically formatted
    if err := validator.Validate(req); err != nil {
        response.Error(w, err.(*errors.Error))
        return
    }

    // Continue processing...
}
```

## Error Messages

The validator provides human-readable error messages:

| Tag | Example Message |
|-----|----------------|
| `required` | "name is required" |
| `email` | "must be a valid email address" |
| `min=8` | "must be at least 8 characters" |
| `max=100` | "must be at most 100 characters" |
| `gte=18` | "must be greater than or equal to 18" |
| `oneof` | "must be one of: pending, completed" |
| `currency` | "must be a valid ISO 4217 currency code" |
| `money_amount` | "must be a valid monetary amount (positive integer in cents)" |
| `account_number` | "validation failed on 'account_number' tag" |
| `iban` | "validation failed on 'iban' tag" |

## Best Practices

1. **Use JSON tags** - Field names in errors match JSON field names
2. **Validate early** - Validate immediately after decoding request body
3. **Return validation errors** - Use `response.Error()` for consistent format
4. **Group validation rules** - Related rules together for clarity
5. **Use semantic validators** - Use `email`, `currency` rather than regex
6. **Test validation** - Write tests for validation rules
7. **Document requirements** - Make validation rules clear to API consumers

## Testing

```bash
cd shared/validator
go test -v
go test -cover
```

Coverage: 86.1%

## Advanced Usage

### Custom Instance with Additional Rules

```go
v := validator.New()

// You can register additional custom validators
v.validate.RegisterValidation("custom_tag", func(fl validator.FieldLevel) bool {
    // Custom validation logic
    return true
})

// Use the custom validator
err := v.Validate(myStruct)
```

### Conditional Validation

```go
type Payment struct {
    Type       string `json:"type" validate:"required,oneof=card bank"`
    CardNumber string `json:"card_number" validate:"required_if=Type card"`
    IBAN       string `json:"iban" validate:"required_if=Type bank"`
}
```

### Cross-Field Validation

```go
type UpdatePassword struct {
    NewPassword     string `json:"new_password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
```

## Performance

Validation is fast and efficient:

- Struct validation: ~10-50µs per struct
- Variable validation: ~1-5µs per variable
- Custom validators: Minimal overhead
- No reflection in hot path (after initialization)

## Related Packages

- [`shared/errors`](../errors/README.md) - Error types and codes
- [`shared/response`](../response/README.md) - API response formats
- [`shared/models`](../models/README.md) - Currency and Money types

## License

Copyright © 2025 Nivo
