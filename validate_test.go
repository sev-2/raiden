package raiden_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/sev-2/raiden"
)

// Mock payload for testing
type TestPayload struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=0,lte=130"`
}

type TestAllPayload struct {
	Name                    string `validate:"required"`
	MinMax                  string `validate:"min=5,max=10"`
	EqualText               string `validate:"eq=custom"`
	EqualIgnoreCase         string `validate:"eq_ignore_case=custom"`
	NotEqualText            string `validate:"ne=custom"`
	NotEqualIgnoreCase      string `validate:"ne_ignore_case=custom"`
	AlphaText               string `validate:"alpha"`
	AlphaNumericText        string `validate:"alphanum"`
	NumericOnly             string `validate:"numeric"`
	BooleanOnly             string `validate:"boolean"`
	AlphaUnicodeText        string `validate:"alphaunicode"`
	AlphaNumericUnicodeText string `validate:"alphanumunicode"`
	AsciiText               string `validate:"ascii"`
	ContainsText            string `validate:"contains=custom"`
	ContainsAnyText         string `validate:"containsany=custom"`
	ContainsRuneText        string `validate:"containsrune=custom"`
	EndsNotWith             string `validate:"endsnotwith=custom"`
	EndsWith                string `validate:"endswith=custom"`
	ExcludeText             string `validate:"excludes=custom"`
	ExcludesAllText         string `validate:"excludesall=custom"`
	ExcludesRuneText        string `validate:"excludesrune=custom"`
	LowerCaseOnly           string `validate:"lowercase"`
	MultiByteText           string `validate:"multibyte"`
	FqdnOnly                string `validate:"fqdn"`
	HostnameOnly            string `validate:"hostname"`
	IPOnly                  string `validate:"ip"`
	IPv4Only                string `validate:"ipv4"`
	IPv6Only                string `validate:"ipv6"`
	MACOnly                 string `validate:"mac"`
	URIOnly                 string `validate:"uri"`
	URLOnly                 string `validate:"url"`
	SHA256Only              string `validate:"sha256"`
	DatetimeOnly            string `validate:"datetime"`
}

// Custom validation function
func customValidation(fl validator.FieldLevel) bool {
	return fl.Field().String() == "custom"
}

// Custom validation function name
const customValidationName = "custom_validation"

func TestValidate_Success(t *testing.T) {
	payload := TestPayload{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Age:   30,
	}

	err := raiden.Validate(context.Background(), payload)
	assert.NoError(t, err)
}

func TestValidate_Failure(t *testing.T) {
	payload := TestPayload{
		Name:  "",
		Email: "invalid-email",
		Age:   150,
	}

	err := raiden.Validate(context.Background(), payload)
	assert.Error(t, err)

	validationErr, ok := err.(*raiden.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, fasthttp.StatusBadRequest, validationErr.StatusCode)

	var details map[string][]string
	err1 := json.Unmarshal([]byte(validationErr.Details.(string)), &details)
	assert.NoError(t, err1)

	assert.Contains(t, details["Name"], "name is required")
	assert.Contains(t, details["Email"], "email should be a valid email address")
	assert.Contains(t, details["Age"], "age should be less than or equal 130")
}

func TestValidate_SuccessAll(t *testing.T) {
	payload := TestAllPayload{
		Name:                    "John Doe",
		MinMax:                  "123456",
		EqualText:               "custom",
		EqualIgnoreCase:         "Custom",
		NotEqualText:            "not-custom",
		NotEqualIgnoreCase:      "Not-Custom",
		AlphaText:               "abc",
		AlphaNumericText:        "abc123",
		NumericOnly:             "123",
		BooleanOnly:             "true",
		AlphaUnicodeText:        "abc",
		AlphaNumericUnicodeText: "abc123",
		AsciiText:               "abc",
		ContainsText:            "custom",
		ContainsAnyText:         "custom",
		ContainsRuneText:        "custom",
		EndsNotWith:             "missing",
		EndsWith:                "custom",
		ExcludeText:             "missing",
		ExcludesAllText:         "new",
		ExcludesRuneText:        "missing",
		LowerCaseOnly:           "abc",
		FqdnOnly:                "example.com",
		HostnameOnly:            "example.com",
		IPOnly:                  "192.0.0.1",
		IPv4Only:                "127.0.0.1",
		IPv6Only:                "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		MACOnly:                 "00:00:5e:00:53:01",
		URIOnly:                 "http://example.com",
		URLOnly:                 "http://example.com",
		SHA256Only:              "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
	}

	err := raiden.Validate(context.Background(), payload)
	assert.NoError(t, err)
}

func TestValidate_FailureAll(t *testing.T) {
	payload := TestAllPayload{
		Name:                    "",
		MinMax:                  "123",
		EqualText:               "not-custom",
		EqualIgnoreCase:         "Not-Custom",
		NotEqualText:            "custom",
		NotEqualIgnoreCase:      "Custom",
		AlphaText:               "abc123",
		AlphaNumericText:        "abc123!",
		NumericOnly:             "abc",
		BooleanOnly:             "yes",
		AlphaUnicodeText:        "abc123",
		AlphaNumericUnicodeText: "abc123!",
		AsciiText:               "abc123",
		ContainsText:            "missing",
		ContainsAnyText:         "missing",
		ContainsRuneText:        "missing",
		EndsNotWith:             "custom",
		EndsWith:                "missing",
		ExcludeText:             "custom",
		ExcludesAllText:         "old",
		ExcludesRuneText:        "custom",
		LowerCaseOnly:           "ABC",
		MultiByteText:           "abc123",
		FqdnOnly:                "example",
		HostnameOnly:            "example",
		IPOnly:                  "not-an-ip",
		IPv4Only:                "not-an-ipv4",
		IPv6Only:                "not-an-ipv6",
		MACOnly:                 "not-a-mac",
		URIOnly:                 "example.com",
		URLOnly:                 "example.com",
		SHA256Only:              "not-a-sha256",
		DatetimeOnly:            "not-a-datetime",
	}

	err := raiden.Validate(context.Background(), payload)
	assert.Error(t, err)

	validationErr, ok := err.(*raiden.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, fasthttp.StatusBadRequest, validationErr.StatusCode)

	var details map[string][]string
	err1 := json.Unmarshal([]byte(validationErr.Details.(string)), &details)
	assert.NoError(t, err1)

	assert.Contains(t, details["Name"], "name is required")
	assert.Contains(t, details["MinMax"], "minmax should be at least 5")
	assert.Contains(t, details["AlphaNumericText"], "alphanumerictext should contain only alphanumeric characters")
	assert.Contains(t, details["NumericOnly"], "numericonly should be a numeric value")
	assert.Contains(t, details["BooleanOnly"], "booleanonly should be a boolean value (true or false)")
	assert.Contains(t, details["AlphaText"], "alphatext should contain only alphabetical characters")
	assert.Contains(t, details["AlphaUnicodeText"], "alphaunicodetext should contain only alphabetical Unicode characters")
	assert.Contains(t, details["ContainsRuneText"], "containsrunetext should contain the specified Unicode rune")
	assert.Contains(t, details["ContainsText"], "containstext should contain the specified substring")
	assert.Contains(t, details["EndsNotWith"], "endsnotwith should not end with the specified suffix")
	assert.Contains(t, details["EndsWith"], "endswith should end with the specified suffix")
	assert.Contains(t, details["EqualIgnoreCase"], "equalignorecase should be equal to the specified value (case-insensitive)")
	assert.Contains(t, details["EqualText"], "equaltext should be equal to custom")
	assert.Contains(t, details["ExcludeText"], "excludetext should not contain the specified substring")
	assert.Contains(t, details["ExcludesAllText"], "excludesalltext should not contain any of the specified substrings")
	assert.Contains(t, details["ExcludesRuneText"], "excludesrunetext should not contain the specified Unicode rune")
	assert.Contains(t, details["LowerCaseOnly"], "lowercaseonly should be in lowercase")
	assert.Contains(t, details["MultiByteText"], "multibytetext should contain only multi-byte characters")
	assert.Contains(t, details["FqdnOnly"], "fqdnonly should be a Full Qualified Domain Name (FQDN)")
	assert.Contains(t, details["IPOnly"], "iponly should be an Internet Protocol Address (IP)")
	assert.Contains(t, details["IPv4Only"], "ipv4only should be an Internet Protocol Address (IPv4)")
	assert.Contains(t, details["IPv6Only"], "ipv6only should be an Internet Protocol Address (IPv6)")
	assert.Contains(t, details["MACOnly"], "maconly should be a Media Access Control Address (MAC)")
	assert.Contains(t, details["URIOnly"], "urionly should be a URI String")
	assert.Contains(t, details["URLOnly"], "urlonly should be a URL String")
	assert.Contains(t, details["SHA256Only"], "sha256only should be a SHA256 hash")
	assert.Contains(t, details["DatetimeOnly"], "datetimeonly should be a Datetime")
}

func TestValidate_CustomValidator(t *testing.T) {
	payload := TestPayload{
		Name:  "custom",
		Email: "john.doe@example.com",
		Age:   30,
	}

	err := raiden.Validate(context.Background(), payload, raiden.ValidatorFunc{Name: customValidationName, Validator: customValidation})
	assert.NoError(t, err)
}

// TestRequiredForMethodValidator tests validation logic based on HTTP methods
func TestRequiredForMethodValidator(t *testing.T) {

	// Helper function to create a validation context with HTTP method
	var createValidationContext = func(method string) context.Context {
		return context.WithValue(context.Background(), raiden.MethodContextKey, method)
	}

	validate := validator.New()
	err := validate.RegisterValidationCtx("requiredForMethod", raiden.RequiredForMethodValidator)
	assert.NoError(t, err)

	one := 1
	uintOne := uint(1)
	floatOne := float32(1)
	zero := 0
	str := "hello"
	emptyStr := ""
	slice := []int{1, 2, 3}
	emptySlice := []int{}
	m := map[string]int{"key": 1}
	emptyMap := map[string]int{}
	customStruct := struct{ Name string }{"Test"}
	emptyStruct := struct{ Name string }{}

	tests := []struct {
		name     string
		method   string
		data     interface{}
		expected bool
	}{
		// Integers
		{"Valid int for POST", fasthttp.MethodPost, struct {
			ID int `validate:"requiredForMethod=POST"`
		}{ID: one}, true},
		{"Valid pointer int for POST", fasthttp.MethodPost, struct {
			ID *int `validate:"requiredForMethod=POST"`
		}{ID: &one}, true},
		{"Valid uint for POST", fasthttp.MethodPost, struct {
			ID uint `validate:"requiredForMethod=POST"`
		}{ID: uintOne}, true},
		{"Valid float for POST", fasthttp.MethodPost, struct {
			ID float32 `validate:"requiredForMethod=POST"`
		}{ID: floatOne}, true},
		{"Invalid int=0 for POST", fasthttp.MethodPost, struct {
			ID int `validate:"requiredForMethod=POST"`
		}{ID: zero}, false},
		{"Valid *int for GET", fasthttp.MethodGet, struct {
			ID *int `validate:"requiredForMethod=GET"`
		}{ID: &one}, true},
		{"Invalid nil *int for GET", fasthttp.MethodGet, struct {
			ID *int `validate:"requiredForMethod=GET"`
		}{ID: nil}, false},

		// Strings
		{"Valid string for POST", fasthttp.MethodPost, struct {
			Name string `validate:"requiredForMethod=POST"`
		}{Name: str}, true},
		{"Invalid empty string for POST", fasthttp.MethodPost, struct {
			Name string `validate:"requiredForMethod=POST"`
		}{Name: emptyStr}, false},

		// Slices
		{"Valid slice for PUT", fasthttp.MethodPut, struct {
			Tags []int `validate:"requiredForMethod=PUT"`
		}{Tags: slice}, true},
		{"Invalid empty slice for PUT", fasthttp.MethodPut, struct {
			Tags []int `validate:"requiredForMethod=PUT"`
		}{Tags: emptySlice}, false},

		// Maps
		{"Valid map for PATCH", fasthttp.MethodPatch, struct {
			Config map[string]int `validate:"requiredForMethod=PATCH"`
		}{Config: m}, true},
		{"Invalid empty map for PATCH", fasthttp.MethodPatch, struct {
			Config map[string]int `validate:"requiredForMethod=PATCH"`
		}{Config: emptyMap}, false},

		// Structs
		{"Valid struct for DELETE", fasthttp.MethodDelete, struct {
			Info struct{ Name string } `validate:"requiredForMethod=DELETE"`
		}{Info: customStruct}, true},
		{"Invalid empty struct for DELETE", fasthttp.MethodDelete, struct {
			Info struct{ Name string } `validate:"requiredForMethod=DELETE"`
		}{Info: emptyStruct}, false},

		// Optional cases (not required)
		{"Valid missing field for GET (not required)", fasthttp.MethodGet, struct {
			Unused string `validate:"requiredForMethod=POST"`
		}{Unused: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createValidationContext(tt.method)
			err := validate.StructCtx(ctx, tt.data)
			if tt.expected {
				assert.NoError(t, err, "Validation should pass")
			} else {
				assert.Error(t, err, "Validation should fail")
			}
		})
	}

}
