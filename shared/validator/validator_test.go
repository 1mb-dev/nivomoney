package validator

import (
	"testing"

	"github.com/vnykmshr/nivo/shared/errors"
)

func TestValidate(t *testing.T) {
	v := New()

	t.Run("valid struct passes validation", func(t *testing.T) {
		type User struct {
			Name  string `json:"name" validate:"required,min=3"`
			Email string `json:"email" validate:"required,email"`
			Age   int    `json:"age" validate:"required,gte=18"`
		}

		user := User{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   25,
		}

		err := v.Validate(user)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		type User struct {
			Name  string `json:"name" validate:"required"`
			Email string `json:"email" validate:"required"`
		}

		user := User{
			Name: "John",
			// Email is missing
		}

		err := v.Validate(user)
		if err == nil {
			t.Fatal("expected validation error")
		}

		validationErr, ok := err.(*errors.Error)
		if !ok {
			t.Fatal("expected errors.Error type")
		}

		if validationErr.Code != errors.ErrCodeValidation {
			t.Error("expected validation error code")
		}

		if validationErr.Details["email"] == nil {
			t.Error("expected email field in details")
		}
	})

	t.Run("min length validation", func(t *testing.T) {
		type User struct {
			Name string `json:"name" validate:"min=5"`
		}

		user := User{Name: "Jo"}

		err := v.Validate(user)
		if err == nil {
			t.Fatal("expected validation error")
		}

		validationErr := err.(*errors.Error)
		if validationErr.Details["name"] == nil {
			t.Error("expected name field in details")
		}
	})

	t.Run("max length validation", func(t *testing.T) {
		type User struct {
			Name string `json:"name" validate:"max=10"`
		}

		user := User{Name: "This is a very long name"}

		err := v.Validate(user)
		if err == nil {
			t.Fatal("expected validation error")
		}

		validationErr := err.(*errors.Error)
		if validationErr.Details["name"] == nil {
			t.Error("expected name field in details")
		}
	})

	t.Run("email validation", func(t *testing.T) {
		type User struct {
			Email string `json:"email" validate:"email"`
		}

		user := User{Email: "invalid-email"}

		err := v.Validate(user)
		if err == nil {
			t.Fatal("expected validation error")
		}

		validationErr := err.(*errors.Error)
		if validationErr.Details["email"] == nil {
			t.Error("expected email field in details")
		}
	})

	t.Run("numeric comparison validation", func(t *testing.T) {
		type Product struct {
			Price    int `json:"price" validate:"gt=0"`
			Quantity int `json:"quantity" validate:"gte=1,lte=100"`
		}

		// Invalid: price is 0
		product := Product{Price: 0, Quantity: 5}
		err := v.Validate(product)
		if err == nil {
			t.Fatal("expected validation error for price")
		}

		// Invalid: quantity is 0
		product = Product{Price: 100, Quantity: 0}
		err = v.Validate(product)
		if err == nil {
			t.Fatal("expected validation error for quantity")
		}

		// Valid
		product = Product{Price: 100, Quantity: 50}
		err = v.Validate(product)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("oneof validation", func(t *testing.T) {
		type Order struct {
			Status string `json:"status" validate:"oneof=pending processing completed cancelled"`
		}

		order := Order{Status: "invalid"}

		err := v.Validate(order)
		if err == nil {
			t.Fatal("expected validation error")
		}

		validationErr := err.(*errors.Error)
		if validationErr.Details["status"] == nil {
			t.Error("expected status field in details")
		}
	})

	t.Run("multiple field errors", func(t *testing.T) {
		type User struct {
			Name  string `json:"name" validate:"required,min=3"`
			Email string `json:"email" validate:"required,email"`
			Age   int    `json:"age" validate:"required,gte=18"`
		}

		user := User{
			Name:  "Jo",
			Email: "invalid",
			Age:   15,
		}

		err := v.Validate(user)
		if err == nil {
			t.Fatal("expected validation error")
		}

		validationErr := err.(*errors.Error)

		if validationErr.Details["name"] == nil {
			t.Error("expected name field in details")
		}
		if validationErr.Details["email"] == nil {
			t.Error("expected email field in details")
		}
		if validationErr.Details["age"] == nil {
			t.Error("expected age field in details")
		}
	})
}

func TestValidateVar(t *testing.T) {
	v := New()

	t.Run("validates single variable", func(t *testing.T) {
		email := "test@example.com"
		err := v.ValidateVar(email, "email")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("fails on invalid variable", func(t *testing.T) {
		email := "invalid-email"
		err := v.ValidateVar(email, "email")
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}

func TestCurrencyValidator(t *testing.T) {
	v := New()

	t.Run("valid currency", func(t *testing.T) {
		type Transaction struct {
			Currency string `json:"currency" validate:"currency"`
		}

		validCurrencies := []string{"USD", "EUR", "GBP", "JPY"}

		for _, curr := range validCurrencies {
			tx := Transaction{Currency: curr}
			err := v.Validate(tx)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", curr, err)
			}
		}
	})

	t.Run("invalid currency", func(t *testing.T) {
		type Transaction struct {
			Currency string `json:"currency" validate:"currency"`
		}

		invalidCurrencies := []string{"XXX", "INVALID", "US", ""}

		for _, curr := range invalidCurrencies {
			tx := Transaction{Currency: curr}
			err := v.Validate(tx)
			if err == nil {
				t.Errorf("expected %s to be invalid", curr)
			}
		}
	})
}

func TestMoneyAmountValidator(t *testing.T) {
	v := New()

	t.Run("valid amount", func(t *testing.T) {
		type Transaction struct {
			Amount int64 `json:"amount" validate:"money_amount"`
		}

		validAmounts := []int64{1, 100, 1000, 999999999}

		for _, amount := range validAmounts {
			tx := Transaction{Amount: amount}
			err := v.Validate(tx)
			if err != nil {
				t.Errorf("expected %d to be valid, got error: %v", amount, err)
			}
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		type Transaction struct {
			Amount int64 `json:"amount" validate:"money_amount"`
		}

		invalidAmounts := []int64{0, -1, -100}

		for _, amount := range invalidAmounts {
			tx := Transaction{Amount: amount}
			err := v.Validate(tx)
			if err == nil {
				t.Errorf("expected %d to be invalid", amount)
			}
		}
	})
}

func TestAccountNumberValidator(t *testing.T) {
	v := New()

	t.Run("valid account number", func(t *testing.T) {
		type Account struct {
			Number string `json:"number" validate:"account_number"`
		}

		validNumbers := []string{
			"1234567890",
			"ABC1234567890",
			"1234567890ABCDEFGHIJ",
		}

		for _, num := range validNumbers {
			acc := Account{Number: num}
			err := v.Validate(acc)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", num, err)
			}
		}
	})

	t.Run("invalid account number", func(t *testing.T) {
		type Account struct {
			Number string `json:"number" validate:"account_number"`
		}

		invalidNumbers := []string{
			"123",                     // Too short
			"12345678901234567890123", // Too long
			"123456789a",              // Lowercase not allowed
			"12345-6789",              // Special characters
		}

		for _, num := range invalidNumbers {
			acc := Account{Number: num}
			err := v.Validate(acc)
			if err == nil {
				t.Errorf("expected %s to be invalid", num)
			}
		}
	})
}

func TestIBANValidator(t *testing.T) {
	v := New()

	t.Run("valid IBAN format", func(t *testing.T) {
		type BankAccount struct {
			IBAN string `json:"iban" validate:"iban"`
		}

		validIBANs := []string{
			"GB82WEST12345698765432",
			"DE89370400440532013000",
			"FR1420041010050500013M02606",
		}

		for _, iban := range validIBANs {
			acc := BankAccount{IBAN: iban}
			err := v.Validate(acc)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", iban, err)
			}
		}
	})

	t.Run("invalid IBAN format", func(t *testing.T) {
		type BankAccount struct {
			IBAN string `json:"iban" validate:"iban"`
		}

		invalidIBANs := []string{
			"GB",                     // Too short
			"1234567890123456789012", // Doesn't start with letters
			"GBAA1234567890",         // Invalid check digits (should be 2 digits)
			"GB821234567890123456789012345678901234567890", // Too long
		}

		for _, iban := range invalidIBANs {
			acc := BankAccount{IBAN: iban}
			err := v.Validate(acc)
			if err == nil {
				t.Errorf("expected %s to be invalid", iban)
			}
		}
	})
}

func TestSortCodeValidator(t *testing.T) {
	v := New()

	t.Run("valid sort code", func(t *testing.T) {
		type BankAccount struct {
			SortCode string `json:"sort_code" validate:"sort_code"`
		}

		validCodes := []string{
			"123456",
			"12-34-56",
		}

		for _, code := range validCodes {
			acc := BankAccount{SortCode: code}
			err := v.Validate(acc)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", code, err)
			}
		}
	})

	t.Run("invalid sort code", func(t *testing.T) {
		type BankAccount struct {
			SortCode string `json:"sort_code" validate:"sort_code"`
		}

		invalidCodes := []string{
			"12345",    // Too short
			"1234567",  // Too long
			"12-34-5A", // Contains letter
			"ABCDEF",   // All letters
		}

		for _, code := range invalidCodes {
			acc := BankAccount{SortCode: code}
			err := v.Validate(acc)
			if err == nil {
				t.Errorf("expected %s to be invalid", code)
			}
		}
	})
}

func TestRoutingNumberValidator(t *testing.T) {
	v := New()

	t.Run("valid routing number", func(t *testing.T) {
		type BankAccount struct {
			RoutingNumber string `json:"routing_number" validate:"routing_number"`
		}

		validNumbers := []string{
			"021000021",
			"111000025",
			"026009593",
		}

		for _, num := range validNumbers {
			acc := BankAccount{RoutingNumber: num}
			err := v.Validate(acc)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", num, err)
			}
		}
	})

	t.Run("invalid routing number", func(t *testing.T) {
		type BankAccount struct {
			RoutingNumber string `json:"routing_number" validate:"routing_number"`
		}

		invalidNumbers := []string{
			"12345678",   // Too short
			"1234567890", // Too long
			"02100002A",  // Contains letter
			"021-000-021", // Contains hyphens
		}

		for _, num := range invalidNumbers {
			acc := BankAccount{RoutingNumber: num}
			err := v.Validate(acc)
			if err == nil {
				t.Errorf("expected %s to be invalid", num)
			}
		}
	})
}

func TestGlobalValidator(t *testing.T) {
	t.Run("default validator is initialized", func(t *testing.T) {
		v := Default()
		if v == nil {
			t.Fatal("expected default validator to be initialized")
		}
	})

	t.Run("global Validate function", func(t *testing.T) {
		type User struct {
			Name string `json:"name" validate:"required"`
		}

		user := User{Name: "John"}
		err := Validate(user)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		user = User{}
		err = Validate(user)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("global ValidateVar function", func(t *testing.T) {
		email := "test@example.com"
		err := ValidateVar(email, "email")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		email = "invalid"
		err = ValidateVar(email, "email")
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}

func TestErrorMessages(t *testing.T) {
	v := New()

	testCases := []struct {
		name          string
		input         interface{}
		expectedField string
	}{
		{
			name: "required field",
			input: struct {
				Name string `json:"name" validate:"required"`
			}{},
			expectedField: "name",
		},
		{
			name: "email field",
			input: struct {
				Email string `json:"email" validate:"email"`
			}{Email: "invalid"},
			expectedField: "email",
		},
		{
			name: "min length",
			input: struct {
				Password string `json:"password" validate:"min=8"`
			}{Password: "short"},
			expectedField: "password",
		},
		{
			name: "numeric comparison",
			input: struct {
				Age int `json:"age" validate:"gte=18"`
			}{Age: 15},
			expectedField: "age",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(tc.input)
			if err == nil {
				t.Fatal("expected validation error")
			}

			validationErr := err.(*errors.Error)
			if validationErr.Details[tc.expectedField] == nil {
				t.Errorf("expected %s field in details", tc.expectedField)
			}

			message := validationErr.Details[tc.expectedField].(string)
			if message == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

func TestJSONTagNames(t *testing.T) {
	v := New()

	t.Run("uses json tag names in errors", func(t *testing.T) {
		type User struct {
			FullName string `json:"full_name" validate:"required"`
		}

		user := User{}
		err := v.Validate(user)
		if err == nil {
			t.Fatal("expected validation error")
		}

		validationErr := err.(*errors.Error)
		if validationErr.Details["full_name"] == nil {
			t.Error("expected 'full_name' (json tag) in details, not 'FullName' (struct field)")
		}
	})
}
