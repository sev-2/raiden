package raiden

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidHeaders(t *testing.T) {
	tests := []struct {
		name           string
		requestHeaders string
		allowedHeaders []string
		expected       bool
		description    string
	}{
		{
			name:           "Empty allowed headers list - should allow any headers",
			requestHeaders: "content-type, authorization",
			allowedHeaders: []string{},
			expected:       true,
			description:    "When no headers are restricted, all requests should pass",
		},
		{
			name:           "Empty request headers - should allow",
			requestHeaders: "",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Requests with no custom headers should be allowed",
		},
		{
			name:           "Whitespace only request headers - should allow",
			requestHeaders: "   ",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Requests with only whitespace headers should be allowed",
		},
		{
			name:           "Headers without spaces after comma",
			requestHeaders: "content-type,authorization",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Should support comma-only separated headers (no spaces)",
		},
		{
			name:           "Headers with spaces after comma",
			requestHeaders: "content-type, authorization",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Should support traditional comma-space separated headers",
		},
		{
			name:           "Mixed case headers - should normalize to lowercase",
			requestHeaders: "Content-Type,Authorization",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Header names should be case-insensitive",
		},
		{
			name:           "Headers with extra whitespace",
			requestHeaders: "  content-type  ,  authorization  ",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Should trim whitespace from header names",
		},
		{
			name:           "Headers with empty entries (trailing comma)",
			requestHeaders: "content-type,authorization,",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Should skip empty entries from trailing commas",
		},
		{
			name:           "Headers with empty entries (multiple commas)",
			requestHeaders: "content-type,,authorization",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Should skip empty entries from consecutive commas",
		},
		{
			name:           "Single valid header",
			requestHeaders: "content-type",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       true,
			description:    "Single header should be validated correctly",
		},
		{
			name:           "Invalid header - not in allowed list",
			requestHeaders: "x-custom-header",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       false,
			description:    "Headers not in allowed list should be rejected",
		},
		{
			name:           "Mixed valid and invalid headers",
			requestHeaders: "content-type,x-custom-header",
			allowedHeaders: []string{"content-type", "authorization"},
			expected:       false,
			description:    "Should reject if any header is not allowed",
		},
		{
			name:           "Case mismatch with mixed separators",
			requestHeaders: "Content-Type,Authorization, X-Requested-With",
			allowedHeaders: []string{"content-type", "authorization", "x-requested-with"},
			expected:       true,
			description:    "Should handle mixed case with both separator styles",
		},
		{
			name:           "Complex realistic scenario",
			requestHeaders: "content-type, authorization,x-requested-with",
			allowedHeaders: []string{"content-type", "authorization", "x-requested-with"},
			expected:       true,
			description:    "Should handle real-world mixed formatting",
		},
		{
			name:           "All empty entries",
			requestHeaders: ",,,",
			allowedHeaders: []string{"content-type"},
			expected:       true,
			description:    "Should handle all empty entries gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidHeaders(tt.requestHeaders, tt.allowedHeaders)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestIsValidHeadersEdgeCases(t *testing.T) {
	t.Run("Nil allowed headers slice", func(t *testing.T) {
		result := isValidHeaders("content-type", nil)
		assert.True(t, result, "Nil allowed headers should allow all requests")
	})

	t.Run("Very long header list", func(t *testing.T) {
		longHeaders := "header1,header2,header3,header4,header5,header6,header7,header8,header9,header10"
		allowedHeaders := []string{
			"header1", "header2", "header3", "header4", "header5",
			"header6", "header7", "header8", "header9", "header10",
		}
		result := isValidHeaders(longHeaders, allowedHeaders)
		assert.True(t, result, "Should handle long header lists")
	})

	t.Run("Headers with tabs and special whitespace", func(t *testing.T) {
		result := isValidHeaders("content-type\t,\tauthorization", []string{"content-type", "authorization"})
		assert.True(t, result, "Should handle tab characters")
	})

	t.Run("Single comma only", func(t *testing.T) {
		result := isValidHeaders(",", []string{"content-type"})
		assert.True(t, result, "Single comma should be treated as empty")
	})
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		arr      []string
		str      string
		expected bool
	}{
		{
			name:     "String exists in array",
			arr:      []string{"content-type", "authorization"},
			str:      "content-type",
			expected: true,
		},
		{
			name:     "String does not exist in array",
			arr:      []string{"content-type", "authorization"},
			str:      "x-custom-header",
			expected: false,
		},
		{
			name:     "Empty array",
			arr:      []string{},
			str:      "content-type",
			expected: false,
		},
		{
			name:     "Empty string search",
			arr:      []string{"content-type", "authorization"},
			str:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.arr, tt.str)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCanonicalHeaderKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase with no spaces",
			input:    "content-type",
			expected: "content-type",
		},
		{
			name:     "Mixed case with spaces",
			input:    "Content Type",
			expected: "content_type",
		},
		{
			name:     "Uppercase with spaces",
			input:    "CONTENT TYPE",
			expected: "content_type",
		},
		{
			name:     "Multiple spaces",
			input:    "X  Custom  Header",
			expected: "x__custom__header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCanonicalHeaderKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
